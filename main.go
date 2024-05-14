package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thomas-introini/pocket-cli/db"
	"github.com/thomas-introini/pocket-cli/lib"
	"github.com/thomas-introini/pocket-cli/models"
	styles "github.com/thomas-introini/pocket-cli/views"
	"github.com/thomas-introini/pocket-cli/views/auth"
	"github.com/thomas-introini/pocket-cli/views/saves"
)

var POCKET_CONSUMER_KEY = os.Getenv("POCKET_CONSUMER_KEY")

var p *tea.Program

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
	Quit key.Binding
}

func (m keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.Quit},
		{m.Quit},
	}
}

func (m keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		m.Quit,
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
		go loadSaves(m.auth.User)
	}
	return tea.Batch(tea.SetWindowTitle("Pocket CLI"), tea.EnterAltScreen, m.auth.Init(), m.saves.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
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
	case authResult:
		if !m.authenticating {
			return m, nil
		}
		if msg.authFailure != "" {
			m.auth, cmd = m.auth.Update(auth.ChangeLabel{
				Label: msg.authFailure + "\n",
			})
			return m, cmd
		} else if msg.requestToken {
			m.auth, cmd = m.auth.Update(auth.ChangeLabel{
				Label: "Retrieving token in progress...\n",
			})
			return m, cmd
		} else if msg.openBrowser {
			m.auth, cmd = m.auth.Update(auth.ChangeLabel{
				Label: "Continue authentication in browser...\n",
			})
			return m, cmd
		} else if msg.accessToken != "" {
			closeServer <- true
			m.authenticating = false
			m.user = models.PocketUser{
				AccessToken: msg.accessToken,
				Username:    msg.username,
			}
			user, err := db.SaveUser(msg.accessToken, msg.username)
			if err != nil {
				m.auth, _ = m.auth.Update(auth.ChangeLabel{
					Label: "Could not save user...\n",
				})
			}
			go loadSaves(user)
			m.saves.SetUser(user)
			return m, cmd
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
		view += strings.Repeat(" ", (m.window.width-lipgloss.Width(tmp))/2) + tmp // I don't know why -45 it's needed
	} else if !m.IsAuthenticated() && !m.authenticating {
		view += strings.Repeat("\n", (m.window.height/2)-strings.Count(view, "\n")-2)
		tmp := styles.TitleRedStyle.Render("Welcome to Pocket CLI!\n")
		view += strings.Repeat(" ", (m.window.width-lipgloss.Width(tmp))/2) + tmp
		tmp = styles.TitleRedStyle.Render("Press '") + styles.TitleBoldRedStyle.Render("Enter") + styles.TitleRedStyle.Render("' to start the authentication\n")
		view += strings.Repeat(" ", (m.window.width-lipgloss.Width(tmp))-45/2) + tmp // I don't know why -45 it's needed
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

func initModel(user models.PocketUser) model {
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
		},
	}
}

var closeServer = make(chan bool)

func startAuthentication() {
	port := "8124"
	callbackUrl := "http://localhost:" + port + "/callback"
	srv := &http.Server{Addr: ":" + port}
	code, state, err := lib.GetRequestToken(POCKET_CONSUMER_KEY, callbackUrl)
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token, username, err := lib.GetAccesToken(POCKET_CONSUMER_KEY, state, code)
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
	saves, err := db.GetPocketSaves()
	if (err != nil && err == db.NoSavesErr) || len(saves) == 0 {
		response, err := lib.GetAllPocketSaves(user.AccessToken)
		if err != nil {
			p.Send(getSavesResult{err: err})
		}
		fmt.Printf("get %d saves\n", len(response.Saves))
		go db.InsertSaves(int32(response.Since), response.Saves)
		p.Send(getSavesResult{count: len(response.Saves), saves: response.Saves})
	} else if err != nil {
		p.Send(getSavesResult{err: err})
	} else {
		p.Send(getSavesResult{count: len(saves), saves: saves})
	}
}

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "info")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
		l := log.New(f, "pocket ", log.LstdFlags)
		l.Println("test")
	}
	if POCKET_CONSUMER_KEY == "" {
		fmt.Println("set POCKET_CONSUMER_KEY environment variable")
		os.Exit(1)
	}
	err := db.ConnectDB()
	if err != nil {
		fmt.Println("error connecting to database:", err)
		os.Exit(1)
	}
	user, err := db.GetLoggedUser()
	if err != nil && err != db.NoUserErr {
		fmt.Println("Error while retrieving user from database:", err)
		os.Exit(1)
	}
	p = tea.NewProgram(initModel(user))
	if _, err = p.Run(); err != nil {
		fmt.Println("Could not run the program", err)
		os.Exit(1)
	}
}
