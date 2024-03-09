package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/sashabaranov/go-openai"

	"golang.org/x/term"
)

type Message struct {
	Content string
	Role    string
}

type Model struct {
	focusedPane string
	messages    []Message
	textInput   textinput.Model
	files       []string
	currentChan chan string
}

func initialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter a message..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	return Model{
		textInput:   ti,
		messages:    []Message{},
		focusedPane: "input",
		files:       []string{},
	}
}

type responseMsg string

func waitForActivity(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}

func CallOpenAI(m chan string, messages []Message) tea.Cmd {
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i := range messages {
		message := messages[i]
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		}
	}

	return func() tea.Msg {
		c := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
		ctx := context.Background()

		req := openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: 200,
			Messages:  openaiMessages,
			Stream:    true,
		}
		stream, err := c.CreateChatCompletionStream(ctx, req)
		if err != nil {
			fmt.Printf("ChatCompletionStream error: %v\n", err)
			return nil
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}

			if err != nil {
				return nil
			}

			m <- response.Choices[0].Delta.Content
		}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("Grocery List")
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.messages = append(m.messages, Message{
				Content: m.textInput.Value(),
				Role:    "user",
			})
			m.textInput.SetValue("")
			m.currentChan = make(chan string)
			return m, tea.Batch(waitForActivity(m.currentChan),
				CallOpenAI(m.currentChan, m.messages))

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case responseMsg:
		if m.messages[len(m.messages)-1].Role != "assistant" {
			m.messages = append(m.messages, Message{
				Content: string(msg),
				Role:    "assistant",
			})
		} else {
			m.messages[len(m.messages)-1].Content += string(msg)
		}
		return m, waitForActivity(m.currentChan)
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func DrawMessages(messages []Message) string {
	var s string
	for _, m := range messages {
		s += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}
	return s
}

func (m Model) View() string {
	doc := strings.Builder{}
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))

	docStyle := lipgloss.NewStyle().
		Width(width).Height(height)

	chatStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(lipgloss.Color("#FAFAFA")).
		Border(lipgloss.NormalBorder()).
		Width(width - leftColumnWidth - (4 * borderWidth)).
		Height(height - 5)

	inputStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(lipgloss.Color("#FAFAFA")).
		Border(lipgloss.NormalBorder()).
		Width(width - (2 * borderWidth))

	leftColumn := lipgloss.JoinVertical(0,
		RenderFiles(m.focusedPane == "files"),
		RenderHistory(m.focusedPane == "history"))

	doc.WriteString(lipgloss.JoinHorizontal(0, leftColumn, chatStyle.Render(DrawMessages(m.messages))))
	doc.WriteString("\n")
	doc.WriteString(inputStyle.Render(m.textInput.View()))
	return fmt.Sprint(docStyle.Render(doc.String()))
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
