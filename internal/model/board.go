// Package model provides the core data structures and logic for a Battleship game.
//
//go:generate stringer -type=ShotResult,Orientation,GameState -output=board_string.go
package model

import (
	"errors"
	"iter"
	"slices"
)

var (
	// ErrInvalidDimensions is returned when the board is created with non-positive dimensions.
	ErrInvalidDimensions = errors.New("invalid dimensions")
	// ErrShipOutOfBounds is returned when a ship placement goes out of the board bounds.
	ErrShipOutOfBounds = errors.New("ship placement out of bounds")
	// ErrShipOverlap is returned when a ship placement overlaps with another ship.
	ErrShipOverlap = errors.New("ship placement overlaps with another ship")
	// ErrInvalidShipSize is returned when a ship tries to be created with a non-positive size.
	ErrInvalidShipSize = errors.New("invalid ship size")
)

// GridSize defines the size of the Battleship grid.
const GridSize = 10

type tile struct {
	isHit bool
	ship  *Ship
}

// Board represents the Battleship game board.
type Board struct {
	tiles   [GridSize][GridSize]tile
	history [GridSize][GridSize]ShotResult
}

// ShotResult represents the outcome of a shot fired at a coordinate.
type ShotResult int

// Possible ShotResult values returned when a shot is fired.
const (
	ShotResultInvalid ShotResult = iota
	ShotResultMiss
	ShotResultHit
	ShotResultSunk
)

// Orientation represents the orientation of a ship on the board.
type Orientation int

// Possible Orientation values for placing ships.
const (
	Horizontal Orientation = iota
	Vertical
)

// Coordinate represents a position on the Battleship grid.
type Coordinate struct{ X, Y int }

// Vector returns the row and column deltas for the given orientation.
func (o Orientation) Vector() (dx, dy int) {
	switch o {
	case Horizontal:
		return 1, 0
	case Vertical:
		return 0, 1
	}
	return 0, 0
}

// Ship represent a battleship ship.
type Ship struct{ size int }

// NewShip creates a new Ship with the given size.
func NewShip(s int) (*Ship, error) {
	if s <= 0 {
		return nil, ErrInvalidShipSize
	}
	return &Ship{size: s}, nil
}

// Size returns the size of the ship.
func (s *Ship) Size() int { return s.size }

// NewBoard creates a new Board with the given number of rows and columns.
// Negative or zero dimensions will return an error.
func NewBoard() *Board {
	return &Board{
		tiles:   [GridSize][GridSize]tile{},
		history: [GridSize][GridSize]ShotResult{},
	}
}

// PlaceShip places a ship on the board at the given coordinate with the specified orientation.
// If the ship cannot be placed (e.g., out of bounds or overlapping another ship), an error is returned.
func (b *Board) PlaceShip(c Coordinate, s *Ship, o Orientation) error {
	segments := calculateSegments(c, s.Size(), o)

	if err := b.canPlaceShip(segments); err != nil {
		return err
	}

	b.placeShipAt(segments, s)

	return nil
}

// ReceiveShot processes a shot fired at the given coordinate.
// It returns the result of the shot (hit, miss, sunk, or invalid).
func (b *Board) ReceiveShot(c Coordinate) ShotResult {
	if b.isOutOfBounds(c) {
		return ShotResultInvalid
	}

	t := &b.tiles[c.Y][c.X]
	if t.isHit {
		return ShotResultInvalid
	}

	t.isHit = true

	switch {
	case t.ship == nil: // Miss
		b.history[c.Y][c.X] = ShotResultMiss
		return ShotResultMiss
	case b.isShipSunk(t.ship): // Sunk
		b.history[c.Y][c.X] = ShotResultSunk
		return ShotResultSunk
	default: // Hit
		b.history[c.Y][c.X] = ShotResultHit
		return ShotResultHit
	}
}

// AllShipsSunk checks if every ship on the board has been destroyed.
func (b *Board) AllShipsSunk() bool {
	for _, t := range b.Cells() {
		if t.ship != nil && !t.isHit {
			return false
		}
	}

	return true
}

// Cells returns an iterator over the board.
// It yields the coordinates and a POINTER to the tile.
func (b *Board) Cells() iter.Seq2[Coordinate, *tile] {
	return func(yield func(Coordinate, *tile) bool) {
		for y := range b.tiles {
			for x := range b.tiles[y] {
				if !yield(Coordinate{X: x, Y: y}, &b.tiles[y][x]) {
					return
				}
			}
		}
	}
}

// --- Internal Helper Functions --- //

func (b *Board) isOutOfBounds(c Coordinate) bool {
	return c.Y < 0 || c.Y >= len(b.tiles) || c.X < 0 || c.X >= len(b.tiles[0])
}

func (b *Board) isOccupied(c Coordinate) bool {
	return b.tiles[c.Y][c.X].ship != nil
}

func (b *Board) isShipSunk(s *Ship) bool {
	for _, t := range b.Cells() {
		if t.ship == s && !t.isHit {
			return false
		}
	}
	return true
}

func (b *Board) canPlaceShip(s []Coordinate) error {
	if slices.ContainsFunc(s, b.isOutOfBounds) {
		return ErrShipOutOfBounds
	}

	if slices.ContainsFunc(s, b.isOccupied) {
		return ErrShipOverlap
	}

	return nil
}

func (b *Board) placeShipAt(s []Coordinate, ship *Ship) {
	for _, c := range s {
		b.tiles[c.Y][c.X].ship = ship
	}
}

func calculateSegments(start Coordinate, size int, o Orientation) []Coordinate {
	dx, dy := o.Vector()

	segments := make([]Coordinate, size)
	for i := range segments {
		segments[i] = Coordinate{
			Y: start.Y + i*dy,
			X: start.X + i*dx,
		}
	}

	return segments
}
