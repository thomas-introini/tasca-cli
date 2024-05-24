package itemdetail

import (
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thomas-introini/pocket-cli/models"
	styles "github.com/thomas-introini/pocket-cli/views"
)

type Model struct {
	width    int
	height   int
	item     models.PocketSave
	viewport viewport.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	}
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return m.viewport.View()
}

func (m *Model) SetItem(item models.PocketSave) {
	m.item = item
	m.viewport = viewport.New(m.width -4, m.height -20)
	m.viewport.Style = m.viewport.Style.MarginLeft(3)
	m.viewport.YOffset = 20
	if m.IsItemSet() {
		m.viewport.SetContent(getContent(m.item))
	} else {
		m.viewport.SetContent("")
	}
}

func (m Model) IsItemSet() bool {
	return m.item != models.PocketSave{}
}

func getContent(item models.PocketSave) string {
	addedOn := time.Unix(int64(item.UpdatedOn), 0)
	content := ""
	content += styles.TitleBoldRedStyle.Render("Title:") + " " + item.SaveTitle + "\n"
	content += styles.TitleBoldRedStyle.Render("URL:") + " " + item.Url + "\n"
	content += styles.TitleBoldRedStyle.Render("Added on:") + " " + addedOn.Format("Mon Jan 2 2006 15:04") + "\n"
	content += "\n"
	content += item.SaveDescription
	return content
}
