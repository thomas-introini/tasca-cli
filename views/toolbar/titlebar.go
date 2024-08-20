package titlebar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	styles "github.com/thomas-introini/pocket-cli/views"
	"github.com/thomas-introini/pocket-cli/views/spinnerlabel"
)

type window struct {
	width  int
	height int
}

type Model struct {
	user    string
	window  window
	message spinnerlabel.Model
}

func New(user, title string) Model {
	return Model{
		user:    user,
		window:  window{},
		message: spinnerlabel.New("", title),
	}
}

func (m Model) Init() tea.Cmd {
	return m.message.Init()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.window.width, m.window.height = msg.Width, msg.Height
	}
	m.message, cmd = m.message.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var msg = m.message.View()
	toolbarMaxWidth := m.window.width - 5
	toolbarUser := lipgloss.NewStyle().MarginRight(1).Render(m.user)
	toolbarMessage := lipgloss.NewStyle().MarginLeft(1).Width(toolbarMaxWidth - 1 - lipgloss.Width(toolbarUser)).Render(msg)
	return styles.ToolbarMessage.Width(toolbarMaxWidth).Render(toolbarMessage+toolbarUser) + "\n"
}

func (m *Model) ShowMessage(msg string) {
	m.message.SetLabel(msg)
	m.message.SetShow(true)
}

func (m *Model) ClearMessage() {
	m.message.SetShow(false)
	m.message.SetLabel("")
}
