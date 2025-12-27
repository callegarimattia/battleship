package tui

import (
	"fmt"

	"github.com/callegarimattia/battleship/internal/client"
	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/callegarimattia/battleship/internal/tui/rules"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// --- Global Keys (Always generic) ---
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// --- Error Handling ---
	// Block other updates while error is shown
	if m.Err != nil {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "q", "esc":
				m.Err = nil // Dismiss error
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case error:
		m.Err = msg
		return m, nil
	}

	switch m.State {
	case StateLogin:
		return m.updateLogin(msg)
	case StateLobby:
		return m.updateLobby(msg)
	case StateGame:
		return m.updateGame(msg)
	}
	return m, cmd
}

// --- Sub-Update Functions ---

func (m *Model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.LoginInput, cmd = m.LoginInput.Update(msg)

	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter {
		username := m.LoginInput.Value()
		return m, func() tea.Msg {
			_, err := m.Client.Login(username)
			if err != nil {
				return err
			}
			return PerformLoginMsg{}
		}
	}

	if _, ok := msg.(PerformLoginMsg); ok {
		m.State = StateLobby
		return m, fetchMatchesCmd(m.Client)
	}
	return m, cmd
}

func (m *Model) updateLobby(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case GotMatchesMsg:
		m.Matches = msg
	case tea.KeyMsg:
		return m.handleLobbyKeys(msg)
	case MatchJoinedMsg:
		return m.handleMatchJoined(msg)
	}
	return m, nil
}

func (m *Model) handleLobbyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Matches)-1 {
			m.Cursor++
		}
	case "r":
		return m, fetchMatchesCmd(m.Client)
	case "c":
		return m, func() tea.Msg {
			id, err := m.Client.CreateMatch()
			if err != nil {
				return err
			}
			return MatchJoinedMsg{ID: id}
		}
	case "enter":
		if len(m.Matches) > 0 {
			selectedID := m.Matches[m.Cursor].ID
			return m, func() tea.Msg {
				_, err := m.Client.JoinMatch(selectedID)
				if err != nil {
					return err
				}
				return MatchJoinedMsg{ID: selectedID}
			}
		}
	}
	return m, nil
}

func (m *Model) handleMatchJoined(msg MatchJoinedMsg) (tea.Model, tea.Cmd) {
	m.GameID = msg.ID
	m.State = StateGame
	// Initialize game state params
	m.CursorX = 0
	m.CursorY = 0
	m.CurrentShipIdx = 0
	m.SetupPhase = true
	// Kick off WS listener and initial fetch
	return m, tea.Batch(
		func() tea.Msg { // Initial fetch
			g, err := m.Client.GetGameState(m.GameID)
			if err != nil {
				return err
			}
			return GotGameMsg(g)
		},
		subToWSCmd(m.Client, m.GameID),
	)
}

func subToWSCmd(c *client.Client, matchID string) tea.Cmd {
	return func() tea.Msg {
		ch, err := c.SubscribeToMatch(matchID)
		if err != nil {
			return err
		}
		return listenForUpdates(ch)
	}
}

// listenForUpdates waits for a signal from the WS channel
func listenForUpdates(ch <-chan *dto.WSEvent) tea.Msg {
	evt, ok := <-ch
	if !ok {
		return nil
	}
	return GameUpdateMsg{Event: evt, Channel: ch}
}

func (m *Model) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case GotGameMsg:
		return m.handleGotGame(msg)
	case tea.KeyMsg:
		return m.handleGameKeys(msg)
	case ShipPlacedMsg:
		m.CurrentShipIdx++
		return m.handleGotGame(GotGameMsg(msg.Game))
	case GameUpdateMsg:
		// Handle Event
		var cmd tea.Cmd
		if msg.Event.Type == "game_update" && msg.Event.Payload != nil {
			// Update state
			var newModel tea.Model
			newModel, cmd = m.handleGotGame(GotGameMsg(msg.Event.Payload))
			m = newModel.(*Model) // Type assertion due to interface return
		} else if msg.Event.Type == "error" {
			m.Err = fmt.Errorf("server error: %s", msg.Event.Error)
		}

		// Listen for next event
		return m, tea.Batch(
			cmd,
			func() tea.Msg {
				return listenForUpdates(msg.Channel)
			},
		)
	}
	return m, nil
}

func (m *Model) handleGotGame(msg GotGameMsg) (tea.Model, tea.Cmd) {
	if msg == nil {
		return m, nil
	}
	m.GameView = msg
	switch m.GameView.State {
	case dto.StatePlaying, dto.StateFinished:
		m.SetupPhase = false
	default:
		m.SetupPhase = true
	}
	return m, nil
}

func (m *Model) handleGameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.CursorY > 0 {
			m.CursorY--
		}
	case "down", "j":
		if m.CursorY < BoardSize-1 {
			m.CursorY++
		}
	case "left", "h":
		if m.CursorX > 0 {
			m.CursorX--
		}
	case "right", "l":
		if m.CursorX < BoardSize-1 {
			m.CursorX++
		}
	case "r":
		if m.SetupPhase {
			m.ShipOrientation = !m.ShipOrientation
		}
	case "enter", "space":
		return m.handleAction()
	}
	return m, nil
}

func (m *Model) handleAction() (tea.Model, tea.Cmd) {
	if m.GameView == nil {
		return m, nil
	}

	if m.SetupPhase {
		return m.handleSetupAction()
	} else if m.GameView.State == dto.StatePlaying && m.GameView.Turn == m.GameView.Me.ID {
		return m.handlePlayAction()
	}
	return m, nil
}

func (m *Model) handleSetupAction() (tea.Model, tea.Cmd) {
	if m.CurrentShipIdx >= len(m.ShipsToPlace) {
		return m, nil
	}

	size := m.ShipsToPlace[m.CurrentShipIdx]
	cx, cy, vert := m.CursorX, m.CursorY, m.ShipOrientation

	// Validation: Check Game State
	if m.GameView.State != dto.StateSetup {
		return m, nil
	}

	// Validation: Check Rules
	if err := rules.CanPlaceShip(m.GameView.Me.Board, size, cx, cy, vert); err != nil {
		return m, func() tea.Msg {
			return err
		}
	}

	return m, func() tea.Msg {
		g, err := m.Client.PlaceShip(m.GameID, size, cx, cy, vert)
		if err != nil {
			return err
		}
		return ShipPlacedMsg{Game: g}
	}
}

func (m *Model) handlePlayAction() (tea.Model, tea.Cmd) {
	cx, cy := m.CursorX, m.CursorY

	// Validation: Check if cell can be attacked
	if err := rules.CanAttack(m.GameView.Enemy.Board, cx, cy); err != nil {
		return m, func() tea.Msg {
			return err
		}
	}

	return m, func() tea.Msg {
		g, err := m.Client.Attack(m.GameID, cx, cy)
		if err != nil {
			return err
		}
		return GotGameMsg(g)
	}
}

func fetchMatchesCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		matches, err := c.ListMatches()
		if err != nil {
			return err
		}
		return GotMatchesMsg(matches)
	}
}
