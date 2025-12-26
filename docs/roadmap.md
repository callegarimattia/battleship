# Battleship Project Roadmap

This document outlines the implementation strategy for the next major phases of the Battleship project: Persistence, Discord Bot, Web Client, and CLI Client.

## 1. Database Persistence

**Goal**: Replace the in-memory storage with a persistent database to ensure game state survives server restarts and crashes.

### Strategy

- **Database**: PostgreSQL is recommended for relational data (Games <-> Players).
- **Schema Design**:
  - `games` table: `id` (UUID), `state`, `turn`, `winner`, `created_at`, `updated_at`.
  - `players` table: `id` (UUID), `game_id` (FK), `user_id` (active user), `board_state` (JSONB), `fleet_state` (JSONB).
  - Using JSONB for `board_state` avoids creating 100 rows per game for grid cells while maintaining queryability if needed.
- **Implementation**:
  - Create a new implementation of `LobbyService` and `GameService` (e.g., `PostgresService`) in `internal/service/postgres.go`.
  - Use `pgx` or `gorm` for database interaction.
  - Update `main.go` to initialize `PostgresService` instead of `MemoryService` based on configuration.

## 2. Discord Bot

**Goal**: Allow users to play Battleship directly within a Discord channel.

### Strategy

- **Library**: `github.com/bwmarrin/discordgo`.
- **Architecture**:
  - The Bot runs in the same process as the HTTP server.
  - It acts as a primary adapter, calling `AppController` just like the HTTP handlers do.
- **Interaction Model**:
  - Commands: `/battleship challenge @user`, `/battleship place <ship> <coords>`, `/battleship fire <coords>`.
  - Visuals: Use Discord Embeds to render the board grids using emojis (e.g., 🟦 for water, 🟥 for hit, ⬜ for miss).
- **State Management**:
  - Map Discord Channel IDs or User IDs to internal Game IDs.

## 3. Web Client (GUI)

**Goal**: A modern, responsive web application for playing the game in a browser.

### Strategy

- **Tech Stack**: React, Vue, or Svelte (via Vite).
- **Communication**: REST API (consume the existing endpoints documented in `openapi.yaml`).
- **Features**:
  - **Lobby**: List available games, create new game.
  - **Game Board**: Interactive grid for placing ships (drag & drop) and firing shots (point & click).
  - **Real-time Updates**: Polling (MVP) or WebSocket (Upgrade) for turn updates.
- **Deployment**: Hosted on the same server (served via `embed.FS`) or separate static hosting (Vercel/Netlify).

## 4. CLI Client

**Goal**: A terminal-based interface for geeks and testing.

### Strategy

- **Language**: Go.
- **UI Library**: `github.com/charmbracelet/bubbletea` (TUI framework) for a rich terminal experience.
- **Communication**: REST API client.
- **Key Features**:
  - Terminal-based grid rendering.
  - Keyboard navigation.
  - Configurable API endpoint.
