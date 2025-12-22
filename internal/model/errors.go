package model

import "errors"

var (
	// ErrOutOfBounds indicates that the provided coordinates are outside the valid game board area.
	ErrOutOfBounds = errors.New("coordinates are out of bounds")
	// ErrInvalidShip indicates that the specified ship type is not recognized.
	ErrInvalidShip = errors.New("unknown ship type")
	// ErrShipOverlap indicates that the ship placement overlaps with an existing ship.
	ErrShipOverlap = errors.New("ships cannot overlap")
	// ErrShipTypeDepleted indicates that there are no remaining ships of the specified type to place.
	ErrShipTypeDepleted = errors.New("no remaining ships of this type to place")
	// ErrFleetIncomplete indicates that not all required ships have been placed.
	ErrFleetIncomplete = errors.New("not all ships have been placed")
	// ErrInvalidShotResult indicates that the shot result provided is not valid.
	ErrInvalidShotResult = errors.New("invalid shot result")
	// ErrRepeatedHit indicates that the same location has already been targeted.
	ErrRepeatedHit = errors.New("already hit this ship location")
)
