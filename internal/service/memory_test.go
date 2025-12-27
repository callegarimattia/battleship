package service_test

import (
	"context"
	"testing"

	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/callegarimattia/battleship/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryService_LobbyFlow(t *testing.T) {
	t.Parallel()
	s := service.NewMemoryService(service.NewNotificationService())
	ctx := context.Background()

	matchID, err := s.CreateMatch(ctx, "host-1")
	require.NoError(t, err)
	assert.NotEmpty(t, matchID)

	matches, err := s.ListMatches(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, matches)
	found := false
	for _, m := range matches {
		if m.ID == matchID {
			found = true
			assert.Equal(t, "host-1", m.HostName)
			assert.Equal(t, 1, m.PlayerCount)
		}
	}
	assert.True(t, found, "Match ID should be in the list")

	view, err := s.JoinMatch(ctx, matchID, "guest-1")
	require.NoError(t, err)
	assert.Equal(t, dto.StateSetup, view.State)
	assert.Equal(t, "guest-1", view.Me.ID)

	matches, _ = s.ListMatches(ctx)
	for _, m := range matches {
		if m.ID == matchID {
			assert.Equal(t, 2, m.PlayerCount)
		}
	}
}

func TestMemoryService_JoinErrors(t *testing.T) {
	t.Parallel()
	s := service.NewMemoryService(service.NewNotificationService())
	ctx := context.Background()

	_, err := s.JoinMatch(ctx, "non-existent", "p1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "match not found")
}

func TestMemoryService_GameplayFlow(t *testing.T) {
	t.Parallel()
	s := service.NewMemoryService(service.NewNotificationService())
	ctx := context.Background()

	matchID, _ := s.CreateMatch(ctx, "p1")
	_, _ = s.JoinMatch(ctx, matchID, "p2")

	view, err := s.PlaceShip(ctx, matchID, "p1", 3, 0, 0, true)
	require.NoError(t, err)
	assert.Equal(t, dto.StateSetup, view.State)

	view, err = s.PlaceShip(ctx, matchID, "p2", 3, 0, 0, true)
	require.NoError(t, err)

	state, err := s.GetState(ctx, matchID, "p1")
	require.NoError(t, err)
	assert.Equal(t, dto.StateSetup, state.State)
}

func TestMemoryService_Attack_NotStarted(t *testing.T) {
	t.Parallel()
	s := service.NewMemoryService(service.NewNotificationService())
	ctx := context.Background()

	matchID, _ := s.CreateMatch(ctx, "p1")
	_, err := s.Attack(ctx, matchID, "p1", 0, 0)
	assert.Error(t, err) // Game not started
}

func TestMemoryService_SingleActiveGameLimit(t *testing.T) {
	t.Parallel()
	s := service.NewMemoryService(service.NewNotificationService())
	ctx := context.Background()

	// Create first game
	game1, err := s.CreateMatch(ctx, "alice")
	require.NoError(t, err, "should create first game")
	require.NotEmpty(t, game1)

	// Try to create second game while first is active - should fail
	_, err = s.CreateMatch(ctx, "alice")
	require.Error(t, err, "should not allow creating second game")
	require.Contains(t, err.Error(), "already in an active game")

	// Try to join another game while in first game - should fail
	game2, err := s.CreateMatch(ctx, "bob")
	require.NoError(t, err)

	_, err = s.JoinMatch(ctx, game2, "alice")
	require.Error(t, err, "should not allow joining another game")
	require.Contains(t, err.Error(), "already in an active game")
}
