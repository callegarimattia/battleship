package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/callegarimattia/battleship/internal/model"
	"github.com/google/uuid"
)

var (
	_ controller.LobbyService = (*MemoryService)(nil)
	_ controller.GameService  = (*MemoryService)(nil)
)

// MemoryService is an in-memory implementation of the lobby and game service.
type MemoryService struct {
	games   map[string]*safeGame
	gamesMu sync.RWMutex
}

type safeGame struct {
	id        string
	game      *model.Game
	host      string
	guest     string
	createdAt time.Time
	mu        sync.Mutex
}

// NewMemoryService creates a new in-memory lobby and game service.
func NewMemoryService() *MemoryService {
	return &MemoryService{
		games: make(map[string]*safeGame),
	}
}

// CreateMatch initializes a new game with the host player joined.
func (s *MemoryService) CreateMatch(_ context.Context, hostID string) (string, error) {
	gameID := fmt.Sprintf("game-%v", uuid.NewString())
	sg := &safeGame{
		game:      model.NewGame(),
		id:        gameID,
		createdAt: time.Now(),
		host:      hostID,
	}

	err := sg.game.Join(hostID, model.StandardFleet())
	if err != nil {
		return "", err
	}

	s.gamesMu.Lock()
	s.games[gameID] = sg
	s.gamesMu.Unlock()

	return gameID, nil
}

// ListMatches returns all games and their summaries.
func (s *MemoryService) ListMatches(_ context.Context) ([]dto.MatchSummary, error) {
	s.gamesMu.RLock()
	defer s.gamesMu.RUnlock()

	matches := make([]dto.MatchSummary, len(s.games))
	for matchID, sg := range s.games {
		sg.mu.Lock()
		matches = append(matches, dto.MatchSummary{
			ID:          matchID,
			CreatedAt:   sg.createdAt,
			HostName:    sg.host,
			PlayerCount: playerCountUnsafe(sg),
		})
		sg.mu.Unlock()
	}

	return matches, nil
}

// JoinMatch adds a player to an existing match.
func (s *MemoryService) JoinMatch(
	_ context.Context,
	matchID, playerID string,
) (dto.GameView, error) {
	s.gamesMu.RLock()
	defer s.gamesMu.RUnlock()

	game, err := s.getSafeGame(matchID)
	if err != nil {
		return dto.GameView{}, err
	}

	err = game.game.Join(playerID, nil)
	if err != nil {
		return dto.GameView{}, err
	}

	game.guest = playerID

	return game.game.GetView(playerID)
}

func (s *MemoryService) getSafeGame(matchID string) (*safeGame, error) {
	s.gamesMu.RLock()
	defer s.gamesMu.RUnlock()

	sg, exists := s.games[matchID]
	if !exists {
		return nil, errors.New("match not found")
	}

	return sg, nil
}

func playerCountUnsafe(sg *safeGame) (count int) {
	if sg.host != "" {
		count++
	}
	if sg.guest != "" {
		count++
	}
	return count
}
