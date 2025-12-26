// Package env provides centralized environment variable management.
package env

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration from environment variables.
type Config struct {
	// Server configuration
	Port      string
	RateLimit int
	JWTSecret string

	// Discord bot configuration
	DiscordToken string
	DiscordAppID string
}

// LoadServerConfig loads configuration required for the HTTP server.
func LoadServerConfig() (*Config, error) {
	cfg := &Config{
		Port:      getEnvOrDefault("PORT", "8080"),
		RateLimit: getEnvAsIntOrDefault("RATE_LIMIT", 20),
		JWTSecret: getEnvOrDefault("JWT_SECRET", "secret"),
	}

	return cfg, nil
}

// LoadBotConfig loads configuration required for the Discord bot.
func LoadBotConfig() (*Config, error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is required")
	}

	appID := os.Getenv("DISCORD_APP_ID")
	if appID == "" {
		return nil, fmt.Errorf("DISCORD_APP_ID environment variable is required")
	}

	cfg := &Config{
		DiscordToken: token,
		DiscordAppID: appID,
		JWTSecret:    getEnvOrDefault("JWT_SECRET", "secret"),
	}

	return cfg, nil
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}
