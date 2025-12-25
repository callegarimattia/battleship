package model

import (
	"errors"
	"maps"
)

var (
	// ErrNotYourTurn is returned when a player tries to act out of turn.
	ErrNotYourTurn = errors.New("not your turn")
	// ErrInvalidShot is returned when a shot is made to an invalid coordinate.
	ErrInvalidShot = errors.New("invalid shot")
	// ErrUnknownPlayer is returned when an action is attempted by an unknown player.
	ErrUnknownPlayer = errors.New("unknown player")
	// ErrNoShipsRemaining is returned when a player tries to place a ship of which they have none left.
	ErrNoShipsRemaining = errors.New("no ships remaining of that size")
	// ErrNotInPlay is returned when a playing action is attempted while the game is not in the playing state.
	ErrNotInPlay = errors.New("game not in playing state")
	// ErrNotInSetup is returned when a setup action is attempted while the game is not in the setup state.
	ErrNotInSetup = errors.New("game not in setup state")
	// ErrNotReadyToStart is returned when trying to start the game before both players have placed all their ships.
	ErrNotReadyToStart = errors.New("not all ships placed by both players")
)

// GameState represents the current phase of the game.
type GameState int

// Possible GameState values.
const (
	StateSetup GameState = iota
	StatePlaying
	StateFinished
)

// Game acts as the refeeree between two players.
// It holds the state and enforces the rules of the game.
type Game struct {
	player1 *Player
	player2 *Player
	turn    string
	state   GameState
	winner  string
}

// Player represents a participant in the Battleship game.
type Player struct {
	id    string
	fleet map[int]int // Remaining ships to place by size
	board *Board
}

// NewGame initializes a new game with two players identified by their IDs.
// A fleet configuration can be provided; if nil, the standard fleet is used.
func NewGame(p1ID, p2ID string, fleet map[int]int) *Game {
	if fleet == nil {
		fleet = StandardFleet()
	}
	return &Game{
		player1: &Player{id: p1ID, board: NewBoard(), fleet: maps.Clone(fleet)},
		player2: &Player{id: p2ID, board: NewBoard(), fleet: maps.Clone(fleet)},
	}
}

// PlaceShip places a ship for the specified player at the given coordinate and orientation.
// Placing a ship can be done only during the setup phase, but turns are not enforced.
func (g *Game) PlaceShip(playerID string, c Coordinate, size int, o Orientation) error {
	if g.state != StateSetup {
		return ErrNotInSetup
	}

	var p *Player
	if p = g.getPlayerByID(playerID); p == nil {
		return ErrUnknownPlayer
	}

	if shipCount, exists := p.fleet[size]; !exists || shipCount <= 0 {
		return ErrNoShipsRemaining
	}

	if err := p.board.PlaceShip(c, &Ship{size}, o); err != nil {
		return err
	}

	p.fleet[size]--

	return nil
}

// StartGame transitions the game from setup to playing state if both players have placed all their ships.
func (g *Game) StartGame() error {
	switch {
	case g.state != StateSetup:
		return ErrNotInSetup
	case !g.allShipsPlaced():
		return ErrNotReadyToStart
	default:
		g.state = StatePlaying
		g.turn = g.player1.id
		return nil
	}
}

// Attack coordinates a shot from the attacker to the defender.
func (g *Game) Attack(attackerID string, c Coordinate) (ShotResult, error) {
	switch {
	case g.state != StatePlaying:
		return ShotResultInvalid, ErrNotInPlay
	case g.getPlayerByID(attackerID) == nil:
		return ShotResultInvalid, ErrUnknownPlayer
	case g.turn != attackerID:
		return ShotResultInvalid, ErrNotYourTurn
	}

	var d *Player
	if d = g.getOpponent(attackerID); d == nil {
		return ShotResultInvalid, ErrNotYourTurn
	}

	switch res := d.board.ReceiveShot(c); res {
	case ShotResultInvalid:
		return ShotResultInvalid, ErrInvalidShot

	case ShotResultSunk:
		if d.board.AllShipsSunk() {
			g.state = StateFinished
			g.winner = attackerID
			return res, nil
		}
		fallthrough

	case ShotResultHit, ShotResultMiss:
		g.passTurn()
		return res, nil
	}

	return ShotResultInvalid, ErrInvalidShot
}

// Winner returns the ID of the winning player if the game has finished; otherwise, it returns an empty string.
func (g *Game) Winner() string { return g.winner }

// StandardFleet returns the standard Battleship fleet configuration.
// It maps ship sizes to their respective counts.
func StandardFleet() map[int]int {
	return map[int]int{
		5: 1, // Carrier
		4: 1, // Battleship
		3: 2, // Cruiser + Submarine
		2: 1, // Destroyer
	}
}

// --- Internal Helper Functions ---

func (g *Game) allShipsPlaced() bool {
	return g.playerShipsPlaced(g.player1) && g.playerShipsPlaced(g.player2)
}

func (g *Game) passTurn() {
	switch g.turn {
	case g.player1.id:
		g.turn = g.player2.id
	case g.player2.id:
		g.turn = g.player1.id
	}
}

func (g *Game) getPlayerByID(playerID string) *Player {
	switch playerID {
	case g.player1.id:
		return g.player1
	case g.player2.id:
		return g.player2
	default: // Unknown player
		return nil
	}
}

func (g *Game) getOpponent(playerID string) *Player {
	switch playerID {
	case g.player1.id:
		return g.player2
	case g.player2.id:
		return g.player1
	default: // Unknown player
		return nil
	}
}

func (g *Game) playerShipsPlaced(p *Player) bool {
	for _, remaining := range p.fleet {
		if remaining > 0 {
			return false
		}
	}
	return true
}
