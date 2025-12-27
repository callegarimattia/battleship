package tui

import (
	"fmt"
	"strings"

	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/callegarimattia/battleship/internal/tui/rules"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	// Root message
	var content string

	switch m.State {
	case StateLogin:
		content = m.viewLogin()
	case StateLobby:
		content = m.viewLobby()
	case StateGame:
		if m.GameView == nil {
			content = "Loading game state..."
		} else {
			content = m.viewGame()
		}
	default:
		content = "Unknown State"
	}

	// Error Overlay
	if m.Err != nil {
		errBox := StyleErrorBox.Render(
			fmt.Sprintf("ERROR\n\n%v\n\n[Q] Dismiss", m.Err),
		)
		content = fmt.Sprintf("%s\n\n%s", content, errBox)
	}

	if m.Width > 0 && m.Height > 0 {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, content)
	}

	return content
}

// --- View Helpers ---

func (m Model) viewLogin() string {
	return fmt.Sprintf(
		"\n%s\n\n%s\n\n[Enter] Login",
		StyleTitle.Render("BATTLESHIP TUI"),
		m.LoginInput.View(),
	)
}

func (m Model) viewLobby() string {
	var s strings.Builder
	s.WriteString(StyleTitle.Render("LOBBY") + "\n\n")
	if len(m.Matches) == 0 {
		s.WriteString("No active matches found.\n")
	}
	for i, match := range m.Matches {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}

		// "  Host: [Hostname]           [PlayCount/2]"
		line := fmt.Sprintf(
			"%s Host: %-20s [%d/2]",
			cursor,
			match.HostName,
			match.PlayerCount,
		)

		if m.Cursor == i {
			s.WriteString(
				lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("205")).
					Render(line) +
					"\n",
			)
		} else {
			s.WriteString(line + "\n")
		}
	}
	s.WriteString("\n[C] Create New Match | [Enter] Join Selected | [R] Refresh")
	return s.String()
}

func (m Model) viewGame() string {
	// 1. Determine Base Color based on State
	var baseColor lipgloss.Color
	stateLabel := ""

	if m.GameView.State == dto.StateFinished {
		if m.GameView.Winner == m.GameView.Me.ID {
			baseColor = lipgloss.Color("#FFD700") // Gold (Win)
			stateLabel = "VICTORY"
		} else {
			baseColor = lipgloss.Color("#DC143C") // Crimson (Loss)
			stateLabel = "DEFEAT"
		}
	} else if m.SetupPhase || m.GameView.State == dto.StateSetup {
		baseColor = lipgloss.Color("#00BFFF") // Deep Sky Blue (Setup)
		stateLabel = "SETUP PHASE"
	} else if m.GameView.Turn == m.GameView.Me.ID {
		baseColor = lipgloss.Color("#00FA9A") // Medium Spring Green (My Turn)
		stateLabel = "YOUR TURN"
	} else {
		baseColor = lipgloss.Color("#FF4500") // Orange Red (Enemy Turn)
		stateLabel = "OPPONENT'S TURN"
	}

	// 2. Styles
	styleBorder := StyleBoardBorder.Copy().BorderForeground(baseColor)
	styleLabel := lipgloss.NewStyle().Foreground(baseColor).Bold(true)

	// 3. Render Content
	instructions := styleLabel.Render(m.getInstructions())

	// Boards
	showMyCursor := m.SetupPhase && m.CurrentShipIdx < len(m.ShipsToPlace)
	showEnemyCursor := !m.SetupPhase && m.GameView.State == dto.StatePlaying &&
		m.GameView.Turn == m.GameView.Me.ID

	myBoard := m.renderBoard(m.GameView.Me.Board, showMyCursor, true, styleBorder)
	enemyBoard := m.renderBoard(m.GameView.Enemy.Board, showEnemyCursor, false, styleBorder)

	boards := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			MarginRight(4).
			Render(lipgloss.JoinVertical(lipgloss.Left, styleLabel.Render(stateLabel), styleLabel.Render("YOUR FLEET"), myBoard)),
		lipgloss.JoinVertical(lipgloss.Left, "", styleLabel.Render("ENEMY WATERS"), enemyBoard),
	)

	return fmt.Sprintf("%s\n\n%s", boards, instructions)
}

func (m Model) getInstructions() string {
	if m.GameView.State == dto.StateFinished {
		res := "LOSE"
		if m.GameView.Winner == m.GameView.Me.ID {
			res = "WIN"
		}
		return fmt.Sprintf("GAME OVER - YOU %s! Winner: %s", res, m.GameView.Winner)
	} else if m.SetupPhase {
		if m.CurrentShipIdx < len(m.ShipsToPlace) {
			size := m.ShipsToPlace[m.CurrentShipIdx]
			orient := "HORZ"
			if m.ShipOrientation {
				orient = "VERT"
			}

			action := "Waiting for game..."
			if m.GameView.State == dto.StateSetup {
				action = "[Enter] Place"
			}

			return fmt.Sprintf("SETUP: Place Ship Size %d (%s) | [Arrows] Move | [R] Rotate | %s", size, orient, action)
		} else {
			return "SETUP: Waiting for opponent..."
		}
	} else {
		if m.GameView.Turn == m.GameView.Me.ID {
			return "YOUR TURN: Select target on enemy board | [Arrows] Move | [Enter] Fire"
		} else {
			return "OPPONENT'S TURN: Please wait..."
		}
	}
}

func (m Model) renderBoard(
	board dto.BoardView,
	showCursor bool,
	isMe bool,
	borderStyle lipgloss.Style,
) string {
	var rows []string

	// Header row: 0 1 2 ...
	header := "  "
	for x := 0; x < board.Size; x++ {
		header += fmt.Sprintf("%d ", x)
	}
	rows = append(rows, header)

	for y := 0; y < board.Size; y++ {
		rowStr := fmt.Sprintf("%c ", 'A'+y)
		for x := 0; x < board.Size; x++ {
			cell := board.Grid[y][x]
			rendered := m.renderCell(x, y, cell, board, isMe, showCursor)
			rowStr += rendered + " "
		}
		rows = append(rows, rowStr)
	}

	return borderStyle.Render(strings.Join(rows, "\n"))
}

func (m Model) renderCell(
	x, y int,
	cell dto.CellState,
	board dto.BoardView,
	isMe, showCursor bool,
) string {
	symbol := "·" // Empty/Unknown default for water
	style := StyleCellEmpty

	switch cell {
	case dto.CellShip:
		symbol = "S"
		style = StyleCellShip
	case dto.CellHit:
		symbol = "X"
		style = StyleCellHit
	case dto.CellMiss:
		symbol = "O"
		style = StyleCellMiss
	case dto.CellSunk:
		symbol = "#"
		style = StyleCellSunk
	case dto.CellUnknown:
		symbol = "~"
		style = StyleCellUnknown
	}

	// Render basic cell
	rendered := style.Render(symbol)

	// Ghost Ship Overlay (Setup only)
	if isMe && m.SetupPhase && m.CurrentShipIdx < len(m.ShipsToPlace) {
		size := m.ShipsToPlace[m.CurrentShipIdx]
		// Check if this cell is part of the ghost ship projection
		isGhost := false
		if m.ShipOrientation { // Vertical
			if x == m.CursorX && y >= m.CursorY && y < m.CursorY+size {
				isGhost = true
			}
		} else { // Horizontal
			if y == m.CursorY && x >= m.CursorX && x < m.CursorX+size {
				isGhost = true
			}
		}

		if isGhost {
			err := rules.CanPlaceShip(
				board,
				size,
				m.CursorX,
				m.CursorY,
				m.ShipOrientation,
			)
			if err == nil {
				rendered = StyleCellGhost.Render(symbol)
			}
		}
	}

	// Cursor overlay
	if showCursor && x == m.CursorX && y == m.CursorY {
		rendered = StyleCursor.Render(symbol)
	}

	return rendered
}
