package dto

import "github.com/callegarimattia/battleship/internal/model"

// GameInfo contains the current status of the game.
type GameInfo struct {
	ID          string   `json:"id"`
	Phase       string   `json:"phase"`
	PlayerIDs   []string `json:"playerIds"`
	CurrentTurn string   `json:"currentTurn"`
	Winner      string   `json:"winner,omitempty"`
}

// PlaceShipRequest represents the payload for placing a ship.
type PlaceShipRequest struct {
	PlayerID    string `json:"playerId"`
	ShipName    string `json:"shipName"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Orientation string `json:"orientation"`
}

// FireRequest represents the payload for firing a shot.
type FireRequest struct {
	AttackerID string `json:"attackerId"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
}

// FireResponse represents the result of a shot.
type FireResponse struct {
	Result string `json:"result"`
}

// Coordinate represents a simple X,Y pair for DTO usage if needed.
type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// ToModel converts a dto.Coordinate to a model.Coordinate
func (c Coordinate) ToModel() model.Coordinate {
	return model.Coordinate{X: c.X, Y: c.Y}
}
