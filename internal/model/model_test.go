package model_test

import (
	"errors"
	"testing"

	. "github.com/callegarimattia/battleship/internal/model"
)

func TestPlayer_PlaceShip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(*Player) // Optional setup per test
		ship        ShipType
		coord       Coordinate
		orient      Orientation
		wantErr     error
		wantDeplete bool // Check if inventory depleted
	}{
		{
			name:    "Valid placement",
			ship:    Carrier,
			coord:   Coordinate{X: 0, Y: 0},
			orient:  Horizontal,
			wantErr: nil,
		},
		{
			name:    "Out of bounds placement",
			ship:    Carrier,
			coord:   Coordinate{X: 8, Y: 0}, // X=8..12
			orient:  Horizontal,
			wantErr: ErrOutOfBounds,
		},
		{
			name: "Overlapping placement (Standard)",
			setup: func(p *Player) {
				_ = p.PlaceShip(Destroyer, Coordinate{X: 0, Y: 0}, Horizontal)
			},
			ship:    Destroyer,
			coord:   Coordinate{X: 0, Y: 0},
			orient:  Vertical,
			wantErr: ErrShipOverlap,
		},
		{
			name: "Inventory depleted",
			setup: func(p *Player) {
				_ = p.PlaceShip(Carrier, Coordinate{X: 0, Y: 0}, Horizontal)
			},
			ship:    Carrier,
			coord:   Coordinate{X: 0, Y: 1}, // Try placing second Carrier
			orient:  Horizontal,
			wantErr: ErrShipTypeDepleted,
		},
		// Collision Edge Cases
		{
			name: "Exact overlap",
			setup: func(p *Player) {
				_ = p.PlaceShip(Destroyer, Coordinate{X: 2, Y: 2}, Horizontal)
			},
			ship:    Submarine,
			coord:   Coordinate{X: 2, Y: 2},
			orient:  Horizontal,
			wantErr: ErrShipOverlap,
		},
		{
			name: "Crossing overlap",
			setup: func(p *Player) {
				_ = p.PlaceShip(Cruiser, Coordinate{X: 2, Y: 2}, Horizontal)
			},
			ship:    Submarine,
			coord:   Coordinate{X: 3, Y: 1},
			orient:  Vertical,
			wantErr: ErrShipOverlap,
		},
		{
			name: "Head-to-Tail Overlap",
			setup: func(p *Player) {
				_ = p.PlaceShip(Destroyer, Coordinate{X: 2, Y: 2}, Horizontal)
			},
			ship:    Cruiser,
			coord:   Coordinate{X: 3, Y: 2},
			orient:  Horizontal,
			wantErr: ErrShipOverlap,
		},
		{
			name: "Parallel Adjacent (Valid)",
			setup: func(p *Player) {
				_ = p.PlaceShip(Destroyer, Coordinate{X: 2, Y: 2}, Horizontal)
			},
			ship:    Submarine,
			coord:   Coordinate{X: 2, Y: 3},
			orient:  Horizontal,
			wantErr: nil,
		},
		{
			name: "Tip-to-Tip Adjacent (Valid)",
			setup: func(p *Player) {
				_ = p.PlaceShip(Destroyer, Coordinate{X: 2, Y: 2}, Horizontal)
			},
			ship:    Submarine,
			coord:   Coordinate{X: 4, Y: 2},
			orient:  Horizontal,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewPlayer()
			if tt.setup != nil {
				tt.setup(p)
			}

			err := p.PlaceShip(tt.ship, tt.coord, tt.orient)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("PlaceShip() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlayer_ReceiveAttack(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		setup      func(*Player)
		coord      Coordinate
		wantResult ShotResult
		wantErr    error
	}

	tests := []testCase{
		{
			name:       "Shooting out of bounds",
			coord:      Coordinate{X: 10, Y: 10},
			wantResult: ResultMiss,
			wantErr:    ErrOutOfBounds,
		},
		{
			name:       "Shooting water",
			coord:      Coordinate{X: 5, Y: 5},
			wantResult: ResultMiss,
			wantErr:    nil,
		},
		{
			name: "Shooting a ship (Hit)",
			setup: func(p *Player) {
				_ = p.PlaceShip(Destroyer, Coordinate{X: 0, Y: 0}, Horizontal)
			},
			coord:      Coordinate{X: 0, Y: 0},
			wantResult: ResultHit,
			wantErr:    nil,
		},
		{
			name: "Repeated miss",
			setup: func(p *Player) {
				_, _ = p.ReceiveAttack(Coordinate{X: 0, Y: 0})
			},
			coord:      Coordinate{X: 0, Y: 0},
			wantResult: ResultMiss,
			wantErr:    nil,
		},
		{
			name: "Sinking a ship",
			setup: func(p *Player) {
				_ = p.PlaceShip(Destroyer, Coordinate{X: 0, Y: 0}, Horizontal) // Target
				_ = p.PlaceShip(
					Battleship,
					Coordinate{X: 0, Y: 2},
					Horizontal,
				) // Dummy to keep game alive
				_, _ = p.ReceiveAttack(Coordinate{X: 0, Y: 0}) // Hit 1
			},
			coord:      Coordinate{X: 1, Y: 0}, // Hit 2
			wantResult: ResultSunk,
			wantErr:    nil,
		},
		{
			name: "Repeated hit on ship (Should error or not increment)",
			setup: func(p *Player) {
				_ = p.PlaceShip(Destroyer, Coordinate{X: 0, Y: 0}, Horizontal)
				_, _ = p.ReceiveAttack(Coordinate{X: 0, Y: 0}) // Hit 1
			},
			coord:      Coordinate{X: 0, Y: 0}, // Hit 2 (Same spot)
			wantResult: ResultInvalid,          // Expecting error/invalid based on user request
			wantErr:    ErrRepeatedHit,         // Correct error type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewPlayer()
			if tt.setup != nil {
				tt.setup(p)
			}

			res, err := p.ReceiveAttack(tt.coord)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ReceiveAttack() error = %v, wantErr %v", err, tt.wantErr)
			}
			if res != tt.wantResult {
				t.Errorf("ReceiveAttack() result = %v, wantResult %v", res, tt.wantResult)
			}
		})
	}
}

func TestPlayer_SetReady(t *testing.T) {
	t.Parallel()

	p := NewPlayer()

	// 1. Try setting ready before placing ships -> Error
	if err := p.SetReady(); !errors.Is(err, ErrFleetIncomplete) {
		t.Errorf("Expected ErrFleetIncomplete, got %v", err)
	}

	if p.IsReady() {
		t.Error("Player should not be ready yet")
	}

	// 2. Place all ships
	placements := []struct {
		sType  ShipType
		coord  Coordinate
		orient Orientation
	}{
		{Carrier, Coordinate{X: 0, Y: 0}, Horizontal},
		{Battleship, Coordinate{X: 0, Y: 1}, Horizontal},
		{Cruiser, Coordinate{X: 0, Y: 2}, Horizontal},
		{Submarine, Coordinate{X: 0, Y: 3}, Horizontal},
		{Destroyer, Coordinate{X: 0, Y: 4}, Horizontal},
	}

	for _, pl := range placements {
		err := p.PlaceShip(pl.sType, pl.coord, pl.orient)
		if err != nil {
			t.Fatalf("Failed to place %s: %v", pl.sType, err)
		}
	}

	// 3. Set ready -> Success
	if err := p.SetReady(); err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if !p.IsReady() {
		t.Error("Player should be ready")
	}
}

func TestPlayer_Helpers(t *testing.T) {
	t.Run("ID and String", func(t *testing.T) {
		t.Parallel()

		p := NewPlayer()
		if p.ID() == "" {
			t.Error("Player ID should not be empty")
		}
		expectedStr := "Player(" + p.ID() + ")"
		if p.String() != expectedStr {
			t.Errorf("Expected String() to be %q, got %q", expectedStr, p.String())
		}
	})

	t.Run("HasLost", func(t *testing.T) {
		t.Parallel()

		p := NewPlayer()
		// New player has 0 ships afloat -> HasLost = true
		if !p.HasLost() {
			t.Error("New player should have lost (0 ships afloat)")
		}

		_ = p.PlaceShip(Carrier, Coordinate{X: 0, Y: 0}, Horizontal)
		if p.HasLost() {
			t.Error("Player with ships should not have lost")
		}
	})

	t.Run("MarksShotResult", func(t *testing.T) {
		t.Parallel()

		p := NewPlayer()

		// Out of bounds
		if err := p.MarksShotResult(Coordinate{X: -1, Y: 0}, ResultMiss); !errors.Is(
			err,
			ErrOutOfBounds,
		) {
			t.Errorf("Expected ErrOutOfBounds, got %v", err)
		}

		// Invalid result
		if err := p.MarksShotResult(Coordinate{X: 0, Y: 0}, ResultInvalid); !errors.Is(
			err,
			ErrInvalidShotResult,
		) {
			t.Errorf("Expected ErrInvalidShotResult, got %v", err)
		}

		// Valid result
		if err := p.MarksShotResult(Coordinate{X: 0, Y: 0}, ResultHit); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("ShipType Size", func(t *testing.T) {
		t.Parallel()

		invalid := ShipType("Invalid")
		if size := invalid.Size(); size != 0 {
			t.Errorf("Expected size 0 for invalid ship type, got %d", size)
		}
	})
}
