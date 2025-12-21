package controller

import (
	"github.com/callegarimattia/battleship/internal/model"
	"github.com/google/uuid"
)

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
// It accepts primitive types for ship name (string) and orientation (string).
// It converts internal/model errors into controller-specific errors.
//
// Possible errors:
//   - ErrWrongGamePhase: If game is not in PhaseSetup.
//   - ErrUnknownPlayer: If playerID is invalid.
//   - ErrInvalidShipType: If shipName is unknown.
//   - ErrInvalidOrientation: If orientation is invalid.
//   - ErrOverlap: If the ship overlaps with an existing ship.
//   - ErrShipTypeDepleted: If the player has no more ships of this type to place.
//   - ErrInvalidCoordinates: If placements are out of bounds.
func (c *Controller) PlaceShip(playerID string, shipName string, x, y int, orientation string) error {
	player, err := c.validateRequest(playerID, PhaseSetup)
	if err != nil {
		return err
	}

	sType, err := c.parseShipType(shipName)
	if err != nil {
		return err
	}

	orient, err := c.parseOrientation(orientation)
	if err != nil {
		return err
	}

	// Delegate to Model and map errors
	err = player.PlaceShip(sType, c.toModelCoordinate(x, y), orient)
	return c.mapModelError(err)
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
		return c.mapModelError(err)
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
func (c *Controller) Fire(attackerID string, x, y int) (string, error) {
	attacker, err := c.validateRequest(attackerID, PhasePlay)
	if err != nil {
		return ResultInvalid, err
	}

	defender := c.getOpponent(attackerID)
	// Defender should exist if game is in PhasePlay
	if defender == nil {
		return ResultInvalid, ErrUnknownPlayer
	}

	coords := c.toModelCoordinate(x, y)

	// Perform Action
	modelResult, err := defender.ReceiveAttack(coords)
	if err != nil {
		return ResultInvalid, c.mapModelError(err)
	}

	// Error is ignored because coordinates are verified by ReceiveAttack
	_ = attacker.MarksShotResult(coords, modelResult)

	c.advanceGame(defender.ID())

	return c.formatResult(modelResult), nil
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

// mapModelError maps model errors to controller errors.
func (c *Controller) mapModelError(err error) error {
	switch err {
	case model.ErrOutOfBounds:
		return ErrInvalidCoordinates
	case model.ErrShipOverlap:
		return ErrOverlap
	case model.ErrShipTypeDepleted:
		return ErrShipTypeDepleted
	case model.ErrFleetIncomplete:
		return ErrFleetIncomplete
	case model.ErrRepeatedHit:
		return ErrAlreadyAttacked

	default:
		return err
	}
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
