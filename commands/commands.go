package commands

import tea "github.com/charmbracelet/bubbletea"

type SetLabelMsg struct {
	Show    bool
	Message string
}

func SetLabelCmd(msg string) tea.Cmd {
	return func() tea.Msg { return SetLabelMsg{msg != "", msg} }
}
