package service

import (
	"context"

	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/callegarimattia/battleship/internal/model"
)

// PlaceShip handles the complex logic of setup.
// It bridges the gap between simple inputs (bool, int) and Model types (Orientation, pointers).
func (s *MemoryService) PlaceShip(
	_ context.Context,
	matchID, playerID string,
	size, x, y int,
	vertical bool,
) (dto.GameView, error) {
	sg, err := s.getSafeGame(matchID)
	if err != nil {
		return dto.GameView{}, err
	}

	sg.mu.Lock()
	defer sg.mu.Unlock()

	orientation := model.Horizontal
	if vertical {
		orientation = model.Vertical
	}

	coord := model.Coordinate{X: x, Y: y}

	if err := sg.game.PlaceShip(playerID, coord, size, orientation); err != nil {
		return dto.GameView{}, err // Returns ErrShipOverlap, ErrNoShipsRemaining, etc.
	}

	_ = sg.game.StartGame()

	return sg.game.GetView(playerID)
}

// Attack handles the firing logic.
func (s *MemoryService) Attack(
	_ context.Context,
	matchID, playerID string,
	x, y int,
) (dto.GameView, error) {
	sg, err := s.getSafeGame(matchID)
	if err != nil {
		return dto.GameView{}, err
	}

	sg.mu.Lock()
	defer sg.mu.Unlock()

	coord := model.Coordinate{X: x, Y: y}
	if _, err := sg.game.Attack(playerID, coord); err != nil {
		return dto.GameView{}, err // Returns ErrNotYourTurn, ErrInvalidShot, etc.
	}

	return sg.game.GetView(playerID)
}

// GetState retrieves the current game state for a player.
func (s *MemoryService) GetState(
	_ context.Context,
	matchID, playerID string,
) (dto.GameView, error) {
	sg, err := s.getSafeGame(matchID)
	if err != nil {
		return dto.GameView{}, err
	}

	sg.mu.Lock()
	defer sg.mu.Unlock()

	return sg.game.GetView(playerID)
}
