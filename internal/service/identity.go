package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/golang-jwt/jwt/v5"
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

	jwtSecret string
}

// NewIdentityService initializes the storage.
func NewIdentityService(jwtSecret string) *MemoryIdentityService {
	if jwtSecret == "" {
		jwtSecret = "secret"
	}
	return &MemoryIdentityService{
		users:      make(map[string]dto.User),
		identities: make(map[string]string),
		jwtSecret:  jwtSecret,
	}
}

// LoginOrRegister finds an existing user or creates a new one.
// source: "web", "discord", "cli"
// extID: The unique ID provided by that platform (e.g. username for web, UserID for Discord)
func (s *MemoryIdentityService) LoginOrRegister(
	_ context.Context,
	username, source, extID string,
) (dto.AuthResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var user dto.User
	lookupKey := fmt.Sprintf("%s:%s", source, extID)

	if internalID, exists := s.identities[lookupKey]; exists {
		user = s.users[internalID]
	} else {
		newUserID := fmt.Sprintf("user-%s", uuid.NewString())
		newUser := dto.User{
			ID:       newUserID,
			Username: username,
		}

		s.users[newUserID] = newUser
		s.identities[lookupKey] = newUserID
		user = newUser
	}

	// Generate JWT
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"name": user.Username,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return dto.AuthResponse{
		Token: signedToken,
		User:  user,
	}, nil
}
