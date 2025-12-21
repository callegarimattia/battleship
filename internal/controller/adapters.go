package controller

import (
	"strings"

	"github.com/callegarimattia/battleship/internal/model"
)

func (c *Controller) toModelCoordinate(x, y int) model.Coordinate {
	return model.Coordinate{X: x, Y: y}
}

func (c *Controller) parseShipType(name string) (model.ShipType, error) {
	switch strings.ToLower(name) {
	case "carrier":
		return model.Carrier, nil
	case "battleship":
		return model.Battleship, nil
	case "cruiser":
		return model.Cruiser, nil
	case "submarine":
		return model.Submarine, nil
	case "destroyer":
		return model.Destroyer, nil
	default:
		return "", ErrInvalidShipType
	}
}

func (c *Controller) parseOrientation(o string) (model.Orientation, error) {
	switch strings.ToLower(o) {
	case "horizontal", "h":
		return model.Horizontal, nil
	case "vertical", "v":
		return model.Vertical, nil
	default:
		return 0, ErrInvalidOrientation
	}
}

func (c *Controller) formatResult(r model.ShotResult) string {
	switch r {
	case model.ResultHit:
		return ResultHit
	case model.ResultMiss:
		return ResultMiss
	case model.ResultSunk:
		return ResultSunk
	default:
		return ResultInvalid
	}
}
