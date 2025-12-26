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
	user1, err := auth.LoginOrRegister(ctx, "Alice", "web", "Alice")
	require.NoError(t, err)
	assert.NotEmpty(t, user1.ID)
	assert.Equal(t, "Alice", user1.Username)

	// 2. Login existing user (same source/extID)
	user2, err := auth.LoginOrRegister(ctx, "AliceChanged", "web", "Alice")
	require.NoError(t, err)
	assert.Equal(t, user1.ID, user2.ID, "Should return same user ID")
	assert.Equal(t, "Alice", user2.Username, "Should return original username (no update logic implemented)")

	// 3. Register different user
	user3, err := auth.LoginOrRegister(ctx, "Bob", "discord", "12345")
	require.NoError(t, err)
	assert.NotEqual(t, user1.ID, user3.ID)
}
