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

	filesStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FAFAFA")).
			Border(filesBorder).
			BorderForeground(highlightColor).
			Width(leftColumnWidth)

	historyStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FAFAFA")).
			Border(historyBorder).
			Width(leftColumnWidth)
)
