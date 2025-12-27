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
	GameUpdateMsg   struct {
		Event   *dto.WSEvent
		Channel <-chan *dto.WSEvent
	}
)

// TickCmd returns a command that triggers a tick.
func TickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
