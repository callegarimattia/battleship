package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorWin    = lipgloss.Color("#FFD700") // Gold
	ColorLose   = lipgloss.Color("#DC143C") // Crimson
	ColorSetup  = lipgloss.Color("#00BFFF") // Deep Sky Blue
	ColorMyTurn = lipgloss.Color("#00FA9A") // Medium Spring Green
	ColorOpTurn = lipgloss.Color("#FF4500") // Orange Red

	// General Styles
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	StyleBoardBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1)

	StyleCellEmpty   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // Dark Gray
	StyleCellShip    = lipgloss.NewStyle().Foreground(lipgloss.Color("212")) // Pink
	StyleCellHit     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	StyleCellMiss    = lipgloss.NewStyle().Foreground(lipgloss.Color("45"))  // Cyan
	StyleCellSunk    = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	StyleCellUnknown = lipgloss.NewStyle().Foreground(lipgloss.Color("237")) // Gray
	StyleCellGhost   = lipgloss.NewStyle().Foreground(lipgloss.Color("57"))  // Purple/Ghost
	StyleCursor      = lipgloss.NewStyle().
				Background(lipgloss.Color("252")).
				Foreground(lipgloss.Color("0"))

	StyleErrorBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("196")). // Red
			Foreground(lipgloss.Color("196")).
			Padding(1, 2).
			Align(lipgloss.Center)
)
