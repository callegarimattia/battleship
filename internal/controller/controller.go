package controller

import (
	"github.com/callegarimattia/battleship/internal/model"
	"github.com/google/uuid"
)

type GameController interface {
	Info() GameInfo
	Join() (string, error)
	PlaceShip(string, model.ShipType, model.Coordinate, model.Orientation) error
	Ready(string) error
	Fire(string, model.Coordinate) (model.ShotResult, error)
}

// Controller manages the game state and logic for a Battleship game.
// It handles player interactions, validating requests, and updating the state.
type Controller struct {
	ID      string
	players map[string]*model.Player
	turn    string
	phase   GamePhase
	winner  string
}

// NewController creates a new Controller instance.
// The game starts in the PhaseWaiting phase.
func NewController() *Controller {
	return &Controller{
		ID:      uuid.NewString(),
		players: make(map[string]*model.Player),
		phase:   PhaseWaiting,
	}
}

// Info returns the current game state information, including
// phase, current turn, players, and winner.
func (c *Controller) Info() GameInfo {
	return GameInfo{
		ID:          c.ID,
		Phase:       c.phase,
		PlayerIDs:   c.playerIDs(),
		CurrentTurn: c.turn,
		Winner:      c.winner,
	}
}

// Join allows a new player to join the game.
// It returns the new player's ID and nil error on success.
// It returns ErrGameFull if the game already has two players or is not in PhaseWaiting.
func (c *Controller) Join() (string, error) {
	if c.phase != PhaseWaiting {
		return "", ErrGameFull
	}

	newPlayer := model.NewPlayer()
	c.players[newPlayer.ID()] = newPlayer

	if c.turn == "" {
		c.turn = newPlayer.ID()
	}

	if len(c.players) == 2 {
		c.phase = PhaseSetup
	}

	return newPlayer.ID(), nil
}

// PlaceShip places a ship for the specified player.
// It accepts model types for ship, position, and orientation.
// It converts internal/model errors into controller-specific errors.
//
// Possible errors:
//   - ErrWrongGamePhase: If game is not in PhaseSetup.
//   - ErrUnknownPlayer: If playerID is invalid.
//   - ErrOverlap: If the ship overlaps with an existing ship.
//   - ErrShipTypeDepleted: If the player has no more ships of this type to place.
//   - ErrInvalidCoordinates: If placements are out of bounds.
func (c *Controller) PlaceShip(playerID string, ship model.ShipType, start model.Coordinate, orientation model.Orientation) error {
	player, err := c.validateRequest(playerID, PhaseSetup)
	if err != nil {
		return err
	}

	// Delegate to Model
	return player.PlaceShip(ship, start, orientation)
}

// Ready marks a player as ready to start the game.
// A player can only be ready if they have placed all their ships.
// When both players are ready, the game transitions to PhasePlay.
//
// Possible errors:
//   - ErrWrongGamePhase: If game is not in PhaseSetup.
//   - ErrUnknownPlayer: If playerID is invalid.
//   - ErrFleetIncomplete: If the player has not placed all required ships.
func (c *Controller) Ready(playerID string) error {
	player, err := c.validateRequest(playerID, PhaseSetup)
	if err != nil {
		return err
	}

	if err := player.SetReady(); err != nil {
		return err
	}

	if c.getOpponent(playerID).IsReady() {
		c.phase = PhasePlay
	}

	return nil
}

// Fire processes a player's attack on the opponent.
// It validates turn order and game phase before executing the attack.
//
// Possible errors:
//   - ErrWrongGamePhase: If game is not in PhasePlay.
//   - ErrUnknownPlayer: If attackerID is invalid.
//   - ErrNotYourTurn: If it is not the attacker's turn.
//   - ErrInvalidCoordinates: If target coordinates are out of bounds.
//   - ErrGameOver: If the game has already ended.
//
// Returns the result of the shot (Hit, Miss, Sunk) and nil error on success.
func (c *Controller) Fire(attackerID string, target model.Coordinate) (model.ShotResult, error) {
	attacker, err := c.validateRequest(attackerID, PhasePlay)
	if err != nil {
		return model.ResultInvalid, err
	}

	defender := c.getOpponent(attackerID)
	// Defender should exist if game is in PhasePlay
	if defender == nil {
		return model.ResultInvalid, ErrUnknownPlayer
	}

	// Perform Action
	modelResult, err := defender.ReceiveAttack(target)
	if err != nil {
		return model.ResultInvalid, err
	}

	// Error is ignored because coordinates are verified by ReceiveAttack
	_ = attacker.MarksShotResult(target, modelResult)

	c.advanceGame(defender.ID())

	return modelResult, nil
}

// validateRequest checks if the request is valid for the current game state.
func (c *Controller) validateRequest(playerID string, expectedPhase GamePhase) (*model.Player, error) {
	if c.phase == PhaseGameOver {
		return nil, ErrGameOver
	}

	if c.phase != expectedPhase {
		return nil, ErrWrongGamePhase
	}

	player, exists := c.players[playerID]
	if !exists {
		return nil, ErrUnknownPlayer
	}

	if expectedPhase == PhasePlay && c.turn != playerID {
		return nil, ErrNotYourTurn
	}

	return player, nil
}

func (c *Controller) getOpponent(id string) *model.Player {
	for pid, p := range c.players {
		if pid != id {
			return p
		}
	}
	return nil
}

func (c *Controller) advanceGame(nextTurnID string) {
	defender := c.players[nextTurnID]
	if defender.HasLost() {
		c.phase = PhaseGameOver
		c.winner = c.turn
		return
	}
	c.turn = nextTurnID
}

func (c *Controller) playerIDs() []string {
	ids := make([]string, 0, len(c.players))
	for id := range c.players {
		ids = append(ids, id)
	}
	return ids
}
