package app

import (
	"context"
	"farkle/view"
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

// pass session to game and  have it fill info from there re name etc, should update automatically as hopefuly itll just push the updated ui (say changed name) to everyone ?! Probably using a p.send to emit a cmd within bt?!

const (
	host = "localhost"
	port = "6969"
)

type app struct {
	*ssh.Server
	players []*tea.Program
}

func newApp() *app {
	a := new(app)
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

	a.Server = s
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

// this does handle how we can return the bt app to our players but not sure on how we're using it at this exact moment
func (a *app) ProgramHandler(s ssh.Session) *tea.Program {

	pty, _, _ := s.Pty()

	bg := "light"
	if renderer.HasDarkBackground() {
		bg = "dark"
	}

	model := view.InitialModel(pty.Term, bg)
	model.app = a
	model.id = s.User()

	p := tea.NewProgram(model, bubbletea.MakeOptions(s)...)
	a.players = append(a.players, p)

	return p
}
