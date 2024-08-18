package root

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomas-introini/pocket-cli/commands"
	"github.com/thomas-introini/pocket-cli/db"
	"github.com/thomas-introini/pocket-cli/globals"
	"github.com/thomas-introini/pocket-cli/helpkeys"
	"github.com/thomas-introini/pocket-cli/lib"
	"github.com/thomas-introini/pocket-cli/models"
	styles "github.com/thomas-introini/pocket-cli/views"
	"github.com/thomas-introini/pocket-cli/views/auth"
	"github.com/thomas-introini/pocket-cli/views/itemdetail"
	"github.com/thomas-introini/pocket-cli/views/saves"
	"github.com/thomas-introini/pocket-cli/views/spinnerlabel"
)

type getSavesResult struct {
	saves []models.PocketSave
	count int
	err   error
}

type authResult struct {
	authFailure string
	openBrowser bool
	accessToken string
	username    string
}

type keyMap struct {
	Quit  key.Binding
	Enter key.Binding
}

func (m keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.Quit},
		{m.Enter},
	}
}

func (m keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		m.Quit,
		m.Enter,
	}
}

type window struct {
	width  int
	height int
}

type model struct {
	window         window
	user           models.PocketUser
	authenticating bool
	auth           auth.Model
	saves          saves.Model
	help           help.Model
	itemdetail     itemdetail.Model
	errorMessage   string
	message        spinnerlabel.Model
	keys           keyMap
}

func (m model) IsAuthenticated() bool {
	return m.user != models.NoUser
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("Pocket CLI"),
		tea.EnterAltScreen,
		m.auth.Init(),
		m.saves.Init(),
		m.message.Init(),
		loadSaves(m),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.window.width, m.window.height = msg.Width, msg.Height
		m.help.Width = m.window.width - 5
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if !m.IsAuthenticated() {
				m.authenticating = true
				cmds = append(cmds, startAuthentication())
			}
		case "esc":
			m.itemdetail.SetItem(models.PocketSave{})
		}
	case commands.SetLabelMsg:
		m.message.SetShow(msg.Show)
		m.message.SetLabel(msg.Message)
	case saves.RefreshSavesCmd:
		cmds = append(cmds, refreshSaves(m))
		m.message.SetShow(true)
		m.message.SetLabel("Refreshing saves...")
	case saves.ViewSaveCmd:
		if msg.Open || m.itemdetail.IsItemSet() {
			save := msg.Save
			m.itemdetail.SetItem(save)
		}
	case authResult:
		if !m.authenticating {
			return m, nil
		}
		if msg.authFailure != "" {
			m.auth.SetLabel(msg.authFailure + "\n")
		} else if msg.openBrowser {
			m.auth.SetLabel("Continue authentication in browser...\n")
		} else if msg.accessToken != "" {
			closeServer <- true
			m.authenticating = false
			m.user = models.PocketUser{
				AccessToken: msg.accessToken,
				Username:    msg.username,
			}
			_, err := db.SaveUser(msg.accessToken, msg.username)
			if err != nil {
				m.auth.SetLabel("Could not save user...\n")
				m.authenticating = false
			}
			cmds = append(cmds, loadSaves(m))
		}
	case getSavesResult:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
		} else {
			m.saves.SetSaves(msg.saves)
		}
		m.message.SetShow(false)
	}

	m.saves, cmd = m.saves.Update(msg)
	cmds = append(cmds, cmd)
	m.auth, cmd = m.auth.Update(msg)
	cmds = append(cmds, cmd)
	m.message, cmd = m.message.Update(msg)
	cmds = append(cmds, cmd)
	m.itemdetail, cmd = m.itemdetail.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.window.width == 0 {
		return "\n"
	}
	view := ""
	helpView := m.help.View(m.keys)
	if m.errorMessage != "" {
		view += strings.Repeat("\n", (m.window.height/2)-strings.Count(view, "\n")-2)
		tmp := styles.TitleRedStyle.Render("! ERROR: " + m.errorMessage + " !\n")
		view += strings.Repeat(" ", (m.window.width-lipgloss.Width(tmp))/2) + tmp
	} else if !m.IsAuthenticated() && !m.authenticating {
		view += strings.Repeat("\n", (m.window.height/2)-strings.Count(view, "\n")-2)
		tmp := styles.TitleRedStyle.Render("Welcome to Pocket CLI!") + "\n"
		view += strings.Repeat(" ", (m.window.width/2)-(lipgloss.Width(tmp)/2)) + tmp
		tmp = styles.TitleRedStyle.Render("Press '") + styles.TitleBoldRedStyle.Render("Enter") + styles.TitleRedStyle.Render("' to start the authentication") + "\n"
		view += strings.Repeat(" ", (m.window.width/2)-(lipgloss.Width(tmp)/2)) + tmp
	} else if m.authenticating {
		view += strings.Repeat("\n", (m.window.height/2)-strings.Count(view, "\n")-1)
		tmp := m.auth.View()
		view += strings.Repeat(" ", (m.window.width-lipgloss.Width(tmp))/2) + tmp
	} else {
		var msg string
		msg = m.message.View()
		toolbarMaxWidth := m.window.width - 5
		toolbarUser := lipgloss.NewStyle().MarginRight(1).Render(m.user.Username)
		toolbarMessage := lipgloss.NewStyle().MarginLeft(1).Width(toolbarMaxWidth - 1 - lipgloss.Width(toolbarUser)).Render(msg)
		view += styles.ToolbarMessage.Width(toolbarMaxWidth).Render(toolbarMessage+toolbarUser) + "\n"
		if m.itemdetail.IsItemSet() {
			view += m.itemdetail.View()
			helpView = m.help.View(getItemDetailKeys(m.itemdetail.GetItem()))
		} else {
			view += m.saves.View()
			helpView = ""
		}
	}
	height := strings.Count(view, "\n") + strings.Count(helpView, "\n")
	remainingHeight := math.Max(float64(m.window.height-height-1), 0)
	return view + strings.Repeat("\n", int(remainingHeight)) + helpView
}

func New(user models.PocketUser) model {
	return model{
		window:         window{},
		authenticating: false,
		user:           user,
		auth:           auth.New(),
		saves:          saves.New(user),
		help:           help.New(),
		message:        spinnerlabel.New("", "Tasca"),
		itemdetail:     itemdetail.New(),
		keys: keyMap{
			Quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
			Enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "start atuthentication"),
			),
		},
	}
}

func getItemDetailKeys(save models.PocketSave) help.KeyMap {
	keys := helpkeys.ItemdetailsKeys{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open"),
		),
		GetContent: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "get article content"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete"),
		),
	}
	if save.Status == models.StatusOK {
		keys.Archive = key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "archive"),
		)
	} else {
		keys.Unarchive = key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "move to saves"),
		)
	}
	return keys
}

var closeServer = make(chan bool)

func startAuthentication() tea.Cmd {
	return func() tea.Msg {
		p := globals.GetProgram()
		port := rand.Intn(20) + 7500
		localAddress := fmt.Sprintf("http://localhost:%d", port)
		callbackUrl := localAddress + "/callback"
		srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
		code, state, err := lib.GetRequestToken(callbackUrl)
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			token, username, err := lib.GetAccesToken(state, code)
			if err != nil {
				p.Send(authResult{authFailure: err.Error()})
			} else {
				p.Send(authResult{accessToken: token, username: username})
			}
			fmt.Fprint(w, "Authentication successful. You can close this tab now.")
		})
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				p.Send(authResult{authFailure: err.Error()})
			}
		}()
		if err != nil {
			return authResult{authFailure: err.Error()}
		} else {
			go func() {
				<-closeServer
				srv.Shutdown(context.Background())
			}()
			lib.OpenAuthorizationURL(code, callbackUrl)
			return authResult{openBrowser: true}
		}
	}
}

func loadSaves(m model) tea.Cmd {
	if m.IsAuthenticated() {
		m.message.SetShow(true)
		m.message.SetLabel("Refreshing saves...")
		return func() tea.Msg {
			saves, err := db.GetPocketSaves()
			if (err != nil && err == db.NoSavesErr) || len(saves) == 0 {
				response, err := lib.GetAllPocketSaves(m.user.AccessToken, 0)
				saves = response.Saves
				if err != nil {
					return getSavesResult{err: err}
				}
				saves, err := db.InsertSaves(response.Since, saves)
				if err != nil {
					return getSavesResult{err: err}
				}
				sort.Sort(models.ByAddedOnDesc(saves))
				return getSavesResult{count: len(saves), saves: saves}
			} else if err != nil {
				return getSavesResult{err: err}
			} else {
				return getSavesResult{count: len(saves), saves: saves}
			}
		}
	} else {
		return nil
	}
}

func refreshSaves(m model) tea.Cmd {
	if m.IsAuthenticated() {
		return func() tea.Msg {
			response, err := lib.GetAllPocketSaves(m.user.AccessToken, float64(m.user.SavesUpdatedOn))
			if err != nil {
				return getSavesResult{err: err}
			}
			saves, err := db.InsertSaves(response.Since, response.Saves)
			if err != nil {
				return getSavesResult{err: err}
			}
			sort.Sort(models.ByAddedOnDesc(saves))
			return getSavesResult{count: len(saves), saves: saves}
		}
	} else {
		return nil
	}
}
