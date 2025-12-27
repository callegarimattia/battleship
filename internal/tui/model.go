// Package tui implements the TUI for Battleship
package tui

import (
	"log"

	"github.com/callegarimattia/battleship/internal/client"
	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/callegarimattia/battleship/internal/env"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SessionState represents the current state of the application.
type SessionState int

const (
	StateLogin SessionState = iota
	StateLobby
	StateGame
)

const BoardSize = 10

// Model is the main TUI model.
type Model struct {
	State  SessionState
	Client *client.Client

	// Login
	LoginInput textinput.Model

	// Lobby
	Matches []dto.MatchSummary
	Cursor  int

	// Game
	GameID   string
	GameView *dto.GameView

	// Game Interaction
	CursorX, CursorY int

	// Setup Phase
	SetupPhase      bool
	ShipsToPlace    []int // sizes
	CurrentShipIdx  int
	ShipOrientation bool // false = horizontal, true = vertical

	// Error Handling
	Err error

	// UI
	Width, Height int
}

func New() *Model {
	cfg, err := env.LoadClientConfig()
	if err != nil {
		log.Fatalf("Failed to load client config: %v", err)
	}

	ti := textinput.New()
	ti.Placeholder = "Commander Name"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 30

	return &Model{
		State:        StateLogin,
		Client:       client.New(cfg.BaseURL),
		LoginInput:   ti,
		ShipsToPlace: []int{5, 4, 3, 3, 2}, // Standard Battleship fleet
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}
