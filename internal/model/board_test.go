package model_test

import (
	"testing"

	m "github.com/callegarimattia/battleship/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helper for creating ships without error checks in tests ---
func mustNewShip(t *testing.T, size int) *m.Ship {
	t.Helper()
	s, err := m.NewShip(size)
	require.NoErrorf(t, err, "failed to create ship of size %d", size)
	return s
}

func TestNewShip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		size    int
		wantErr error
	}{
		{"Valid size 1", 1, nil},
		{"Valid size 5", 5, nil},
		{"Invalid size 0", 0, m.ErrInvalidShipSize},
		{"Invalid size negative", -1, m.ErrInvalidShipSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := m.NewShip(tt.size)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got, "NewShip() expected nil ship on error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.size, got.Size())
			}
		})
	}
}

func TestPlaceShip(t *testing.T) {
	t.Parallel()

	ship2 := mustNewShip(t, 2)
	ship3 := mustNewShip(t, 3)

	tests := []struct {
		name        string
		setup       func(*m.Board)
		coord       m.Coordinate
		ship        *m.Ship
		orientation m.Orientation
		wantErr     error
	}{
		{
			name:        "Valid Horizontal",
			coord:       m.Coordinate{X: 0, Y: 0},
			ship:        ship3,
			orientation: m.Horizontal,
			wantErr:     nil,
		},
		{
			name:        "Valid Vertical",
			coord:       m.Coordinate{X: 5, Y: 5},
			ship:        ship3,
			orientation: m.Vertical,
			wantErr:     nil,
		},
		{
			name:        "Out of Bounds - Start X",
			coord:       m.Coordinate{X: -1, Y: 0},
			ship:        ship2,
			orientation: m.Horizontal,
			wantErr:     m.ErrShipOutOfBounds,
		},
		{
			name:        "Out of Bounds - Start Y",
			coord:       m.Coordinate{X: 0, Y: 10},
			ship:        ship2,
			orientation: m.Vertical,
			wantErr:     m.ErrShipOutOfBounds,
		},
		{
			name:        "Out of Bounds - End Extends X",
			coord:       m.Coordinate{X: 9, Y: 0},
			ship:        ship2, // Size 2 needs X=9, X=10(invalid)
			orientation: m.Horizontal,
			wantErr:     m.ErrShipOutOfBounds,
		},
		{
			name:        "Out of Bounds - End Extends Y",
			coord:       m.Coordinate{X: 0, Y: 8},
			ship:        ship3, // Size 3 needs Y=8, Y=9, Y=10(invalid)
			orientation: m.Vertical,
			wantErr:     m.ErrShipOutOfBounds,
		},
		{
			name: "Overlap Collision",
			setup: func(b *m.Board) {
				// Place a ship at 2,2 (Vertical size 3 -> 2,2; 2,3; 2,4)
				_ = b.PlaceShip(m.Coordinate{X: 2, Y: 2}, ship3, m.Vertical)
			},
			coord:       m.Coordinate{X: 1, Y: 3},
			ship:        ship3, // Horizontal size 3 -> 1,3; 2,3(COLLISION); 3,3
			orientation: m.Horizontal,
			wantErr:     m.ErrShipOverlap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Fresh board for each test case
			b := m.NewBoard()
			if tt.setup != nil {
				tt.setup(b)
			}

			err := b.PlaceShip(tt.coord, tt.ship, tt.orientation)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReceiveShot(t *testing.T) {
	t.Parallel()

	// Setup: Create a board with one ship
	// Ship is at (0,0) and (1,0) [Horizontal size 2]
	b := m.NewBoard()
	ship := mustNewShip(t, 2)
	err := b.PlaceShip(m.Coordinate{X: 0, Y: 0}, ship, m.Horizontal)
	require.NoError(t, err, "Setup failed")

	tests := []struct {
		name       string
		coord      m.Coordinate
		wantResult m.ShotResult
	}{
		{
			name:       "Shot Out of Bounds Negative",
			coord:      m.Coordinate{X: -1, Y: 0},
			wantResult: m.ShotResultInvalid,
		},
		{
			name:       "Shot Out of Bounds Large",
			coord:      m.Coordinate{X: 10, Y: 10},
			wantResult: m.ShotResultInvalid,
		},
		{
			name:       "Miss Empty Water",
			coord:      m.Coordinate{X: 5, Y: 5},
			wantResult: m.ShotResultMiss,
		},
		{
			name:       "Hit First Segment",
			coord:      m.Coordinate{X: 0, Y: 0},
			wantResult: m.ShotResultHit,
		},
		{
			name:       "Duplicate Shot on Hit",
			coord:      m.Coordinate{X: 0, Y: 0}, // Same spot
			wantResult: m.ShotResultInvalid,
		},
		{
			name:       "Sunk Second Segment",
			coord:      m.Coordinate{X: 1, Y: 0},
			wantResult: m.ShotResultSunk,
		},
		{
			name:       "Duplicate Shot on Sunk Ship",
			coord:      m.Coordinate{X: 1, Y: 0},
			wantResult: m.ShotResultInvalid,
		},
	}

	for _, tt := range tests {
		got := b.ReceiveShot(tt.coord)
		assert.Equal(
			t,
			tt.wantResult,
			got,
			"ReceiveShot(%v) = %v, want %v",
			tt.coord,
			got,
			tt.wantResult,
		)
	}
}

func TestAllShipsSunk(t *testing.T) {
	t.Parallel()

	b := m.NewBoard()

	// Scenario 1: Empty board should count as "All Sunk"
	assert.True(t, b.AllShipsSunk(), "New/Empty board should return true for AllShipsSunk")

	// Scenario 2: Add ships
	s1 := mustNewShip(t, 1) // At 0,0
	s2 := mustNewShip(t, 2) // At 5,5 -> 5,6

	_ = b.PlaceShip(m.Coordinate{X: 0, Y: 0}, s1, m.Horizontal)
	_ = b.PlaceShip(m.Coordinate{X: 5, Y: 5}, s2, m.Vertical)

	assert.False(t, b.AllShipsSunk(), "Board with healthy ships should NOT be sunk")

	// Scenario 3: Sink first ship
	b.ReceiveShot(m.Coordinate{X: 0, Y: 0})
	assert.False(t, b.AllShipsSunk(), "Board with one remaining ship should NOT be sunk")

	// Scenario 4: Damage second ship
	b.ReceiveShot(m.Coordinate{X: 5, Y: 5})
	assert.False(t, b.AllShipsSunk(), "Board with partially damaged ship should NOT be sunk")

	// Scenario 5: Sink last segment
	b.ReceiveShot(m.Coordinate{X: 5, Y: 6})

	assert.True(t, b.AllShipsSunk(), "All ships are destroyed, should return true")
}
