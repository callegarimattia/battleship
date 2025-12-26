package service

import (
	"context"
	"testing"
	"time"

	"github.com/callegarimattia/battleship/internal/events"
	mocks_events "github.com/callegarimattia/battleship/internal/mocks/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	hostID  = "host-123"
	guestID = "guest-456"
)

func TestMemoryService_JoinMatch_EmitsEvent(t *testing.T) {
	t.Parallel()

	mockBus := mocks_events.NewMockEventBus(t)
	svc := NewMemoryService(mockBus)
	ctx := context.Background()

	matchID, err := svc.CreateMatch(ctx, hostID)
	require.NoError(t, err)

	// Expect event to be published when guest joins
	mockBus.EXPECT().Publish(mock.MatchedBy(func(event *events.GameEvent) bool {
		return event.Type == events.EventPlayerJoined &&
			event.MatchID == matchID &&
			event.PlayerID == guestID &&
			event.TargetID == hostID
	})).Once()

	// Join the match
	_, err = svc.JoinMatch(ctx, matchID, guestID)
	require.NoError(t, err)

	// Verify expectations
	mockBus.AssertExpectations(t)
}

func TestMemoryService_PlaceShip_EmitsEvent(t *testing.T) {
	t.Parallel()

	mockBus := mocks_events.NewMockEventBus(t)
	svc := NewMemoryService(mockBus)
	ctx := context.Background()

	// Ignore setup events
	mockBus.EXPECT().Publish(mock.MatchedBy(func(event *events.GameEvent) bool {
		return event.Type != events.EventShipPlaced
	})).Return()

	// Setup: Create match and join
	matchID, err := svc.CreateMatch(ctx, hostID)
	require.NoError(t, err)
	_, err = svc.JoinMatch(ctx, matchID, guestID)
	require.NoError(t, err)

	// Expect event when ship is placed
	size := 5
	x, y := 0, 0
	vertical := true
	mockBus.EXPECT().Publish(mock.MatchedBy(func(event *events.GameEvent) bool {
		if event.Type != events.EventShipPlaced {
			return false
		}
		if event.MatchID != matchID || event.PlayerID != hostID || event.TargetID != guestID {
			return false
		}
		data, ok := event.Data.(events.ShipPlacedEventData)
		return ok && data.Size == size && data.X == x && data.Y == y && data.Vertical == vertical
	})).Once()

	// Place a ship
	_, err = svc.PlaceShip(ctx, matchID, hostID, size, x, y, vertical)
	require.NoError(t, err)

	mockBus.AssertExpectations(t)
}

func TestMemoryService_Attack_EmitsEvent(t *testing.T) {
	t.Parallel()

	mockBus := mocks_events.NewMockEventBus(t)
	svc := NewMemoryService(mockBus)
	ctx := context.Background()

	// Ignore setup events
	mockBus.EXPECT().Publish(mock.MatchedBy(func(event *events.GameEvent) bool {
		return event.Type != events.EventAttackMade
	})).Return()

	// Setup: Create match, join, and place all ships
	matchID, err := svc.CreateMatch(ctx, hostID)
	require.NoError(t, err)
	_, err = svc.JoinMatch(ctx, matchID, guestID)
	require.NoError(t, err)

	// Place ships for both players
	ships := []struct{ size, x, y int }{
		{5, 0, 0}, {4, 1, 0}, {3, 2, 0}, {3, 3, 0}, {2, 4, 0},
	}
	for _, ship := range ships {
		_, err = svc.PlaceShip(ctx, matchID, hostID, ship.size, ship.x, ship.y, true)
		require.NoError(t, err)
		_, err = svc.PlaceShip(ctx, matchID, guestID, ship.size, ship.x, ship.y, true)
		require.NoError(t, err)
	}

	// Expect event when attack is made
	attackX, attackY := 5, 5
	mockBus.EXPECT().Publish(mock.MatchedBy(func(event *events.GameEvent) bool {
		if event.Type != events.EventAttackMade {
			return false
		}
		if event.MatchID != matchID || event.PlayerID != hostID || event.TargetID != guestID {
			return false
		}
		data, ok := event.Data.(events.AttackEventData)
		return ok && data.X == attackX && data.Y == attackY && data.Result == "miss"
	})).Once()

	// Attack
	_, err = svc.Attack(ctx, matchID, hostID, attackX, attackY)
	require.NoError(t, err)

	mockBus.AssertExpectations(t)
}

func TestMemoryService_Attack_HitEvent(t *testing.T) {
	t.Parallel()

	mockBus := mocks_events.NewMockEventBus(t)
	svc := NewMemoryService(mockBus)
	ctx := context.Background()

	// Ignore setup events
	mockBus.EXPECT().Publish(mock.MatchedBy(func(event *events.GameEvent) bool {
		return event.Type != events.EventAttackMade
	})).Return()

	// Setup: Create match, join, and place ships
	matchID, err := svc.CreateMatch(ctx, hostID)
	require.NoError(t, err)
	_, err = svc.JoinMatch(ctx, matchID, guestID)
	require.NoError(t, err)

	// Place ships
	ships := []struct{ size, x, y int }{
		{5, 0, 0}, {4, 1, 0}, {3, 2, 0}, {3, 3, 0}, {2, 4, 0},
	}
	for _, ship := range ships {
		_, err = svc.PlaceShip(ctx, matchID, hostID, ship.size, ship.x, ship.y, true)
		require.NoError(t, err)
		_, err = svc.PlaceShip(ctx, matchID, guestID, ship.size, ship.x, ship.y, true)
		require.NoError(t, err)
	}

	// Expect hit event
	mockBus.EXPECT().Publish(mock.MatchedBy(func(event *events.GameEvent) bool {
		if event.Type != events.EventAttackMade {
			return false
		}
		data, ok := event.Data.(events.AttackEventData)
		return ok && data.Result == "hit"
	})).Once()

	// Attack a position where guest has a ship (0, 0)
	_, err = svc.Attack(ctx, matchID, hostID, 0, 0)
	require.NoError(t, err)

	mockBus.AssertExpectations(t)
}

func TestMemoryService_NoEventBus_DoesNotPanic(t *testing.T) {
	t.Parallel()

	// Create service with nil event bus
	svc := NewMemoryService(nil)
	ctx := context.Background()

	// Should not panic
	matchID, err := svc.CreateMatch(ctx, hostID)
	require.NoError(t, err)

	_, err = svc.JoinMatch(ctx, matchID, guestID)
	require.NoError(t, err)

	_, err = svc.PlaceShip(ctx, matchID, hostID, 5, 0, 0, true)
	require.NoError(t, err)
}

func TestMemoryService_EventTimestamp(t *testing.T) {
	t.Parallel()

	mockBus := mocks_events.NewMockEventBus(t)
	svc := NewMemoryService(mockBus)
	ctx := context.Background()

	matchID, err := svc.CreateMatch(ctx, hostID)
	require.NoError(t, err)

	// Capture the event to verify timestamp
	var capturedEvent *events.GameEvent
	mockBus.EXPECT().
		Publish(mock.AnythingOfType("*events.GameEvent")).
		Run(func(event *events.GameEvent) {
			capturedEvent = event
		}).
		Once()

	_, err = svc.JoinMatch(ctx, matchID, guestID)
	require.NoError(t, err)

	require.NotNil(t, capturedEvent)
	assert.WithinDuration(t, time.Now(), capturedEvent.Timestamp, 2*time.Second)
}
