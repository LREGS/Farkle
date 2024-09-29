package view

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
)

type BaseGame struct {
	term   string
	bg     string
	log    *slog.Logger
	screen string
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

func InitialModel(bg, term string) *BaseGame {
	return &BaseGame{
		term:   term,
		bg:     bg,
		screen: "hello",
	}
}

func (b BaseGame) Init() tea.Cmd {
	return nil
}

func (b BaseGame) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return b, tea.Quit
		}
	}
	return b, nil
}

func (b BaseGame) View() string {
	return b.screen
}
