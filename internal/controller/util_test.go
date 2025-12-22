package controller_test

import (
	"testing"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/model"
)

// setupLobby creates a controller and joins two players, returning the controller and IDs.
// The game is in PhaseSetup.
func setupLobby(t *testing.T) (*controller.Controller, string, string) {
	t.Helper()

	c := controller.NewController()
	p1, _ := c.Join()
	p2, _ := c.Join()

	return c, p1, p2
}

// placeFullFleet places a non-overlapping standard fleet for a specific player.
// Rows: 0=Carrier, 1=Battleship, 2=Cruiser, 3=Submarine, 4=Destroyer
func placeFullFleet(t *testing.T, c *controller.Controller, pid string) {
	t.Helper()
	ships := []model.ShipType{
		model.Carrier,
		model.Battleship,
		model.Cruiser,
		model.Submarine,
		model.Destroyer,
	}

	for row, s := range ships {
		err := c.PlaceShip(pid, s, model.Coordinate{X: 0, Y: row}, model.Horizontal)
		if err != nil {
			t.Fatalf("Failed to place ship %s for %s: %v", s, pid, err)
		}
	}
}

// setupActiveGame creates a fully initialized game in PhasePlay.
// P1 goes first.
func setupActiveGame(t *testing.T) (*controller.Controller, string, string) {
	t.Helper()

	c, p1, p2 := setupLobby(t)

	placeFullFleet(t, c, p1)
	placeFullFleet(t, c, p2)

	_ = c.Ready(p1)
	_ = c.Ready(p2)

	return c, p1, p2
}

func TestSetupLobby(t *testing.T) {
	t.Parallel()

	t.Run("Test setupLobby utility", func(t *testing.T) {
		t.Parallel()

		c, p1, p2 := setupLobby(t)

		if p1 == "" || p2 == "" {
			t.Fatalf("Expected valid player IDs, got p1: %v, p2: %v", p1, p2)
		}

		if c.Info().Phase != controller.PhaseSetup {
			t.Fatalf("Expected game phase to be Setup, got %v", c.Info().Phase)
		}

		if c.Info().Phase != controller.PhaseSetup {
			t.Fatalf("Expected game to be in PhaseSetup after two joins, got %v", c.Info().Phase)
		}

		if len(c.Info().PlayerIDs) != 2 {
			t.Fatalf("Expected 2 players in game, got %d", len(c.Info().PlayerIDs))
		}
	})

	t.Run("Test setupActiveGame utility", func(t *testing.T) {
		t.Parallel()

		c, p1, _ := setupActiveGame(t)

		if c.Info().Phase != controller.PhasePlay {
			t.Fatalf(
				"Expected game to be in PhasePlay after both players are ready, got %v",
				c.Info().Phase,
			)
		}

		if c.Info().CurrentTurn != p1 {
			t.Fatalf("Expected current turn to be %s, got %s", p1, c.Info().CurrentTurn)
		}
	})
}
