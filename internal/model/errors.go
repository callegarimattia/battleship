package model

import "errors"

var (
	ErrOutOfBounds       = errors.New("coordinates are out of bounds")
	ErrInvalidShip       = errors.New("unknown ship type")
	ErrShipOverlap       = errors.New("ships cannot overlap")
	ErrShipTypeDepleted  = errors.New("no remaining ships of this type to place")
	ErrFleetIncomplete   = errors.New("not all ships have been placed")
	ErrInvalidShotResult = errors.New("invalid shot result")
	ErrRepeatedHit       = errors.New("already hit this ship location")
)
