// Package controller contains the main application controller orchestrating the flow.
package controller

import (
	"context"

	"github.com/callegarimattia/battleship/internal/dto"
)

// NotificationService handles event publishing and subscription.
type NotificationService interface {
	Subscribe(matchID string) (Subscription, <-chan *dto.GameEvent)
	Publish(event *dto.GameEvent)
}

// Subscription represents a subscription to events.
type Subscription interface {
	Unsubscribe()
}

// IdentityService handles user registration and login.
type IdentityService interface {
	// LoginOrRegister finds an existing user or creates a new one.
	// source: "web", "discord", "cli"
	// extID: The unique ID from the platform (e.g. Discord User ID, or just the username for Web)
	LoginOrRegister(ctx context.Context, username, source, extID string) (dto.AuthResponse, error)
}

// LobbyService handles finding and creating matches.
type LobbyService interface {
	// CreateMatch initializes a game in 'Waiting' state with the host joined.
	CreateMatch(ctx context.Context, hostID string) (string, error)
	// ListMatches returns all games currently in 'Waiting' state.
	ListMatches(ctx context.Context) ([]dto.MatchSummary, error)
	// JoinMatch adds the player to the game.
	// If successful, the game transitions to 'Setup'.
	JoinMatch(ctx context.Context, matchID, playerID string) (dto.GameView, error)
}

// GameService handles the actual gameplay (Setup -> Playing -> GameOver).
type GameService interface {
	// PlaceShip handles the setup phase.
	PlaceShip(
		ctx context.Context,
		matchID, playerID string,
		shipID int,
		x, y int,
		vertical bool,
	) (dto.GameView, error)
	// Attack handles the playing phase.
	Attack(ctx context.Context, matchID, playerID string, x, y int) (dto.GameView, error)
	// GetState is used for refreshing the UI.
	GetState(ctx context.Context, matchID, playerID string) (dto.GameView, error)
}

// AppController is the main controller orchestrating the application flow.
type AppController struct {
	auth     IdentityService
	lobby    LobbyService
	game     GameService
	notifier NotificationService
}

// NewAppController wires everything together.
// NewAppController wires everything together.
func NewAppController(
	a IdentityService,
	l LobbyService,
	g GameService,
	n NotificationService,
) *AppController {
	return &AppController{auth: a, lobby: l, game: g, notifier: n}
}

// Login handles user authentication and registration.
func (c *AppController) Login(
	ctx context.Context,
	username, source, platformID string,
) (dto.AuthResponse, error) {
	return c.auth.LoginOrRegister(ctx, username, source, platformID)
}

// HostGameAction handles a player's request to host a new game.
func (c *AppController) HostGameAction(ctx context.Context, playerID string) (string, error) {
	return c.lobby.CreateMatch(ctx, playerID)
}

// ListGamesAction retrieves the list of current games in the lobby.
func (c *AppController) ListGamesAction(ctx context.Context) ([]dto.MatchSummary, error) {
	return c.lobby.ListMatches(ctx)
}

// JoinGameAction handles a player's request to join an existing game.
func (c *AppController) JoinGameAction(
	ctx context.Context,
	matchID, playerID string,
) (dto.GameView, error) {
	return c.lobby.JoinMatch(ctx, matchID, playerID)
}

// PlaceShipAction handles a ship placement action from a player.
func (c *AppController) PlaceShipAction(
	ctx context.Context,
	matchID, playerID string,
	size, x, y int,
	vertical bool,
) (dto.GameView, error) {
	return c.game.PlaceShip(ctx, matchID, playerID, size, x, y, vertical)
}

// AttackAction handles an attack action from a player.
func (c *AppController) AttackAction(
	ctx context.Context,
	matchID, playerID string,
	x, y int,
) (dto.GameView, error) {
	return c.game.Attack(ctx, matchID, playerID, x, y)
}

// GetGameStateAction retrieves the current state of the game for a player.
func (c *AppController) GetGameStateAction(
	ctx context.Context,
	matchID, playerID string,
) (dto.GameView, error) {
	return c.game.GetState(ctx, matchID, playerID)
}

// SubscribeToMatch allows the handler to subscribe to match events.
func (c *AppController) SubscribeToMatch(
	matchID string,
) (sub Subscription, eventChan <-chan *dto.GameEvent) {
	return c.notifier.Subscribe(matchID)
}
