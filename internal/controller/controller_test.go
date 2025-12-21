package controller_test

import (
	"testing"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/model"
)

func TestController_Join(t *testing.T) {
	c := controller.NewController()

	t.Run("Game starts in Waiting Phase", func(t *testing.T) {
		if c.Info().Phase != controller.PhaseWaiting {
			t.Errorf("Expected PhaseWaiting, got %v", c.Info().Phase)
		}
	})

	t.Run("transitions correctly", func(t *testing.T) {
		// 1. First player
		id1, err := c.Join()
		if err != nil {
			t.Fatalf("Expected success for 1st player, got %v", err)
		}
		if id1 == "" {
			t.Error("Got empty ID string")
		}
		if c.Info().Phase != controller.PhaseWaiting {
			t.Error("Phase should remain Waiting with only 1 player")
		}

		// 2. Second player
		id2, err := c.Join()
		if err != nil {
			t.Fatalf("Expected success for 2nd player, got %v", err)
		}
		if id2 == "" {
			t.Error("Got empty ID string")
		}
		if c.Info().Phase != controller.PhaseSetup {
			t.Errorf("Expected PhaseSetup, got %v", c.Info().Phase)
		}

		// 3. Third player (Full)
		_, err = c.Join()
		if err != controller.ErrGameFull {
			t.Errorf("Expected ErrGameFull, got %v", err)
		}
	})
}

func TestController_PlaceShip(t *testing.T) {
	// Helper for clean tests
	type testCase struct {
		name    string
		setup   func() (*controller.Controller, string) // Returns controller and playerID
		ship    model.ShipType
		start   model.Coordinate
		orient  model.Orientation
		wantErr error
	}

	tests := []testCase{
		{
			name: "Valid placement",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				return c, p1
			},
			ship:    model.Carrier,
			start:   model.Coordinate{X: 0, Y: 0},
			orient:  model.Horizontal,
			wantErr: nil,
		},
		{
			name: "Rejects placement in Waiting phase",
			setup: func() (*controller.Controller, string) {
				c := controller.NewController()
				p1, _ := c.Join()
				return c, p1
			},
			ship:    model.Carrier,
			start:   model.Coordinate{X: 0, Y: 0},
			orient:  model.Horizontal,
			wantErr: controller.ErrWrongGamePhase,
		},
		{
			name: "Rejects unknown player",
			setup: func() (*controller.Controller, string) {
				c, _, _ := setupLobby(t)
				return c, "hacker"
			},
			ship:    model.Carrier,
			start:   model.Coordinate{X: 0, Y: 0},
			orient:  model.Horizontal,
			wantErr: controller.ErrUnknownPlayer,
		},
		{
			name: "Duplicate ship (Inventory)",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				_ = c.PlaceShip(p1, model.Carrier, model.Coordinate{X: 0, Y: 0}, model.Horizontal)
				return c, p1
			},
			ship:    model.Carrier,
			start:   model.Coordinate{X: 5, Y: 5},
			orient:  model.Horizontal,
			wantErr: model.ErrShipTypeDepleted,
		},
		{
			name: "Overlap error",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				_ = c.PlaceShip(p1, model.Carrier, model.Coordinate{X: 0, Y: 0}, model.Horizontal)
				return c, p1
			},
			ship:    model.Destroyer,
			start:   model.Coordinate{X: 0, Y: 0},
			orient:  model.Vertical,
			wantErr: model.ErrShipOverlap,
		},
		{
			name: "Invalid ship type",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				return c, p1
			},
			ship:    "SpaceShip",
			start:   model.Coordinate{X: 0, Y: 0},
			orient:  model.Horizontal,
			wantErr: model.ErrInvalidShip,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, pid := tt.setup()
			err := c.PlaceShip(pid, tt.ship, tt.start, tt.orient)

			if err != tt.wantErr {
				t.Errorf("PlaceShip() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_Ready(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*controller.Controller, string)
		wantErr error
	}{
		{
			name: "Valid Ready (Setup to Play transition check done in Lifecycle)",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				placeFullFleet(t, c, p1)
				return c, p1
			},
			wantErr: nil,
		},
		{
			name: "Rejects incorrect phase (Waiting)",
			setup: func() (*controller.Controller, string) {
				c := controller.NewController()
				p1, _ := c.Join()
				return c, p1
			},
			wantErr: controller.ErrWrongGamePhase,
		},
		{
			name: "Rejects unknown player",
			setup: func() (*controller.Controller, string) {
				c, _, _ := setupLobby(t)
				return c, "ghost"
			},
			wantErr: controller.ErrUnknownPlayer,
		},
		{
			name: "Rejects incomplete fleet",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				return c, p1
			},
			wantErr: model.ErrFleetIncomplete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, pid := tt.setup()
			err := c.Ready(pid)

			if err != tt.wantErr {
				t.Errorf("Ready() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_Fire(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() (*controller.Controller, string)
		target     model.Coordinate
		wantResult model.ShotResult
		wantErr    error
	}{
		{
			name: "Valid Hit",
			setup: func() (*controller.Controller, string) {
				// P1 vs P2 (P2 has ship at 0,0)
				c, p1, _ := setupActiveGame(t)
				return c, p1
			},
			target:     model.Coordinate{X: 0, Y: 0},
			wantResult: model.ResultHit,
			wantErr:    nil,
		},
		{
			name: "Valid Miss",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupActiveGame(t)
				return c, p1
			},
			target:     model.Coordinate{X: 9, Y: 9},
			wantResult: model.ResultMiss,
			wantErr:    nil,
		},
		{
			name: "Rejects incorrect phase",
			setup: func() (*controller.Controller, string) {
				c := controller.NewController()
				p1, _ := c.Join()
				return c, p1
			},
			target:     model.Coordinate{X: 0, Y: 0},
			wantResult: model.ResultInvalid,
			wantErr:    controller.ErrWrongGamePhase,
		},
		{
			name: "Rejects unknown player",
			setup: func() (*controller.Controller, string) {
				c, _, _ := setupActiveGame(t)
				return c, "random"
			},
			target:     model.Coordinate{X: 0, Y: 0},
			wantResult: model.ResultInvalid,
			wantErr:    controller.ErrUnknownPlayer,
		},
		{
			name: "Enforces turn order (Not your turn)",
			setup: func() (*controller.Controller, string) {
				c, _, p2 := setupActiveGame(t)
				return c, p2 // P2 trying to fire during P1's turn
			},
			target:     model.Coordinate{X: 0, Y: 0},
			wantResult: model.ResultInvalid,
			wantErr:    controller.ErrNotYourTurn,
		},
		{
			name: "Out of bounds error",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupActiveGame(t)
				return c, p1
			},
			target:     model.Coordinate{X: 100, Y: 100},
			wantResult: model.ResultInvalid,
			wantErr:    model.ErrOutOfBounds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, pid := tt.setup()
			res, err := c.Fire(pid, tt.target)

			if err != tt.wantErr {
				t.Errorf("Fire() error = %v, wantErr %v", err, tt.wantErr)
			}
			if res != tt.wantResult {
				t.Errorf("Fire() result = %v, wantResult %v", res, tt.wantResult)
			}
		})
	}
}

func TestController_GameLifecycle_TransitionToPlay(t *testing.T) {
	c, p1, p2 := setupLobby(t)

	// 1. P1 places full fleet
	placeFullFleet(t, c, p1)
	_ = c.Ready(p1)

	// 2. Verify Phase is still Setup (P2 not ready)
	if c.Info().Phase != controller.PhaseSetup {
		t.Errorf("Game should remain in Setup until P2 is ready. Got: %v", c.Info().Phase)
	}

	// 3. P2 places full fleet
	placeFullFleet(t, c, p2)
	_ = c.Ready(p2)

	// 4. Verify Phase transitions to Play
	if c.Info().Phase != controller.PhasePlay {
		t.Errorf("Game should transition to PhasePlay when both are ready. Got: %v", c.Info().Phase)
	}

	// 5. Verify Fire works now
	_, err := c.Fire(p1, model.Coordinate{X: 0, Y: 0})
	if err != nil {
		t.Errorf("Expected Fire to work after transition, got: %v", err)
	}
}

func TestController_GameLifecycle_WinCondition(t *testing.T) {
	c, p1, p2 := setupActiveGame(t)

	// Helper to sink a specific ship type at a specific row
	// P1 shoots at P2's ships
	sinkShip := func(row int, size int) {
		for x := range size {
			_, err := c.Fire(p1, model.Coordinate{X: x, Y: row})
			if err != nil {
				t.Fatalf("Failed to fire during sink sequence at %d,%d: %v", x, row, err)
			}
			// Burn the defender's turn (shoot water at 9,9)
			_, _ = c.Fire(p2, model.Coordinate{X: 9, Y: 9})
		}
	}

	// Sink first 4 ships
	sinkShip(0, 5) // Carrier
	sinkShip(1, 4) // Battleship
	sinkShip(2, 3) // Cruiser
	sinkShip(3, 3) // Submarine

	// Sink the final Destroyer (Size 2 at Row 4)

	// 1. Hit first part
	res, _ := c.Fire(p1, model.Coordinate{X: 0, Y: 4})
	if res != model.ResultHit {
		t.Errorf("Expected Hit, got %v", res)
	}
	_, _ = c.Fire(p2, model.Coordinate{X: 9, Y: 9}) // Burn turn

	// 2. Hit last part (Winning Shot)
	res, _ = c.Fire(p1, model.Coordinate{X: 1, Y: 4})
	if res != model.ResultSunk {
		t.Fatalf("Expected ResultSunk on winning shot, got %v", res)
	}

	// Verify Game State
	info := c.Info()
	if info.Phase != controller.PhaseGameOver {
		t.Fatalf("Expected PhaseGameOver, got %v", info.Phase)
	}
	if info.Winner != p1 {
		t.Fatalf("Expected winner to be %s, got %s", p1, info.Winner)
	}

	// Verify further shots are blocked
	_, err := c.Fire(p2, model.Coordinate{X: 5, Y: 5})
	if err != controller.ErrGameOver {
		t.Fatalf("Expected ErrGameOver for post-game shot, got %v", err)
	}
}
