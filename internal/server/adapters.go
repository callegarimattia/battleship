package server

import (
	"errors"
	"strings"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/model"
)

var (
	ErrInvalidOrientation = errors.New("invalid orientation")
	ErrInvalidShipType    = errors.New("invalid ship type")
)

func parseOrientation(orientation string) (model.Orientation, error) {
	orientation = strings.ToLower(orientation)
	switch orientation {
	case "horizontal", "h":
		return model.Horizontal, nil
	case "vertical", "v":
		return model.Vertical, nil
	default:
		return 0, ErrInvalidOrientation
	}
}

func parseShipType(shipType string) (model.ShipType, error) {
	shipType = strings.ToLower(shipType)
	switch shipType {
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

func formatResult(result model.ShotResult) string {
	switch result {
	case model.ResultHit:
		return controller.ResultHit
	case model.ResultMiss:
		return controller.ResultMiss
	case model.ResultSunk:
		return controller.ResultSunk
	default:
		return controller.ResultInvalid
	}
}
