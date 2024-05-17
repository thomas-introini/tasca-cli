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
	"github.com/thomas-introini/pocket-cli/db"
	"github.com/thomas-introini/pocket-cli/globals"
	"github.com/thomas-introini/pocket-cli/lib"
	"github.com/thomas-introini/pocket-cli/models"
	styles "github.com/thomas-introini/pocket-cli/views"
	"github.com/thomas-introini/pocket-cli/views/auth"
	"github.com/thomas-introini/pocket-cli/views/saves"
)

type getSavesResult struct {
	saves []models.PocketSave
	count int
	err   error
}

type authResult struct {
	authFailure  string
	requestToken bool
	openBrowser  bool
	accessToken  string
	username     string
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
	errorMessage   string
	keys           keyMap
}

func (m model) IsAuthenticated() bool {
	return m.user != models.NoUser
}

func (m model) Init() tea.Cmd {
	if m.IsAuthenticated() {
		go loadSaves(m.user)
	}
	return tea.Batch(tea.SetWindowTitle("Pocket CLI"), tea.EnterAltScreen, m.auth.Init(), m.saves.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.window.width, m.window.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if !m.IsAuthenticated() {
				m.authenticating = true
				go startAuthentication()
			}
		}
	case saves.RefreshSavesCmd:
		go refreshSaves(m.user)
	case authResult:
		if !m.authenticating {
			return m, nil
		}
		if msg.authFailure != "" {
			m.auth.SetLabel(msg.authFailure + "\n")
		} else if msg.requestToken {
			m.auth.SetLabel("Retrieving token in progress...\n")
		} else if msg.openBrowser {
			m.auth.SetLabel("Continue authentication in browser...\n")
		} else if msg.accessToken != "" {
			closeServer <- true
			m.authenticating = false
			m.user = models.PocketUser{
				AccessToken: msg.accessToken,
				Username:    msg.username,
			}
			user, err := db.SaveUser(msg.accessToken, msg.username)
			if err != nil {
				m.auth.SetLabel("Could not save user...\n")
				m.authenticating = false
			}
			go loadSaves(user)
			m.saves.SetUser(user)
		}
	case getSavesResult:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
		} else {
			m.saves.SetSaves(msg.saves)
		}
	}
	var saveCmd, authCmd tea.Cmd
	m.saves, saveCmd = m.saves.Update(msg)
	m.auth, authCmd = m.auth.Update(msg)
	return m, tea.Batch(saveCmd, authCmd)
}

func (m model) View() string {
	if m.window.width == 0 {
		return "\n"
	}
	view := ""
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
		view += strings.Repeat("\n", (m.window.height/2)-strings.Count(view, "\n")-1)
		view += m.saves.View()
		return view
	}
	helpView := m.help.View(m.keys)
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

var closeServer = make(chan bool)

func startAuthentication() {
	p := globals.GetProgram()
	port := rand.Intn(20) + 8100
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
		p.Send(authResult{authFailure: err.Error()})
	} else {
		p.Send(authResult{requestToken: true})
		go func() {
			<-closeServer
			srv.Shutdown(context.Background())
		}()
		lib.OpenAuthorizationURL(code, callbackUrl)
		p.Send(authResult{openBrowser: true})
	}
}

func loadSaves(user models.PocketUser) {
	p := globals.GetProgram()
	saves, err := db.GetPocketSaves()
	if (err != nil && err == db.NoSavesErr) || len(saves) == 0 {
		response, err := lib.GetAllPocketSaves(user.AccessToken, 0)
		sort.Sort(models.ByUpdatedOnDesc(response.Saves))
		saves = response.Saves
		if err != nil {
			p.Send(getSavesResult{err: err})
		}
		go db.InsertSaves(response.Since, saves)
		p.Send(getSavesResult{count: len(saves), saves: saves})
	} else if err != nil {
		p.Send(getSavesResult{err: err})
	} else {
		p.Send(getSavesResult{count: len(saves), saves: saves})
	}
}

func refreshSaves(user models.PocketUser) {
	p := globals.GetProgram()
	response, err := lib.GetAllPocketSaves(user.AccessToken, float64(user.SavesUpdatedOn))
	if err != nil {
		p.Send(getSavesResult{err: err})
	}
	err = db.InsertSaves(response.Since, response.Saves)
	if err != nil {
		p.Send(getSavesResult{err: err})
	}
	saves, err := db.GetPocketSaves()
	if err != nil {
		p.Send(getSavesResult{err: err})
	}
	p.Send(getSavesResult{count: len(saves), saves: saves})
}
