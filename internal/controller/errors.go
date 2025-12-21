package controller

import "errors"

var (
	// ErrGameFull indicates that the game has reached its maximum number of players (2).
	ErrGameFull = errors.New("game is already full")
	// ErrWrongGamePhase indicates that the requested action is not valid in the current game phase.
	ErrWrongGamePhase = errors.New("action is not allowed in this phase")
	// ErrGameNotStarted indicates that the game has not started yet (e.g. still in waiting or setup).
	ErrGameNotStarted = errors.New("game has not started yet")
	// ErrGameOver indicates that the game has ended and no further actions are allowed.
	ErrGameOver = errors.New("game is already over")
	// ErrNotYourTurn indicates that a player attempted to act out of turn.
	ErrNotYourTurn = errors.New("it is not your turn")
	// ErrUnknownPlayer indicates that the provided player ID does not match any player in the game.
	ErrUnknownPlayer = errors.New("player ID not found")
	// ErrInvalidShipType indicates that the specified ship name is not recognized.
	ErrInvalidShipType = errors.New("invalid ship type")
	// ErrInvalidOrientation indicates that the orientation string is invalid (must be "horizontal", "h", "vertical", "v").
	ErrInvalidOrientation = errors.New("invalid ship orientation")
	// ErrShipTypeDepleted indicates that the player is trying to place more ships of a type than allowed.
	ErrShipTypeDepleted = errors.New("no more ships of this type")
	// ErrOverlap indicates that a ship placement overlaps with an existing ship.
	ErrOverlap = errors.New("ship overlaps with another ship")
	// ErrFleetIncomplete indicates that a player tried to set Ready without placing all ships.
	ErrFleetIncomplete = errors.New("not all ships have been placed")
	// ErrInvalidCoordinates indicates that the provided coordinates are out of bounds.
	ErrInvalidCoordinates = errors.New("coordinates are out of bounds")
	// ErrAlreadyAttacked indicates that the target cell has already been attacked (and was a hit on a ship).
	ErrAlreadyAttacked = errors.New("already attacked this location")
)
