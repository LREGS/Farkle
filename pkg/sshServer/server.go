package SshServer

import (
	"net"

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
	port = "6969"
)

// The server is responsible for handling and grouping incoming connections, and once ready, starting bt games and returning the applications to the players
type SshServer struct {
	srv         *ssh.Server
	connections []string // just holds a copy of the current connections - maybe this should be a pointer to sessions instead
}

func NewSshServer() *SshServer {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(a.ProgramHandler, termenv.ANSI256),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	return &SshServer{
		srv:         s,
		connections: make([]string, 4),
	}

}
