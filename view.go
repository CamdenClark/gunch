package main

import (

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	borderWidth     = 1
	highlightColor  = lipgloss.AdaptiveColor{Dark: "#7CFC00", Light: "#7CFC00"}


	baseChatStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FAFAFA")).
			Border(lipgloss.NormalBorder())

	baseInputStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FAFAFA")).
			Border(lipgloss.NormalBorder())
)


func RenderChat(isFocused bool,
	terminalWidth int,
	terminalHeight int,
	messages []Message,
) string {
	chatStyle := baseChatStyle.Copy().
		Width(terminalWidth - (4 * borderWidth)).
		Height(terminalHeight - 5)
	if isFocused {
		chatStyle = chatStyle.BorderForeground(highlightColor)
	}
	return chatStyle.Render(DrawMessages(messages))
}

func RenderInput(isFocused bool, terminalWidth int, input textinput.Model) string {
	inputStyle := baseInputStyle.Copy().Width(terminalWidth - (2 * borderWidth))
	if isFocused {
		inputStyle = inputStyle.BorderForeground(highlightColor)
	}
	return inputStyle.Render(input.View())
}
