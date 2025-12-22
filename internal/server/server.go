package server

import (
	"errors"
	"sync"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/model"
)

// ErrGameNotFound is returned when a requested game does not exist.
var ErrGameNotFound = errors.New("game not found")

// Server functions as the entry point for the Battleship application.
// It manages multiple concurrent game instances and routes requests
// to the appropriate game controller.
type Server struct {
	mu    sync.RWMutex
	games map[string]controller.GameController
}

// New creates a new Server instance ready to manage games.
func New() *Server {
	return &Server{
		games: make(map[string]controller.GameController),
	}
}

// CreateGame starts a new game instance and returns its ID.
func (s *Server) CreateGame() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := controller.NewController()
	id := c.Info().ID
	s.games[id] = c

	return id, nil
}

// Info retrieves the state of a specific game.
func (s *Server) Info(gameID string) (controller.GameInfo, error) {
	return execute(s, gameID, func(c controller.GameController) (controller.GameInfo, error) {
		return c.Info(), nil
	})
}

// Join adds a player to a specific game.
func (s *Server) Join(gameID string) (string, error) {
	return execute(s, gameID, func(c controller.GameController) (string, error) {
		return c.Join()
	})
}

// PlaceShip places a ship for a player in a specific game.
func (s *Server) PlaceShip(
	gameID string,
	playerID string,
	shipName string,
	x, y int,
	orientation string,
) error {
	return executeVoid(s, gameID, func(c controller.GameController) error {
		shipType, err := parseShipType(shipName)
		if err != nil {
			return err
		}

		orient, err := parseOrientation(orientation)
		if err != nil {
			return err
		}

		return c.PlaceShip(playerID, shipType, model.Coordinate{X: x, Y: y}, orient)
	})
}

// Ready marks a player as ready in a specific game.
func (s *Server) Ready(gameID string, playerID string) error {
	return executeVoid(s, gameID, func(c controller.GameController) error {
		return c.Ready(playerID)
	})
}

// Fire executes a player's attack in a specific game.
func (s *Server) Fire(gameID string, attackerID string, x, y int) (string, error) {
	return execute(s, gameID, func(c controller.GameController) (string, error) {
		result, err := c.Fire(attackerID, model.Coordinate{X: x, Y: y})
		if err != nil {
			return "", err
		}
		return formatResult(result), nil
	})
}

// Action represents a function that performs an operation on a GameController and returns a result.
type Action[T any] func(c controller.GameController) (T, error)

// VoidAction represents a function that performs an operation on a GameController without returning a result.
type VoidAction func(c controller.GameController) error

// execute is a generic helper that handles game lookup and error propagation for methods returning a value.
func execute[T any](s *Server, gameID string, fn Action[T]) (T, error) {
	s.mu.RLock()
	g, exists := s.games[gameID]
	s.mu.RUnlock()

	if !exists {
		var zero T
		return zero, ErrGameNotFound
	}
	return fn(g)
}

// executeVoid is a helper that handles game lookup and error propagation for methods returning only an error.
func executeVoid(s *Server, gameID string, fn VoidAction) error {
	s.mu.RLock()
	g, exists := s.games[gameID]
	s.mu.RUnlock()

	if !exists {
		return ErrGameNotFound
	}
	return fn(g)
}
