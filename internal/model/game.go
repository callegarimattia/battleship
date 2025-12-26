package model

import (
	"errors"
	"maps"

	"github.com/callegarimattia/battleship/internal/dto"
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
	// ErrGameFull is returned when trying to join a game that already has two players.
	ErrGameFull = errors.New("game already has two players")
)

// GameState represents the current phase of the game.
type GameState int

// Possible GameState values.
const (
	StateWaiting GameState = iota
	StateSetup
	StatePlaying
	StateGameOver
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

// IsGameOver returns true if the game is in the finished state.
func (g *Game) IsGameOver() bool {
	return g.state == StateGameOver
}

// Player represents a participant in the Battleship game.
type Player struct {
	id    string
	fleet map[int]int // Remaining ships to place by size
	board *Board
}

// NewFullGame initializes a new game with two players identified by their IDs.
// A fleet configuration can be provided; if nil, the standard fleet is used.
func NewFullGame(p1ID, p2ID string, fleet map[int]int) *Game {
	return &Game{
		player1: &Player{id: p1ID, board: NewBoard(), fleet: startingFleet(fleet)},
		player2: &Player{id: p2ID, board: NewBoard(), fleet: startingFleet(fleet)},
		state:   StateSetup,
	}
}

// NewGame initializes a new empty game.
func NewGame() *Game {
	return &Game{}
}

// Join adds a player to the game with the specified fleet configuration.
func (g *Game) Join(playerID string, fleet map[int]int) error {
	switch {
	case g.player1 == nil:
		g.player1 = &Player{id: playerID, board: NewBoard(), fleet: startingFleet(fleet)}

		return nil
	case g.player2 == nil:
		g.player2 = &Player{id: playerID, board: NewBoard(), fleet: startingFleet(fleet)}

		g.state = StateSetup // Once both players have joined, move to setup phase

		return nil
	default:
		return ErrGameFull
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
			g.state = StateGameOver
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

// GetView returns the DTO seen by a specific observer (playerID).
func (g *Game) GetView(observerID string) (dto.GameView, error) {
	var me, enemy *Player
	switch observerID {
	case g.player1.id:
		me, enemy = g.player1, g.player2
	case g.player2.id:
		me, enemy = g.player2, g.player1
	default:
		return dto.GameView{}, ErrUnknownPlayer
	}

	return dto.GameView{
		State:  toDTOState(g.state),
		Turn:   g.turn,
		Winner: g.winner,
		Me:     me.GetView(false),   // Full view
		Enemy:  enemy.GetView(true), // Fog of war
	}, nil
}

// GetView returns the DTO representation of the player.
func (p *Player) GetView(hideShips bool) dto.PlayerView {
	return dto.PlayerView{
		ID:    p.id,
		Board: p.board.GetSnapshot(hideShips),
		Fleet: maps.Clone(p.fleet),
	}
}

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

func startingFleet(fleet map[int]int) map[int]int {
	if fleet == nil {
		return StandardFleet()
	}
	return maps.Clone(fleet)
}

// Adapter: Convert internal GameState to DTO GameState
func toDTOState(state GameState) dto.GameState {
	switch state {
	case StateSetup:
		return dto.StateSetup
	case StatePlaying:
		return dto.StatePlaying
	case StateGameOver:
		return dto.StateFinished
	default:
		return ""
	}
}
