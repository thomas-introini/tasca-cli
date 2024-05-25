package saves

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomas-introini/pocket-cli/models"
	"github.com/thomas-introini/pocket-cli/utils"
	styles "github.com/thomas-introini/pocket-cli/views"
)

var (
	titleStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4056")).Bold(true)
	itemStyle              = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemTitleStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("#ef4056")).
				Foreground(lipgloss.Color("#ef4056")).
				Bold(true).
				Padding(0, 0, 0, 1)
	selectedItemDescriptionStyle = lipgloss.NewStyle().
					Border(lipgloss.NormalBorder(), false, false, false, true).
					BorderForeground(lipgloss.Color("#ef4056")).
					Foreground(lipgloss.Color("#ef4056")).
					Padding(0, 0, 0, 1)
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle   = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type RefreshSavesCmd struct {
}

type ViewSaveCmd struct {
	Save models.PocketSave
}

type window struct {
	width  int
	height int
}

type Model struct {
	user         models.PocketUser
	window       window
	list         list.Model
	since        int32
	loading      bool
	spinner      spinner.Model
	errorMessage string
}

type UpdateSaves struct {
	Saves []models.PocketSave
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "R":
				cmd = func() tea.Msg {
					return RefreshSavesCmd{}
				}
			case "o":
				selected, ok := m.list.SelectedItem().(models.PocketSave)
				if ok {
					cmd = open(selected.Url)
				}
			case "enter":
				selected, ok := m.list.SelectedItem().(models.PocketSave)
				if ok {
					return m, func() tea.Msg {
						return ViewSaveCmd{Save: selected}
					}
				}
			case "esc":
				return m, nil
			}
		}
	case openError:
		m.errorMessage = msg.error.Error()
	case tea.WindowSizeMsg:
		m.window.width = msg.Width
		m.window.height = msg.Height - 3

		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 3)
	}
	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	return m, tea.Batch(cmd, spinnerCmd, listCmd)
}

func (m Model) View() string {
	if m.errorMessage != "" {
		tmp := styles.TitleRedStyle.Render("! ERROR" + m.errorMessage + " !")
		view := strings.Repeat(" ", (m.window.width-lipgloss.Width(tmp))/2) + tmp
		return view
	} else if len(m.list.Items()) == 0 {
		tmp := m.spinner.View() + " " + styles.TitleRedStyle.Render("Fetching your saved items...")
		view := strings.Repeat(" ", (m.window.width-lipgloss.Width(tmp))/2) + tmp
		return view
	} else {
		return m.list.View()
	}
}

func (m *Model) SetSaves(saves []models.PocketSave) {
	items := make([]list.Item, 0)
	for _, s := range saves {
		items = append(items, s)
	}
	m.list.SetItems(items)
}

func New(user models.PocketUser) Model {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = styles.TitleRedStyle

	id := list.NewDefaultDelegate()
	id.Styles.SelectedTitle = selectedItemTitleStyle
	id.Styles.SelectedDesc = selectedItemDescriptionStyle

	list := list.New(make([]list.Item, 0), id, 10, 10)
	list.Title = "Saves"
	list.SetShowTitle(false)
	list.Styles.Title = titleStyle
	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("Enter", "<cr>"),
				key.WithHelp("Enter", "View"),
			),
			key.NewBinding(
				key.WithKeys("o"),
				key.WithHelp("o", "Open"),
			),
			key.NewBinding(
				key.WithKeys("R"),
				key.WithHelp("R", "Refresh saves"),
			),
		}
	}

	return Model{
		list:         list,
		loading:      true,
		spinner:      s,
		user:         user,
		errorMessage: "",
	}
}

type openError struct {
	error error
}

func open(url string) tea.Cmd {
	return func() tea.Msg {
		err := utils.OpenInBrowser(url)
		if err != nil {
			return openError{err}
		}
		return nil
	}
}
