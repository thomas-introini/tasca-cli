package spinnerlabel

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	styles "github.com/thomas-introini/pocket-cli/views"
)

type Model struct {
	fallbackLabel      string
	fallbackLabelStyle lipgloss.Style
	showSpinner        bool
	label              string
	labelStyle         lipgloss.Style
	spinner            spinner.Model
}

func New(label, fallbackLabel string) Model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = styles.TitleRedStyle
	return Model{
		fallbackLabel:      fallbackLabel,
		fallbackLabelStyle: styles.TitleBoldRedStyle,
		showSpinner:        false,
		label:              label,
		spinner:            s,
		labelStyle:         styles.TitleRedStyle,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.showSpinner {
		return m.spinner.View() + " " + m.labelStyle.Render(m.label)
	} else {
		return m.fallbackLabelStyle.Render(m.fallbackLabel)
	}
}

func (m *Model) SetLabel(label string) {
	m.label = label
}

func (m *Model) SetShow(show bool) {
	m.showSpinner = show
}
