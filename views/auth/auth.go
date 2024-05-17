package auth

import (

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	styles "github.com/thomas-introini/pocket-cli/views"
)

type Model struct {
	label   string
	spinner spinner.Model
}

func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = styles.TitleRedStyle
	return Model{
		label:   "Authentication in progress...\n",
		spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	default:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m Model) View() string {
	return m.spinner.View() + " " + styles.TitleRedStyle.Render(m.label)
}

func (m *Model) SetLabel(label string) {
	m.label = label
}
