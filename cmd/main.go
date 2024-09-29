package main

import (
	"net"
	"sync"
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

type connection struct {
	game *tea.Program
	Name string
}

type Server struct {
	server      *ssh.Server
	connections []*connection
	connNames   chan string
	mu          sync.Mutex // Added Mutex to prevent race conditions
}

// NewServer initializes the server and returns the Server struct
func NewServer() *Server {
	s := &Server{
		connections: make([]*connection, 0),
		connNames:   make(chan string, 10), // Buffered channel to avoid deadlock
	}

	srv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"), // Ensure this path is valid
		wish.WithMiddleware(
			s.StartInitialGame(),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Failed to create server: ", err)
		return nil // Return nil if server creation fails
	}

	s.server = srv

	return s
}

// CurrPlayerNames returns a list of the names of all currently connected players
func (s *Server) CurrPlayerNames() []string {
	s.mu.Lock()         // Lock the connection list to prevent race conditions
	defer s.mu.Unlock() // Unlock after accessing the connection list

	currConns := []string{}
	for _, conn := range s.connections {
		currConns = append(currConns, conn.Name)
	}

	return currConns
}

type connectionMessage string

// StartInitialGame sets up the game for new SSH connections
func (s *Server) StartInitialGame() wish.Middleware {

	// using a pointer so it survives on the heap when the function exits before callbacks are called
	conn := &connection{}

	newProg := func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
		p := tea.NewProgram(m, opts...)
		conn.game = p
		go func() {
			// Receive new player name from channel and send it to the program
			newPlayers := <-s.connNames
			p.Send(connectionMessage(newPlayers))
		}()
		return p
	}

	teaHandler := func(ss ssh.Session) *tea.Program {
		pty, _, active := ss.Pty()
		if !active {
			wish.Fatalln(ss, "No active terminal, skipping")
			return nil
		}

		conn.Name = ss.User()

		// Log and send the connection name to the channel
		log.Info("Sending connection name to channel: ", conn.Name)
		s.connNames <- conn.Name

		// Build the model with current player names and terminal settings
		m := model{
			players: s.CurrPlayerNames(),
			term:    pty.Term,
			width:   pty.Window.Width,
			height:  pty.Window.Height,
			time:    time.Now(),
		}
		return newProg(m, append(bubbletea.MakeOptions(ss), tea.WithAltScreen())...)
	}

	// Safely append the connection using the mutex
	s.mu.Lock()
	s.connections = append(s.connections, conn)
	s.mu.Unlock()

	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

type model struct {
	players []string
	term    string
	width   int
	height  int
	time    time.Time
}

type timeMsg time.Time

func (m model) Init() tea.Cmd {
	return nil
}

// Update processes incoming messages and updates the model state accordingly
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timeMsg:
		m.time = time.Time(msg)
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case connectionMessage:
		// Append new players to the list
		m.players = append(m.players, string(msg))
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View returns a string representation of the model (displayed to the user)
func (m model) View() string {
	// Format player names for display
	var cPlayers string
	for _, p := range m.players {
		cPlayers += p + "\n"
	}
	return cPlayers
}

func main() {
	// Create the server instance
	s := NewServer()

	// If server creation fails, exit
	if s == nil {
		log.Fatal("Failed to initialize the server")
	}

	// Log that the server is starting
	log.Infof("Starting SSH server on %s:%s", host, port)

	// Start the server and listen for incoming connections
	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
