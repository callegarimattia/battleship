// Package bot provides Discord integration for the Battleship game.
package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/callegarimattia/battleship/internal/controller"
)

// DiscordBot represents the Discord bot instance.
type DiscordBot struct {
	session *discordgo.Session
	ctrl    *controller.AppController
	appID   string

	// Track active match per Discord user
	mu            sync.RWMutex
	activeMatches map[string]string // Discord user ID -> match ID
}

// NewDiscordBot creates a new Discord bot instance.
func NewDiscordBot(token, appID string, ctrl *controller.AppController) (*DiscordBot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	if appID == "" {
		return nil, fmt.Errorf("app ID is required")
	}

	bot := &DiscordBot{
		session:       session,
		ctrl:          ctrl,
		appID:         appID,
		activeMatches: make(map[string]string),
	}

	// Register interaction handler
	session.AddHandler(bot.handleInteraction)

	return bot, nil
}

// Start opens the Discord connection and registers commands.
func (b *DiscordBot) Start(ctx context.Context) error {
	// Open websocket connection
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord connection: %w", err)
	}

	log.Println("Discord bot connected successfully")

	// Register slash commands
	if err := b.registerCommands(); err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	log.Println("Slash commands registered successfully")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stop:
		log.Println("Received shutdown signal")
	case <-ctx.Done():
		log.Println("Context cancelled")
	}

	return b.Shutdown()
}

// Shutdown gracefully closes the Discord connection.
func (b *DiscordBot) Shutdown() error {
	log.Println("Shutting down Discord bot...")
	return b.session.Close()
}
