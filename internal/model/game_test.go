package model_test

import (
	"testing"

	"github.com/callegarimattia/battleship/internal/dto"
	m "github.com/callegarimattia/battleship/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGame verifies initialization logic
func TestNewGame(t *testing.T) {
	t.Parallel()

	g := m.NewFullGame("P1", "P2", nil)

	err := g.PlaceShip("P1", m.Coordinate{X: 0, Y: 0}, 5, m.Horizontal)
	assert.NoError(t, err, "NewGame(nil) should load StandardFleet, but failed to place Carrier")

	miniFleet := map[int]int{2: 1} // Only one destroyer
	g2 := m.NewFullGame("P1", "P2", miniFleet)

	err = g2.PlaceShip("P1", m.Coordinate{X: 0, Y: 0}, 2, m.Horizontal)
	assert.NoError(t, err, "NewGame(custom) failed to place valid ship")

	err = g2.PlaceShip(
		"P1",
		m.Coordinate{X: 0, Y: 1},
		5,
		m.Horizontal,
	)
	assert.ErrorIs(t, err, m.ErrNoShipsRemaining, "NewGame(custom) allowed placing invalid ship size")
}

func TestJoin(t *testing.T) {
	t.Parallel()

	g := m.NewGame()

	// 1. Join first player
	err := g.Join("Alice", nil)
	require.NoError(t, err, "First player should join successfully")

	// 2. Join second player
	err = g.Join("Bob", nil)
	require.NoError(t, err, "Second player should join successfully")

	// 3. Check game state after both joined
	view, err := g.GetView("Alice")
	require.NoError(t, err)
	assert.Equal(t, dto.StateSetup, view.State, "Game should be in Setup state after valid join")

	// 4. Try to join third player
	err = g.Join("Charlie", nil)
	assert.ErrorIs(t, err, m.ErrGameFull, "Third player should not be able to join")
}

// TestPlaceShip_Rules verifies the constraints of placing ships
func TestPlaceShip_Rules(t *testing.T) {
	t.Parallel()

	miniFleet := map[int]int{3: 1}
	g := m.NewFullGame("Alice", "Bob", miniFleet)

	err := g.PlaceShip("Alice", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	assert.NoError(t, err, "Valid PlaceShip failed")

	err = g.PlaceShip("Alice", m.Coordinate{X: 0, Y: 1}, 3, m.Horizontal)
	assert.ErrorIs(t, err, m.ErrNoShipsRemaining, "Expected ErrNoShipsRemaining for duplicate ship")

	err = g.PlaceShip("Alice", m.Coordinate{X: 0, Y: 2}, 5, m.Horizontal)
	assert.ErrorIs(t, err, m.ErrNoShipsRemaining, "Expected ErrNoShipsRemaining for size 5")

	err = g.PlaceShip("Hacker", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	assert.ErrorIs(t, err, m.ErrUnknownPlayer, "Expected ErrUnknownPlayer")
}

// TestStartGame_Transitions verifies the state machine
func TestStartGame_Transitions(t *testing.T) {
	t.Parallel()

	miniFleet := map[int]int{3: 1}
	g := m.NewFullGame("P1", "P2", miniFleet)

	err := g.StartGame()
	assert.ErrorIs(t, err, m.ErrNotReadyToStart, "StartGame should fail on empty board")

	mustPlace(t, g, "P1", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	err = g.StartGame()
	assert.ErrorIs(t, err, m.ErrNotReadyToStart, "StartGame should fail if P2 is not ready")

	mustPlace(t, g, "P2", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)

	err = g.StartGame()
	require.NoError(t, err, "StartGame failed with valid setup")

	err = g.PlaceShip("P1", m.Coordinate{X: 5, Y: 5}, 3, m.Horizontal)
	assert.ErrorIs(t, err, m.ErrNotInSetup, "Expected ErrNotInSetup when placing during game")

	err = g.StartGame()
	assert.ErrorIs(t, err, m.ErrNotInSetup, "Expected ErrNotInSetup when starting already started game")
}

// TestAttack_TurnLogic verifies turn enforcement and switching
func TestAttack_TurnLogic(t *testing.T) {
	t.Parallel()

	g := m.NewFullGame("P1", "P2", map[int]int{3: 1})
	mustPlace(t, g, "P1", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	mustPlace(t, g, "P2", m.Coordinate{X: 0, Y: 0}, 3, m.Horizontal)
	_ = g.StartGame()

	// P1 should start
	_, err := g.Attack("P2", m.Coordinate{X: 0, Y: 0})
	assert.ErrorIs(t, err, m.ErrNotYourTurn, "Expected ErrNotYourTurn for P2")

	res := mustAttack(t, g, "P1", m.Coordinate{X: 5, Y: 5})
	assert.Equal(t, m.ShotResultMiss, res, "Expected Miss")

	_, err = g.Attack("P1", m.Coordinate{X: 0, Y: 0})
	assert.ErrorIs(t, err, m.ErrNotYourTurn, "Turn did not switch to P2 after attack")

	res = mustAttack(t, g, "P2", m.Coordinate{X: 0, Y: 0})
	assert.Equal(t, m.ShotResultHit, res, "Expected Hit")

	_, err = g.Attack("P2", m.Coordinate{X: 0, Y: 1})
	assert.ErrorIs(t, err, m.ErrNotYourTurn, "Turn did not switch back to P1 after Hit")
}

// TestAttack_GameEnd verifies winning condition
func TestAttack_GameEnd(t *testing.T) {
	t.Parallel()

	g := m.NewFullGame("Winner", "Loser", map[int]int{1: 1})

	mustPlace(t, g, "Winner", m.Coordinate{X: 0, Y: 0}, 1, m.Horizontal)
	mustPlace(t, g, "Loser", m.Coordinate{X: 0, Y: 0}, 1, m.Horizontal)

	_ = g.StartGame()

	res := mustAttack(t, g, "Winner", m.Coordinate{X: 0, Y: 0})
	assert.Equal(t, m.ShotResultSunk, res, "Expected Sunk")

	_, err := g.Attack("Loser", m.Coordinate{X: 0, Y: 0})
	assert.ErrorIs(t, err, m.ErrNotInPlay, "Expected ErrNotInPlay (Game Over)")

	assert.Equal(t, "Winner", g.Winner(), "Expected winner to be 'Winner'")
}

// TestAttack_InvalidInputs verifies defensive checks
func TestAttack_InvalidInputs(t *testing.T) {
	t.Parallel()

	g := m.NewFullGame("P1", "P2", map[int]int{1: 1})

	_, err := g.Attack("P1", m.Coordinate{X: 0, Y: 0})
	assert.ErrorIs(t, err, m.ErrNotInPlay, "Attack before start: want ErrNotInPlay")

	mustPlace(t, g, "P1", m.Coordinate{X: 0, Y: 0}, 1, m.Vertical)
	mustPlace(t, g, "P2", m.Coordinate{X: 0, Y: 0}, 1, m.Vertical)
	_ = g.StartGame()

	_, err = g.Attack("Ghost", m.Coordinate{X: 0, Y: 0})
	assert.ErrorIs(t, err, m.ErrUnknownPlayer, "Unknown player: want ErrUnknownPlayer")

	res, err := g.Attack("P1", m.Coordinate{X: 99, Y: 99})
	assert.ErrorIs(t, err, m.ErrInvalidShot, "Out of bounds: want ErrInvalidShot")
	assert.Equal(t, m.ShotResultInvalid, res, "Out of bounds: want ShotResultInvalid")
}

// Helper: Places a ship and fails test if error occurs
func mustPlace(
	t *testing.T,
	g *m.Game,
	playerID string,
	c m.Coordinate,
	size int,
	o m.Orientation,
) {
	t.Helper()
	err := g.PlaceShip(playerID, c, size, o)
	require.NoErrorf(t, err, "Setup failed: could not place ship for %s", playerID)
}

// Helper: Attacks and fails test if error occurs
func mustAttack(t *testing.T, g *m.Game, attackerID string, c m.Coordinate) m.ShotResult {
	t.Helper()
	res, err := g.Attack(attackerID, c)
	require.NoErrorf(t, err, "Attack failed")
	return res
}

func TestGame_GetView(t *testing.T) {
	t.Parallel()

	// Setup a game with 1x1 ships for simplicity
	g := m.NewFullGame("P1", "P2", map[int]int{1: 1})
	mustPlace(t, g, "P1", m.Coordinate{X: 0, Y: 0}, 1, m.Horizontal)
	mustPlace(t, g, "P2", m.Coordinate{X: 9, Y: 9}, 1, m.Horizontal)
	_ = g.StartGame()

	// P1 attacks P2 (Hit)
	mustAttack(t, g, "P1", m.Coordinate{X: 9, Y: 9})

	// P1 View: Should see own ship and opponent hit
	v1, err := g.GetView("P1")
	require.NoError(t, err, "GetView(P1) failed")

	// Board.Grid is accessed as [x][y] because of how internal/model/board.go typically structures it.
	// But let's check board.go first? Assuming standard [x][y] or [y][x].
	// The test used `v1.Me.Board[0][0]`. DTO is `Grid [][]CellState`.
	// Let's assume Grid[x][y] based on `internal/model/board.go` usually being map-like or array.
	assert.Equal(t, "SHIP", string(v1.Me.Board.Grid[0][0]), "P1 should see own ship at 0,0")
	assert.Equal(t, "SUNK", string(v1.Enemy.Board.Grid[9][9]), "P1 should see hit on P2 at 9,9")
	assert.Equal(t, "???", string(v1.Enemy.Board.Grid[0][0]), "P1 should see fog at P2's 0,0")

	// Spectator / Unknown user
	_, err = g.GetView("Ghost")
	assert.ErrorIs(t, err, m.ErrUnknownPlayer, "GetView(Ghost) should fail with ErrUnknownPlayer")
}
