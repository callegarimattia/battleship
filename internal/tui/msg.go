package tui

import (
	"time"

	"github.com/callegarimattia/battleship/internal/dto"
	tea "github.com/charmbracelet/bubbletea"
)

// Messages
type (
	PerformLoginMsg struct{}
	GotMatchesMsg   []dto.MatchSummary
	MatchJoinedMsg  struct{ ID string }
	GotGameMsg      *dto.GameView
	ShipPlacedMsg   struct{ Game *dto.GameView }
	TickMsg         time.Time
)

// Helper to generate tick command
func TickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
