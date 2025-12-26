package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/google/uuid"
)

var _ controller.IdentityService = (*MemoryIdentityService)(nil)

// MemoryIdentityService manages users in memory.
// It implements the IdentityService interface.
type MemoryIdentityService struct {
	mu    sync.RWMutex
	users map[string]dto.User // Map[InternalUserID]User

	// Identity Map: Links a Platform ID (e.g., "discord:123") to an Internal User ID.
	// Key: "source:extID" -> Value: "user-uuid"
	identities map[string]string
}

// NewIdentityService initializes the storage.
func NewIdentityService() *MemoryIdentityService {
	return &MemoryIdentityService{
		users:      make(map[string]dto.User),
		identities: make(map[string]string),
	}
}

// LoginOrRegister finds an existing user or creates a new one.
// source: "web", "discord", "cli"
// extID: The unique ID provided by that platform (e.g. username for web, UserID for Discord)
func (s *MemoryIdentityService) LoginOrRegister(
	_ context.Context,
	username, source, extID string,
) (dto.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lookupKey := fmt.Sprintf("%s:%s", source, extID)

	if internalID, exists := s.identities[lookupKey]; exists {
		return s.users[internalID], nil
	}

	newUserID := fmt.Sprintf("user-%s", uuid.NewString())

	newUser := dto.User{
		ID:       newUserID,
		Username: username,
	}

	s.users[newUserID] = newUser
	s.identities[lookupKey] = newUserID

	return newUser, nil
}
