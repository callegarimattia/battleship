// Package rules defines the game validation rules for the TUI.
package rules

import (
	"fmt"

	"github.com/callegarimattia/battleship/internal/dto"
)

// CanAttack checks if a cell can be attacked.
// Returns an error if the cell is invalid or already attacked.
func CanAttack(board dto.BoardView, x, y int) error {
	if x < 0 || x >= board.Size || y < 0 || y >= board.Size {
		return fmt.Errorf("coordinates out of bounds: %d,%d", x, y)
	}

	cell := board.Grid[y][x]
	if cell == dto.CellHit || cell == dto.CellMiss || cell == dto.CellSunk {
		return fmt.Errorf("cell already attacked: %d,%d", x, y)
	}

	return nil
}

// CanPlaceShip checks if a ship can be placed at the given coordinates.
// Returns nil if valid, error otherwise.
func CanPlaceShip(
	board dto.BoardView,
	size, x, y int,
	vertical bool,
) error {
	// Check bounds
	if vertical {
		if y+size > board.Size {
			return fmt.Errorf("ship out of bounds")
		}
	} else {
		if x+size > board.Size {
			return fmt.Errorf("ship out of bounds")
		}
	}

	// Check for overlap
	for i := 0; i < size; i++ {
		var cx, cy int
		if vertical {
			cx, cy = x, y+i
		} else {
			cx, cy = x+i, y
		}

		if cx < 0 || cx >= board.Size || cy < 0 || cy >= board.Size {
			return fmt.Errorf("coordinates out of bounds")
		}

		cell := board.Grid[cy][cx]
		if cell != dto.CellEmpty {
			return fmt.Errorf("overlap with existing ship at %d,%d", cx, cy)
		}
	}

	return nil
}
