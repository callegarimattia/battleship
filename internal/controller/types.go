package controller

// GameInfo contains the current status of the game.
// It is returned by the Info() method and provides a snapshot of the game.
type GameInfo struct {
	ID          string    `json:"id"`
	Phase       GamePhase `json:"phase"`
	PlayerIDs   []string  `json:"playerIds"`
	CurrentTurn string    `json:"currentTurn"`
	Winner      string    `json:"winner,omitempty"`
}

// GamePhase represents the current state of such game.
type GamePhase string

// Constants for GamePhase
const (
	PhaseWaiting  GamePhase = "Lobby"
	PhaseSetup    GamePhase = "Setup"
	PhasePlay     GamePhase = "Play"
	PhaseGameOver GamePhase = "GameOver"
)
