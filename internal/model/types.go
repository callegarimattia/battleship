package model

// ShipType represents the type of a ship in the game.
type ShipType string

// Constants for ShipType define the available ships.
const (
	Carrier    ShipType = "Carrier"    // Size 5
	Battleship ShipType = "Battleship" // Size 4
	Cruiser    ShipType = "Cruiser"    // Size 3
	Submarine  ShipType = "Submarine"  // Size 3
	Destroyer  ShipType = "Destroyer"  // Size 2
)

// Size returns the length of the ship in grid cells.
func (s ShipType) Size() int {
	switch s {
	case Carrier:
		return 5
	case Battleship:
		return 4
	case Cruiser:
		return 3
	case Submarine:
		return 3
	case Destroyer:
		return 2
	}
	return 0
}

// ShotResult represents the outcome of a shot fired at a coordinate.
type ShotResult int

// Constants for ShotResult
const (
	ResultInvalid ShotResult = -1
	ResultMiss    ShotResult = iota
	ResultHit
	ResultSunk
)

// Orientation represents the placement direction of a ship.
type Orientation int

// Constants for Orientation
const (
	Vertical Orientation = iota
	Horizontal
)

// Coordinate represents a 2D point on the game board (X, Y).
type Coordinate struct {
	X, Y int
}

// Vector returns the delta (dx, dy) for the orientation.
func (o Orientation) Vector() (dx, dy int) {
	if o == Horizontal {
		return 1, 0
	}
	return 0, 1
}

func calculateSegments(start Coordinate, size int, o Orientation) []Coordinate {
	segments := make([]Coordinate, size)
	dx, dy := o.Vector()

	for i := range size {
		segments[i] = Coordinate{
			X: start.X + (i * dx),
			Y: start.Y + (i * dy),
		}
	}
	return segments
}
