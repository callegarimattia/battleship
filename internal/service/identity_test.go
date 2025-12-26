package service_test

import (
	"context"
	"testing"

	"github.com/callegarimattia/battleship/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryIdentityService_LoginOrRegister(t *testing.T) {
	t.Parallel()
	auth := service.NewIdentityService()
	ctx := context.Background()

	// 1. Register new user
	resp1, err := auth.LoginOrRegister(ctx, "Alice", "web", "Alice")
	require.NoError(t, err)
	assert.NotEmpty(t, resp1.User.ID)
	assert.Equal(t, "Alice", resp1.User.Username)

	// 2. Login existing user (same source/extID)
	resp2, err := auth.LoginOrRegister(ctx, "AliceChanged", "web", "Alice")
	require.NoError(t, err)
	assert.Equal(t, resp1.User.ID, resp2.User.ID, "Should return same user ID")
	assert.Equal(
		t,
		"Alice",
		resp2.User.Username,
		"Should return original username (no update logic implemented)",
	)

	// 3. Register different user
	resp3, err := auth.LoginOrRegister(ctx, "Bob", "discord", "12345")
	require.NoError(t, err)
	assert.NotEqual(t, resp1.User.ID, resp3.User.ID)
}
