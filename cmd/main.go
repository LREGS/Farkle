package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
)

const (
	host = "localhost"
	port = "6161"
)

// our server with a list of all current bubbletea programs running (essentialy a connection list)
// we can handle communication between the players here using p.Send() functions to communicate between
// bt instances.

// The app will also be embedded into each Bt instance so that the users of the app itself
// are exposed to the methods on the server. This leads to user initiated communication etc
type app struct {
	*ssh.Server
	progs []*tea.Program
}

// send will despatch a tea.Msg to all the apps within the server. Can be used to send any type of
// tea.Msg that will be defined within the app itself

// so - does this mean that say during init() function we can just send a ConnectionSuccess tea.Msg
// that will broadcast to all new apps that a new connection has been made

// I think it is within this function that we need a switch/case style too
// that depending on which msg is being sent will also communicate with the game proper to calculate scores and state etc
func (a *app) send(msg tea.Msg) {
	for _, p := range a.progs {
		go p.Send(msg)
	}
}

// NewServer initializes the server and returns the Server struct
func NewApp() *app {
	a := new(app)
	srv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"), // Ensure this path is valid
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(a.ProgramHandler, termenv.ANSI256),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Failed to create server: ", err)
		return nil // Return nil if server creation fails
	}

	a.Server = srv

	return a
}

func (a *app) Start() {
	var err error
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = a.ListenAndServe(); err != nil {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := a.Shutdown(ctx); err != nil {
		log.Error("Could not stop server", "error", err)
	}
}

func (a *app) ProgramHandler(s ssh.Session) *tea.Program {
	m := initialModel()
	m.app = a
	m.playerName = s.User()
	m.players = append(m.players, s.User())

	p := tea.NewProgram(m, bubbletea.MakeOptions(s)...)
	a.progs = append(a.progs, p)

	return p
}

type (
	ReadyMsg struct {
		msg string
	}
)

type model struct {
	*app
	playerName string
	players    []string
}

func initialModel() model {
	return model{players: []string{}}
}

func (m model) Init() tea.Cmd {
	return nil
}

// Update processes incoming messages and updates the model state accordingly
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.app.send(ReadyMsg{msg: m.playerName})
		}
	case ReadyMsg:
		m.players = append(m.players, msg.msg)
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	s := ""
	for _, player := range m.players {
		s += player
	}

	return s
}

func main() {

	app := NewApp()
	app.Start()
}
