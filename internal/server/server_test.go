package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_CreateGame(t *testing.T) {
	t.Parallel()

	srv := New()
	id, err := srv.CreateGame()
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// Verify we can retrieve info for the created game
	info, err := srv.Info(id)
	require.NoError(t, err)
	assert.Equal(t, id, info.ID)
}

func TestServer_GetGame_NotFound(t *testing.T) {
	t.Parallel()

	srv := New()
	_, err := srv.Info("non-existent-id")
	assert.ErrorIs(t, err, ErrGameNotFound)
}

func TestServer_Join(t *testing.T) {
	t.Parallel()

	srv := New()
	gameID, err := srv.CreateGame()
	require.NoError(t, err)

	// Join the game
	playerID, err := srv.Join(gameID)
	require.NoError(t, err)
	assert.NotEmpty(t, playerID)

	// Verify player is in the game info
	info, err := srv.Info(gameID)
	require.NoError(t, err)
	assert.Contains(t, info.PlayerIDs, playerID)
}

func TestServer_MultipleGames(t *testing.T) {
	t.Parallel()

	srv := New()

	// Create two games
	gameID1, err := srv.CreateGame()
	require.NoError(t, err)

	gameID2, err := srv.CreateGame()
	require.NoError(t, err)

	assert.NotEqual(t, gameID1, gameID2)

	// Join game 1
	p1, err := srv.Join(gameID1)
	require.NoError(t, err)

	// Verify game 1 has player, game 2 does not
	info1, _ := srv.Info(gameID1)
	info2, _ := srv.Info(gameID2)

	assert.Len(t, info1.PlayerIDs, 1)
	assert.Equal(t, p1, info1.PlayerIDs[0])
	assert.Len(t, info2.PlayerIDs, 0)
}

func TestServer_Delegation(t *testing.T) {
	t.Parallel()

	// A simple test to ensure methods are wired up
	// We count on controller tests for logic, just checking routing here.
	srv := New()
	gameID, _ := srv.CreateGame()
	p1, _ := srv.Join(gameID)

	// Try to ready up (should fail as no ships placed, but proves method is called)
	err := srv.Ready(gameID, p1)
	// Controller returns ErrFleetIncomplete or similar if we haven't placed ships
	// The exact error doesn't matter as much as *some* controller error returning vs a generic panic/nil
	assert.Error(t, err)
	assert.NotEqual(t, ErrGameNotFound, err)

	// Try with bad game ID
	err = srv.Ready("bad-id", p1)
	assert.ErrorIs(t, err, ErrGameNotFound)

	// Try Fire
	_, err = srv.Fire(gameID, p1, 0, 0)
	assert.Error(t, err)

	// PlaceShip
	err = srv.PlaceShip(gameID, p1, "Carrier", 0, 0, "H")
	// Might fail due to phase/turn/etc, but routing should work
	assert.NotEqual(t, ErrGameNotFound, err)
}
