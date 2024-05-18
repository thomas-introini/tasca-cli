package views

import "github.com/charmbracelet/lipgloss"

var (
	ToolbarMessage = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
			Border(lipgloss.DoubleBorder(), true).
			Margin(0, 2, 0, 2).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})

	TitleRedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4056"))
	TitleBoldRedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ef4056"))
)
