// Package dto contains data transfer objects for representing game state.
package dto

import "time"

// CellState describes what a specific coordinate looks like.
type CellState string

// Possible CellState values.
const (
	CellEmpty   CellState = "EMPTY"
	CellShip    CellState = "SHIP" // Only visible to owner
	CellHit     CellState = "HIT"  // Hit on a ship
	CellMiss    CellState = "MISS" // Hit on water
	CellSunk    CellState = "SUNK" // Part of a sunk ship
	CellUnknown CellState = "???"  // Fog of war
)

// GameState represents the current phase of the game.
type GameState string

// Possible GameState values.
const (
	StateSetup    GameState = "SETUP"
	StatePlaying  GameState = "PLAYING"
	StateFinished GameState = "FINISHED"
)

// BoardView is a simplified, immutable snapshot of the board grid.
// It is safe to pass to the frontend/CLI.
type BoardView struct {
	Grid [][]CellState `json:"grid"`
	Size int           `json:"size"`
}

// PlayerView represents a player's public state.
type PlayerView struct {
	ID    string      `json:"id"`
	Board BoardView   `json:"board"`
	Fleet map[int]int `json:"fleet"` // Remaining ships by size
}

// GameView is the full packet sent to an observer (UI).
type GameView struct {
	State  GameState  `json:"state"`
	Turn   string     `json:"turn"`
	Winner string     `json:"winner,omitempty"`
	Me     PlayerView `json:"me"`
	Enemy  PlayerView `json:"enemy"`
}

// User represents a registered user.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// AuthResponse serves the JWT token along with user info.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// MatchSummary is used for the "Lobby List" screen.
type MatchSummary struct {
	ID          string    `json:"match_id"`
	HostName    string    `json:"host_name"`
	PlayerCount int       `json:"player_count"`
	CreatedAt   time.Time `json:"created_at"`
}
