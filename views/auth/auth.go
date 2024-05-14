package auth

import (

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thomas-introini/pocket-cli/models"
	styles "github.com/thomas-introini/pocket-cli/views"
)

type Model struct {
	label   string
	spinner spinner.Model
	User    models.PocketUser
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
	case ChangeLabel:
		m.label = msg.Label
		return m, nil
	default:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m Model) View() string {
	return m.spinner.View() + " " + styles.TitleRedStyle.Render(m.label)
}

type ChangeLabel struct {
	Label string
}
