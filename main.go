package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/sashabaranov/go-openai"
)

type Message struct {
	Content string
	Role    string
}

type Model struct {
	page        string
	focusedPane string

	currentStream chan string
	cancelSignal  chan struct{}
	messages      []Message
	textInput     textinput.Model
}

func initialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter a message..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	return Model{
		textInput:    ti,
		messages:     []Message{},
		page:         "main",
		cancelSignal: make(chan struct{}),
	}
}

type responseMsg string

func waitForActivity(sub chan string) tea.Cmd {
	return func() tea.Msg {
		select {
		case response := <-sub:
			return responseMsg(response)
		}
	}
}

func CallOpenAI(currentStream chan string,
	cancelSignal chan struct{},
	messages []Message,
) tea.Cmd {
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
			Model:     openai.GPT4o,
			MaxTokens: 1000,
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
			select {
			case <-cancelSignal:
				return nil
			default:
				response, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					return nil
				}

				if err != nil {
					return nil
				}

				currentStream <- response.Choices[0].Delta.Content
			}
		}
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func SendCancelSignal(cancelSignal chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return nil
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlD:
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.cancelSignal <- struct{}{}
			return m, cmd
		case tea.KeyEnter:
			m.messages = append(m.messages, Message{
				Content: m.textInput.Value(),
				Role:    "user",
			})
			m.textInput.SetValue("")
			m.currentStream = make(chan string)
			return m, tea.Batch(waitForActivity(m.currentStream),
				CallOpenAI(m.currentStream, m.cancelSignal, m.messages))
		}

		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	case responseMsg:
		if m.messages[len(m.messages)-1].Role != "assistant" {
			m.messages = append(m.messages, Message{
				Content: string(msg),
				Role:    "assistant",
			})
		} else {
			m.messages[len(m.messages)-1].Content += string(msg)
		}
		return m, waitForActivity(m.currentStream)
	}
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
	switch m.page {
	case "main":
		doc := strings.Builder{}
		width, height, _ := term.GetSize(int(os.Stdout.Fd()))

		docStyle := lipgloss.NewStyle().
			Width(width).Height(height)

		doc.WriteString(RenderChat(m.focusedPane == "chat", width, height, m.messages))
		doc.WriteString("\n")
		doc.WriteString(RenderInput(m.focusedPane == "input", width, m.textInput))
		return fmt.Sprint(docStyle.Render(doc.String()))
	}
	return ""
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
