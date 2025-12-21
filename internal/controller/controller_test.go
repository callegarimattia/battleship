package controller_test

import (
	"testing"

	"github.com/callegarimattia/battleship/internal/controller"
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
		ship    string
		x, y    int
		orient  string
		wantErr error
	}

	tests := []testCase{
		{
			name: "Valid placement",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				return c, p1
			},
			ship: "Carrier",
			x:    0, y: 0,
			orient:  "Horizontal",
			wantErr: nil,
		},
		{
			name: "Rejects placement in Waiting phase",
			setup: func() (*controller.Controller, string) {
				c := controller.NewController()
				p1, _ := c.Join()
				return c, p1
			},
			ship: "Carrier",
			x:    0, y: 0,
			orient:  "Horizontal",
			wantErr: controller.ErrWrongGamePhase,
		},
		{
			name: "Rejects unknown player",
			setup: func() (*controller.Controller, string) {
				c, _, _ := setupLobby(t)
				return c, "hacker"
			},
			ship: "Carrier",
			x:    0, y: 0,
			orient:  "Horizontal",
			wantErr: controller.ErrUnknownPlayer,
		},
		{
			name: "Duplicate ship (Inventory)",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				_ = c.PlaceShip(p1, "Carrier", 0, 0, "Horizontal")
				return c, p1
			},
			ship: "Carrier",
			x:    5, y: 5,
			orient:  "Horizontal",
			wantErr: controller.ErrShipTypeDepleted,
		},
		{
			name: "Overlap error",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				_ = c.PlaceShip(p1, "Carrier", 0, 0, "Horizontal")
				return c, p1
			},
			ship: "Destroyer",
			x:    0, y: 0,
			orient:  "Vertical",
			wantErr: controller.ErrOverlap,
		},
		{
			name: "Invalid ship type",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				return c, p1
			},
			ship: "SpaceShip",
			x:    0, y: 0,
			orient:  "h",
			wantErr: controller.ErrInvalidShipType,
		},
		{
			name: "Invalid orientation",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupLobby(t)
				return c, p1
			},
			ship: "Carrier",
			x:    0, y: 0,
			orient:  "Diagonal",
			wantErr: controller.ErrInvalidOrientation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, pid := tt.setup()
			err := c.PlaceShip(pid, tt.ship, tt.x, tt.y, tt.orient)

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
			wantErr: controller.ErrFleetIncomplete,
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
		x, y       int
		wantResult string
		wantErr    error
	}{
		{
			name: "Valid Hit",
			setup: func() (*controller.Controller, string) {
				// P1 vs P2 (P2 has ship at 0,0)
				c, p1, _ := setupActiveGame(t)
				return c, p1
			},
			x: 0, y: 0,
			wantResult: controller.ResultHit,
			wantErr:    nil,
		},
		{
			name: "Valid Miss",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupActiveGame(t)
				return c, p1
			},
			x: 9, y: 9,
			wantResult: controller.ResultMiss,
			wantErr:    nil,
		},
		{
			name: "Rejects incorrect phase",
			setup: func() (*controller.Controller, string) {
				c := controller.NewController()
				p1, _ := c.Join()
				return c, p1
			},
			x: 0, y: 0,
			wantResult: controller.ResultInvalid,
			wantErr:    controller.ErrWrongGamePhase,
		},
		{
			name: "Rejects unknown player",
			setup: func() (*controller.Controller, string) {
				c, _, _ := setupActiveGame(t)
				return c, "random"
			},
			x: 0, y: 0,
			wantResult: controller.ResultInvalid,
			wantErr:    controller.ErrUnknownPlayer,
		},
		{
			name: "Enforces turn order (Not your turn)",
			setup: func() (*controller.Controller, string) {
				c, _, p2 := setupActiveGame(t)
				return c, p2 // P2 trying to fire during P1's turn
			},
			x: 0, y: 0,
			wantResult: controller.ResultInvalid,
			wantErr:    controller.ErrNotYourTurn,
		},
		{
			name: "Out of bounds error",
			setup: func() (*controller.Controller, string) {
				c, p1, _ := setupActiveGame(t)
				return c, p1
			},
			x: 100, y: 100,
			wantResult: controller.ResultInvalid,
			wantErr:    controller.ErrInvalidCoordinates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, pid := tt.setup()
			res, err := c.Fire(pid, tt.x, tt.y)

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
	_, err := c.Fire(p1, 0, 0)
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
			_, err := c.Fire(p1, x, row)
			if err != nil {
				t.Fatalf("Failed to fire during sink sequence at %d,%d: %v", x, row, err)
			}
			// Burn the defender's turn (shoot water at 9,9)
			_, _ = c.Fire(p2, 9, 9)
		}
	}

	// Sink first 4 ships
	sinkShip(0, 5) // Carrier
	sinkShip(1, 4) // Battleship
	sinkShip(2, 3) // Cruiser
	sinkShip(3, 3) // Submarine

	// Sink the final Destroyer (Size 2 at Row 4)

	// 1. Hit first part
	res, _ := c.Fire(p1, 0, 4)
	if res != controller.ResultHit {
		t.Errorf("Expected Hit, got %v", res)
	}
	_, _ = c.Fire(p2, 9, 9) // Burn turn

	// 2. Hit last part (Winning Shot)
	res, _ = c.Fire(p1, 1, 4)
	if res != controller.ResultSunk {
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
	_, err := c.Fire(p2, 5, 5)
	if err != controller.ErrGameOver {
		t.Fatalf("Expected ErrGameOver for post-game shot, got %v", err)
	}
}
