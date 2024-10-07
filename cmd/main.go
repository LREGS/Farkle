package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	c "farkle/common"
	client "farkle/pkg/client"
	view "farkle/pkg/view"

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
	m := InitialBaseModel()
	m.app = a
	// m.playerName = s.User() ** These need changing to be passing game state into the ui object instead
	// m.players = append(m.players, s.User())

	p := tea.NewProgram(m, bubbletea.MakeOptions(s)...)
	a.progs = append(a.progs, p)

	return p
}

type GameClient interface {
	Connect(addr string) error
	Read() (<-chan []byte, <-chan error) // returns channel holding state of the game or error
	Respond([]byte) error                // writes a response to the server - 1 to play atm

}

type tcpResponse []byte

type tcpReadError string

// Base model will hold the UI and Game client. It will be monitored within bubbletea loop
// and call ui and game methods based on user input

type BaseModel struct {
	*app
	log        *log.Logger
	client     GameClient
	UI         *view.UI // this naming is horrible :)
	Display    string
	GameData   <-chan []byte
	GameErrors <-chan error
}

func InitialBaseModel() *BaseModel {
	return &BaseModel{

		client: client.NewClient(),
		UI:     view.NewUI(),
	}
}

type ConnectionSuccess struct{}

// currently this will get stuck on connection failed and like not listen for a new connection or
// something - not sure if its to do with ui state conditional within the update function
type ConnectionFailed struct{ err string }

func (m *BaseModel) AttemptConnection(addr string) tea.Msg {
	err := m.client.Connect(addr)
	if err != nil {
		return ConnectionFailed{err: err.Error()}
	}
	return ConnectionSuccess{}
}

type FailedRespondingToServer struct{}
type SuccessfulResponse struct{}

func (m *BaseModel) sendResponse() tea.Cmd {
	return func() tea.Msg {

		err := m.client.Respond([]byte{'1'})
		if err != nil {
			return FailedRespondingToServer{}
		}
		return SuccessfulResponse{}
	}

}

func (m *BaseModel) monitorChannels() tea.Cmd {

	return func() tea.Msg {
		select {
		case data := <-m.GameData:
			return tcpResponse(data)
		case err := <-m.GameErrors:
			return tcpReadError(err.Error())
		}
	}

}

func (m *BaseModel) Init() tea.Cmd {
	return nil
}

func (m *BaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "c":
			if m.UI.CurrState == view.WelcomeState {
				// Pass the returned message to the Bubble Tea framework
				return m, func() tea.Msg {
					return m.AttemptConnection("localhost:4121")
				}
			}
		case "1":
			if m.UI.CurrState == view.SuccessfulConnection {
			}
		}
		return m, nil
	case ConnectionFailed:
		m.UI.CurrState = view.FailedConnection
		return m, nil
	case ConnectionSuccess:
		m.UI.CurrState = view.SuccessfulConnection
		m.GameData, m.GameErrors = m.client.Read()
		return m, m.monitorChannels()

	case FailedRespondingToServer:

	case tcpResponse:
		var gs c.GameData
		r := bytes.NewReader(msg)
		if err := json.NewDecoder(r).Decode(&gs); err != nil {
			fmt.Print("do something")
		}
		m.UI.Data = gs
		m.UI.CurrState = view.GameLive
		return m, m.monitorChannels()

	}

	return m, nil
}

func (m *BaseModel) View() string {
	// m.log.Print(m.screen)
	return m.UI.Render()
}

// type (
// 	ReadyMsg struct {
// 		msg string
// 	}
// )

// type model struct {
// 	*app
// 	playerName string
// 	players    []string
// 	view       *view.UI
// }

// func initialModel() model {
// 	return model{
// 		players: []string{},
// 		view:    view.NewUI(),
// 	}
// }

// func (m model) Init() tea.Cmd {
// 	return nil
// }

// // Update processes incoming messages and updates the model state accordingly
// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "q", "ctrl+c":
// 			return m, tea.Quit
// 		case "r":
// 			m.app.send(ReadyMsg{msg: m.playerName})
// 		}
// 	case ReadyMsg:
// 		m.players = append(m.players, msg.msg)
// 		return m, nil
// 	}
// 	return m, nil
// }

// func (m model) View() string {
// 	return m.view.Render()
// }

func main() {

	app := NewApp()
	app.Start()
}
