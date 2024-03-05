package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	openai "github.com/sashabaranov/go-openai"
)

type Message struct {
	Content string
	Role    string
}

type model struct {
	messages    []Message
	textInput   textinput.Model
	currentChan chan string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Pikachu"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		textInput: ti,
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

		fmt.Printf("Stream response: ")
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				fmt.Println("\nStream finished")
				return nil
			}

			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				return nil
			}

			m <- response.Choices[0].Delta.Content
		}
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Grocery List")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.messages = append(m.messages, Message{
			Content: string(msg),
			Role:    "assistant",
		})
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

func (m model) View() string {
	return fmt.Sprintf(
		"Messages:\n\n%s\n\n%s\n\n%s",
		DrawMessages(m.messages),
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
