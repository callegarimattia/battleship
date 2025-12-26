// Package main is the entry point for the Discord bot.
package main

import (
	"context"
	"log"

	"github.com/callegarimattia/battleship/internal/bot"
	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/env"
	"github.com/callegarimattia/battleship/internal/events"
	"github.com/callegarimattia/battleship/internal/service"
)

func main() {
	// Load configuration
	cfg, err := env.LoadBotConfig()
	if err != nil {
		log.Fatalf("Failed to load bot config: %v", err)
	}

	// Initialize services
	eventBus := events.NewMemoryEventBus()
	identityService := service.NewIdentityService(cfg.JWTSecret)
	memoryService := service.NewMemoryService(eventBus)

	// Create controller
	ctrl := controller.NewAppController(identityService, memoryService, memoryService)

	// Create and start bot
	discordBot, err := bot.NewDiscordBot(cfg.DiscordToken, cfg.DiscordAppID, ctrl, eventBus)
	if err != nil {
		log.Fatalf("Failed to create Discord bot: %v", err)
	}

	log.Println("Starting Discord bot...")
	if err := discordBot.Start(context.Background()); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
