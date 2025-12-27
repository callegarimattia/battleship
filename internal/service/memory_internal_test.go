package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryService_Cleanup(t *testing.T) {
	t.Parallel()

	s := NewMemoryService(NewNotificationService())
	ctx := context.Background()

	activeID, err := s.CreateMatch(ctx, "host")
	require.NoError(t, err)

	staleID, mlErr := s.CreateMatch(ctx, "stale")
	require.NoError(t, mlErr)

	s.gamesMu.Lock()
	s.games[staleID].updatedAt = time.Now().Add(-25 * time.Hour)
	s.gamesMu.Unlock()

	s.gc()

	s.gamesMu.RLock()
	_, activeExists := s.games[activeID]
	_, staleExists := s.games[staleID]
	s.gamesMu.RUnlock()

	assert.True(t, activeExists, "Active game should exist")
	assert.False(t, staleExists, "Stale game should be removed")
}
