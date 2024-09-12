package view

import (
	"log/slog"

	"github.com/charmbracelet/ssh"
)

type BaseGame struct {
	term   string
	bg     string
	log    *slog.Logger
	player Player
}

// player is derived from the ssh session info passed to the handler
type Player struct {
	// Username provided when player connects
	user string
	// contains info on the terminal dimensions within the window - could maybe just pass window
	Pty ssh.Pty

	// winChanges <- chan Window
	// addr net.Addr
	// key gossh.PublicKey
}

func (b *BaseGame) Init() {

}
