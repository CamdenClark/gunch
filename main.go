package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
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
	page        string
	focusedPane string

	messages    []Message
	textInput   textinput.Model
	files       []string
	filepicker  filepicker.Model
	currentChan chan string
}

func initialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter a message..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	filepicker := filepicker.New()
	// current directory

	return Model{
		textInput:   ti,
		messages:    []Message{},
		focusedPane: "input",
		page:        "main",
		files:       []string{},
		filepicker:  filepicker,
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
	return m.filepicker.Init()
}

func (m Model) switchPane(pane string) (Model, tea.Cmd) {
	m.focusedPane = pane
	return m, nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlD {
			return m, tea.Quit
		}
	}

	m.filepicker, cmd = m.filepicker.Update(msg)
	if m.page == "file-selector" {
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.files = append(m.files, path)
			m.page = "main"
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyRunes && m.focusedPane != "input" {
			switch string(msg.Runes) {
			case "1":
				return m.switchPane("files")
			case "2":
				return m.switchPane("history")
			case "3":
				return m.switchPane("input")
			case "4":
				return m.switchPane("chat")
			}
		}

		switch m.focusedPane {
		case "chat":
			switch msg.Type {
			case tea.KeyRunes:
				switch string(msg.Runes) {
				}
			}
		case "files":
			switch msg.Type {
			case tea.KeyRunes:
				switch string(msg.Runes) {
				case "a":
					m.page = "file-selector"
					m.filepicker.Init()

					return m, cmd
				}
			}
			return m, cmd
		case "history":
			return m, cmd
		case "input":
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
			case tea.KeyEsc:
				return m.switchPane("chat")
			}

			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
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
	case "file-selector":
		return m.filepicker.View()

	case "main":
		doc := strings.Builder{}
		width, height, _ := term.GetSize(int(os.Stdout.Fd()))

		docStyle := lipgloss.NewStyle().
			Width(width).Height(height)

		leftColumn := lipgloss.JoinVertical(0,
			RenderFiles(m.focusedPane == "files", m.files),
			RenderHistory(m.focusedPane == "history"))

		doc.WriteString(lipgloss.JoinHorizontal(0, leftColumn, RenderChat(m.focusedPane == "chat", width, height, m.messages)))
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
