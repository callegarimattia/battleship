package model_test

import (
	"errors"
	"testing"

	m "github.com/callegarimattia/battleship/internal/model"
)

// TestNewGame verifies initialization logic
func TestNewGame(t *testing.T) {
	t.Parallel()

	g := m.NewGame("P1", "P2", nil)

	if err := g.PlaceShip("P1", m.Coordinate{X: 0, Y: 0}, 5, m.Horizontal); err != nil {
		t.Errorf("NewGame(nil) should load StandardFleet, but failed to place Carrier: %v", err)
	}

	miniFleet := map[int]int{2: 1} // Only one destroyer
	g2 := m.NewGame("P1", "P2", miniFleet)

	if err := g2.PlaceShip("P1", m.Coordinate{X: 0, Y: 0}, 2, m.Horizontal); err != nil {
		t.Errorf("NewGame(custom) failed to place valid ship: %v", err)
	}
	if err := g2.PlaceShip("P1", m.Coordinate{X: 0, Y: 1}, 5, m.Horizontal); !errors.Is(err, m.ErrNoShipsRemaining) {
		t.Errorf("NewGame(custom) allowed placing invalid ship size, want ErrNoShipsRemaining, got %v", err)
	}
}

// TestPlaceShip_Rules verifies the constraints of placing ships
func TestPlaceShip_Rules(t *testing.T) {
	t.Parallel()

	miniFleet := map[int]int{3: 1}
	g := m.NewGame("Alice", "Bob", miniFleet)

	err := g.PlaceShip("Alice", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	if err != nil {
		t.Errorf("Valid PlaceShip failed: %v", err)
	}

	err = g.PlaceShip("Alice", m.Coordinate{X: 0, Y: 1}, 3, m.Horizontal)
	if !errors.Is(err, m.ErrNoShipsRemaining) {
		t.Errorf("Expected ErrNoShipsRemaining for duplicate ship, got %v", err)
	}

	err = g.PlaceShip("Alice", m.Coordinate{X: 0, Y: 2}, 5, m.Horizontal)
	if !errors.Is(err, m.ErrNoShipsRemaining) {
		t.Errorf("Expected ErrNoShipsRemaining for size 5, got %v", err)
	}

	err = g.PlaceShip("Hacker", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	if !errors.Is(err, m.ErrUnknownPlayer) {
		t.Errorf("Expected ErrUnknownPlayer, got %v", err)
	}
}

// TestStartGame_Transitions verifies the state machine
func TestStartGame_Transitions(t *testing.T) {
	t.Parallel()

	miniFleet := map[int]int{3: 1}
	g := m.NewGame("P1", "P2", miniFleet)

	if err := g.StartGame(); !errors.Is(err, m.ErrNotReadyToStart) {
		t.Errorf("StartGame should fail on empty board, got %v", err)
	}

	mustPlace(t, g, "P1", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	if err := g.StartGame(); !errors.Is(err, m.ErrNotReadyToStart) {
		t.Errorf("StartGame should fail if P2 is not ready, got %v", err)
	}

	mustPlace(t, g, "P2", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)

	if err := g.StartGame(); err != nil {
		t.Fatalf("StartGame failed with valid setup: %v", err)
	}

	err := g.PlaceShip("P1", m.Coordinate{X: 5, Y: 5}, 3, m.Horizontal)
	if !errors.Is(err, m.ErrNotInSetup) {
		t.Errorf("Expected ErrNotInSetup when placing during game, got %v", err)
	}

	if err := g.StartGame(); !errors.Is(err, m.ErrNotInSetup) {
		t.Errorf("Expected ErrNotInSetup when starting already started game, got %v", err)
	}
}

// TestAttack_TurnLogic verifies turn enforcement and switching
func TestAttack_TurnLogic(t *testing.T) {
	t.Parallel()

	g := m.NewGame("P1", "P2", map[int]int{3: 1})
	mustPlace(t, g, "P1", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	mustPlace(t, g, "P2", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	_ = g.StartGame()

	// P1 should start
	_, err := g.Attack("P2", m.Coordinate{X: 0, Y: 0})
	if !errors.Is(err, m.ErrNotYourTurn) {
		t.Errorf("Expected ErrNotYourTurn for P2, got %v", err)
	}

	res := mustAttack(t, g, "P1", m.Coordinate{X: 5, Y: 5})
	if res != m.ShotResultMiss {
		t.Errorf("Expected Miss, got %s", res)
	}

	if _, err := g.Attack("P1", m.Coordinate{X: 0, Y: 0}); !errors.Is(err, m.ErrNotYourTurn) {
		t.Error("Turn did not switch to P2 after attack")
	}

	res = mustAttack(t, g, "P2", m.Coordinate{X: 0, Y: 0})
	if res != m.ShotResultHit {
		t.Errorf("Expected Hit, got %s", res)
	}

	if _, err := g.Attack("P2", m.Coordinate{X: 0, Y: 1}); !errors.Is(err, m.ErrNotYourTurn) {
		t.Error("Turn did not switch back to P1 after Hit")
	}
}

// TestAttack_GameEnd verifies winning condition
func TestAttack_GameEnd(t *testing.T) {
	t.Parallel()

	g := m.NewGame("Winner", "Loser", map[int]int{1: 1})

	mustPlace(t, g, "Winner", m.Coordinate{X: 0, Y: 0}, 1, m.Horizontal)
	mustPlace(t, g, "Loser", m.Coordinate{X: 0, Y: 0}, 1, m.Horizontal)

	_ = g.StartGame()

	res := mustAttack(t, g, "Winner", m.Coordinate{X: 0, Y: 0})
	if res != m.ShotResultSunk {
		t.Errorf("Expected Sunk, got %s", res)
	}

	_, err := g.Attack("Loser", m.Coordinate{X: 0, Y: 0})
	if !errors.Is(err, m.ErrNotInPlay) {
		t.Errorf("Expected ErrNotInPlay (Game Over), got %v", err)
	}

	if winnerID := g.Winner(); winnerID != "Winner" {
		t.Errorf("Expected winner to be 'Winner', got %s", winnerID)
	}
}

// TestAttack_InvalidInputs verifies defensive checks
func TestAttack_InvalidInputs(t *testing.T) {
	t.Parallel()

	g := m.NewGame("P1", "P2", map[int]int{1: 1})

	_, err := g.Attack("P1", m.Coordinate{X: 0, Y: 0})
	if !errors.Is(err, m.ErrNotInPlay) {
		t.Errorf("Attack before start: want ErrNotInPlay, got %v", err)
	}

	mustPlace(t, g, "P1", m.Coordinate{X: 0, Y: 0}, 1, m.Vertical)
	mustPlace(t, g, "P2", m.Coordinate{X: 0, Y: 0}, 1, m.Vertical)
	_ = g.StartGame()

	_, err = g.Attack("Ghost", m.Coordinate{X: 0, Y: 0})
	if !errors.Is(err, m.ErrUnknownPlayer) {
		t.Errorf("Unknown player: want ErrUnknownPlayer, got %v", err)
	}

	res, err := g.Attack("P1", m.Coordinate{X: 99, Y: 99})
	if !errors.Is(err, m.ErrInvalidShot) {
		t.Errorf("Out of bounds: want ErrInvalidShot, got %v", err)
	}
	if res != m.ShotResultInvalid {
		t.Errorf("Out of bounds: want ShotResultInvalid, got %s", res)
	}
}

// Helper: Places a ship and fails test if error occurs
func mustPlace(t *testing.T, g *m.Game, playerID string, c m.Coordinate, size int, o m.Orientation) {
	t.Helper()
	if err := g.PlaceShip(playerID, c, size, o); err != nil {
		t.Fatalf("Setup failed: could not place ship for %s: %v", playerID, err)
	}
}

// Helper: Attacks and fails test if error occurs
func mustAttack(t *testing.T, g *m.Game, attackerID string, c m.Coordinate) m.ShotResult {
	t.Helper()
	res, err := g.Attack(attackerID, c)
	if err != nil {
		t.Fatalf("Attack failed: %v", err)
	}
	return res
}
