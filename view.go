package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	leftColumnWidth = 25
	borderWidth     = 1
	highlightColor  = lipgloss.AdaptiveColor{Dark: "#7CFC00", Light: "#7CFC00"}

	filesBorder = lipgloss.Border{
		Top:         "Files────────────────────",
		TopLeft:     "┌",
		TopRight:    "┐",
		Bottom:      "─",
		Right:       "│",
		Left:        "│",
		BottomRight: "┘",
		BottomLeft:  "└",
	}

	historyBorder = lipgloss.Border{
		Top:         "History────────────────────",
		TopLeft:     "┌",
		TopRight:    "┐",
		Bottom:      "─",
		Right:       "│",
		Left:        "│",
		BottomRight: "┘",
		BottomLeft:  "└",
	}

	baseFilesStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FAFAFA")).
			Border(filesBorder).
			Width(leftColumnWidth)

	baseHistoryStyle = lipgloss.NewStyle().
				Align(lipgloss.Left).
				Foreground(lipgloss.Color("#FAFAFA")).
				Border(historyBorder).
				Width(leftColumnWidth)

	baseChatStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FAFAFA")).
			Border(lipgloss.NormalBorder())

	baseInputStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FAFAFA")).
			Border(lipgloss.NormalBorder())
)

func RenderFiles(isFocused bool) string {
	filesStyle := baseFilesStyle.Copy()
	if isFocused {
		filesStyle = filesStyle.BorderForeground(highlightColor)
	}

	return filesStyle.Render("Files")
}

func RenderHistory(isFocused bool) string {
	historyStyle := baseHistoryStyle.Copy()

	if isFocused {
		historyStyle = historyStyle.BorderForeground(highlightColor)
	}

	return historyStyle.Render("History")
}

func RenderChat(isFocused bool,
	terminalWidth int,
	terminalHeight int,
	messages []Message,
) string {
	chatStyle := baseChatStyle.Copy().
		Width(terminalWidth - leftColumnWidth - (4 * borderWidth)).
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
