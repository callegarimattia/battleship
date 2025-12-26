package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/callegarimattia/battleship/internal/dto"
)

// CoordinateToChess converts numeric coordinates to chess-style (A-J, 1-10).
func CoordinateToChess(x, y int) string {
	if x < 0 || x > 9 || y < 0 || y > 9 {
		return fmt.Sprintf("(%d,%d)", x, y)
	}
	col := string(rune('A' + x))
	row := y + 1
	return fmt.Sprintf("%s%d", col, row)
}

// ChessToCoordinate converts chess-style coordinates to numeric (0-9, 0-9).
func ChessToCoordinate(chess string) (x, y int, err error) {
	chess = strings.ToUpper(strings.TrimSpace(chess))
	if len(chess) < 2 {
		return 0, 0, fmt.Errorf("invalid coordinate format")
	}

	col := chess[0]
	if col < 'A' || col > 'J' {
		return 0, 0, fmt.Errorf("column must be A-J")
	}
	x = int(col - 'A')

	var row int
	_, err = fmt.Sscanf(chess[1:], "%d", &row)
	if err != nil || row < 1 || row > 10 {
		return 0, 0, fmt.Errorf("row must be 1-10")
	}
	y = row - 1

	return x, y, nil
}

// GetShipName returns the ship name for a given size.
func GetShipName(size int) string {
	switch size {
	case 5:
		return "Carrier"
	case 4:
		return "Battleship"
	case 3:
		return "Cruiser"
	case 2:
		return "Destroyer"
	default:
		return fmt.Sprintf("Ship (size %d)", size)
	}
}

// FormatGameState creates a Discord embed for the game state.
func FormatGameState(view *dto.GameView) *discordgo.MessageEmbed { //nolint:funlen
	embed := &discordgo.MessageEmbed{
		Title: "⚓ Battleship Game",
		Color: getColorForState(view.State),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Game State",
				Value:  string(view.State),
				Inline: true,
			},
		},
	}

	// Add turn information with player ID (we don't have usernames in GameView)
	if view.Turn != "" {
		turnPlayer := "You"
		if view.Enemy.ID == view.Turn {
			turnPlayer = "Opponent"
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Current Turn",
			Value:  turnPlayer,
			Inline: true,
		})
	}

	// Add winner if game is over
	if view.Winner != "" {
		winnerText := "You won! 🎉"
		if view.Winner == view.Enemy.ID {
			winnerText = "Opponent won"
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🏆 Winner",
			Value:  winnerText,
			Inline: false,
		})
	}

	// Add your board with chess coordinates
	myBoard := formatBoardWithChessCoords(view.Me.Board)
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📍 Your Board",
		Value:  myBoard,
		Inline: false,
	})

	// Add enemy board with chess coordinates (if present)
	if view.Enemy.Board.Size != 0 {
		enemyBoard := formatBoardWithChessCoords(view.Enemy.Board)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎯 Enemy Board",
			Value:  enemyBoard,
			Inline: false,
		})
	}

	// Add fleet status with ship names
	myFleet := formatFleetWithNames(view.Me.Fleet)
	enemyFleet := formatFleetWithNames(view.Enemy.Fleet)
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

func formatBoardWithChessCoords(board dto.BoardView) string {
	var sb strings.Builder

	// Header with column letters
	sb.WriteString("```\n   A B C D E F G H I J\n")

	for y := 0; y < board.Size; y++ {
		fmt.Fprintf(&sb, "%2d ", y+1)
		for x := 0; x < board.Size; x++ {
			cell := board.Grid[y][x]
			sb.WriteString(cellToEmoji(cell))
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("```")
	return sb.String()
}

func cellToEmoji(cell dto.CellState) string {
	switch cell {
	case dto.CellEmpty, dto.CellUnknown:
		return "·"
	case dto.CellShip:
		return "■"
	case dto.CellHit:
		return "X"
	case dto.CellMiss:
		return "○"
	case dto.CellSunk:
		return "☠"
	default:
		return "·"
	}
}

func formatFleetWithNames(fleet map[int]int) string {
	if len(fleet) == 0 {
		return "All ships sunk!"
	}

	var sb strings.Builder
	for size := 5; size >= 2; size-- {
		if count, ok := fleet[size]; ok && count > 0 {
			shipName := GetShipName(size)
			fmt.Fprintf(&sb, "%s (size %d): %d\n", shipName, size, count)
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
