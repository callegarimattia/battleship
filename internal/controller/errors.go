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
)
