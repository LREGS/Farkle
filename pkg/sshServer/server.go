package SshServer

import (
	"net"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "localhost"
	port = "6161"
)

type Server struct {
	*ssh.Server
	connections []string
}

// middleware is a function that takes a handler and returns a handler - its a piece in a chain of events
// a handler is a callback for handling established sessions. Basically, as the ssh session is established,
// a handler is invoked which is a function that takes a session.

func NewServer() *Server {
	s := &Server{
		connections: make([]string, 0),
	}

	srv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"), //TODO: find out  what to actually put
		wish.WithMiddleware(
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("failed to create server ", err)
	}

	s.Server = srv

	return s

}
