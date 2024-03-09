package main

import (
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
