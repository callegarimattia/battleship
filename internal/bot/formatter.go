package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/callegarimattia/battleship/internal/dto"
)

// FormatGameState creates a Discord embed for the game state.
func FormatGameState(view *dto.GameView) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: "⚓ Battleship Game",
		Color: getColorForState(view.State),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Game State",
				Value:  string(view.State),
				Inline: true,
			},
			{
				Name:   "Current Turn",
				Value:  view.Turn,
				Inline: true,
			},
		},
	}

	if view.Winner != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🏆 Winner",
			Value:  view.Winner,
			Inline: false,
		})
	}

	// Add your board
	myBoard := formatBoard(view.Me.Board, true)
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📍 Your Board",
		Value:  myBoard,
		Inline: false,
	})

	// Add enemy board
	enemyBoard := formatBoard(view.Enemy.Board, false)
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🎯 Enemy Board",
		Value:  enemyBoard,
		Inline: false,
	})

	// Add fleet status
	myFleet := formatFleet(view.Me.Fleet)
	enemyFleet := formatFleet(view.Enemy.Fleet)
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:   "🚢 Your Fleet",
			Value:  myFleet,
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "🚢 Enemy Fleet",
			Value:  enemyFleet,
			Inline: true,
		},
	)

	return embed
}

func formatBoard(board dto.BoardView, showShips bool) string {
	var sb strings.Builder

	// Header with column numbers
	sb.WriteString("```\n  0 1 2 3 4 5 6 7 8 9\n")

	for y := 0; y < board.Size; y++ {
		sb.WriteString(fmt.Sprintf("%d ", y))
		for x := 0; x < board.Size; x++ {
			cell := board.Grid[y][x]
			sb.WriteString(cellToEmoji(cell, showShips))
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("```")
	return sb.String()
}

func cellToEmoji(cell dto.CellState, showShips bool) string {
	switch cell {
	case dto.CellEmpty:
		return "·"
	case dto.CellShip:
		if showShips {
			return "■"
		}
		return "·"
	case dto.CellHit:
		return "X"
	case dto.CellMiss:
		return "○"
	case dto.CellSunk:
		return "☠"
	case dto.CellUnknown:
		return "?"
	default:
		return "·"
	}
}

func formatFleet(fleet map[int]int) string {
	if len(fleet) == 0 {
		return "All ships sunk!"
	}

	var sb strings.Builder
	for size := 5; size >= 2; size-- {
		if count, ok := fleet[size]; ok && count > 0 {
			fmt.Fprintf(&sb, "Size %d: %d\n", size, count)
		}
	}
	return sb.String()
}

func getColorForState(state dto.GameState) int {
	switch state {
	case dto.StateSetup:
		return 0xffaa00 // Orange
	case dto.StatePlaying:
		return 0x0099ff // Blue
	case dto.StateFinished:
		return 0x00ff00 // Green
	default:
		return 0x808080 // Gray
	}
}
