package service

import (
	"context"
	"time"

	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/callegarimattia/battleship/internal/events"
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

	if err := sg.game.PlaceShip(playerID, coord, size, orientation); err != nil { //nolint
		return dto.GameView{}, err // Returns ErrShipOverlap, ErrNoShipsRemaining, etc.
	}

	_ = sg.game.StartGame()
	sg.updatedAt = time.Now()

	view, err := sg.game.GetView(playerID)
	if err != nil {
		return dto.GameView{}, err
	}

	// Emit event: ship placed
	if s.eventBus != nil {
		// Get opponent ID
		opponentID := ""
		if sg.host == playerID {
			opponentID = sg.guest
		} else {
			opponentID = sg.host
		}

		if opponentID != "" {
			s.eventBus.Publish(&events.GameEvent{
				Type:      events.EventShipPlaced,
				MatchID:   matchID,
				PlayerID:  playerID,
				TargetID:  opponentID,
				Timestamp: time.Now(),
				Data: events.ShipPlacedEventData{
					Size:     size,
					X:        x,
					Y:        y,
					Vertical: vertical,
				},
			})
		}
	}

	return view, nil
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
	result, err := sg.game.Attack(playerID, coord)
	if err != nil {
		return dto.GameView{}, err // Returns ErrNotYourTurn, ErrInvalidShot, etc.
	}

	sg.updatedAt = time.Now()

	view, err := sg.game.GetView(playerID)
	if err != nil {
		return dto.GameView{}, err
	}

	// Emit event: attack made
	if s.eventBus != nil {
		// Get opponent ID
		opponentID := ""
		if sg.host == playerID {
			opponentID = sg.guest
		} else {
			opponentID = sg.host
		}

		if opponentID != "" {
			resultStr := "miss"
			switch result {
			case model.ShotResultHit:
				resultStr = "hit"
			case model.ShotResultSunk:
				resultStr = "sunk"
			}

			s.eventBus.Publish(&events.GameEvent{
				Type:      events.EventAttackMade,
				MatchID:   matchID,
				PlayerID:  playerID,
				TargetID:  opponentID,
				Timestamp: time.Now(),
				Data: events.AttackEventData{
					X:      x,
					Y:      y,
					Result: resultStr,
				},
			})
		}
	}

	return view, nil
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
