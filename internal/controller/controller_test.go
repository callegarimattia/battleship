package controller_test

import (
	"context"
	"errors"
	"testing"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/dto"
	m "github.com/callegarimattia/battleship/internal/mocks/controller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupControllerTest(
	t *testing.T,
) (*controller.AppController,
	*m.MockIdentityService,
	*m.MockLobbyService,
	*m.MockGameService,
	*m.MockNotificationService, //nolint
) {
	mockAuth := m.NewMockIdentityService(t)
	mockLobby := m.NewMockLobbyService(t)
	mockGame := m.NewMockGameService(t)
	mockNotifier := m.NewMockNotificationService(t)
	ctrl := controller.NewAppController(mockAuth, mockLobby, mockGame, mockNotifier)
	return ctrl, mockAuth, mockLobby, mockGame, mockNotifier
}

func TestLogin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		username     string
		source       string
		platformID   string
		mockSetup    func(*m.MockIdentityService)
		expectedResp dto.AuthResponse
		expectedErr  error
	}{
		{
			name:       "Success",
			username:   "Alice",
			source:     "web",
			platformID: "Alice",
			mockSetup: func(m *m.MockIdentityService) {
				m.EXPECT().LoginOrRegister(mock.Anything, "Alice", "web", "Alice").
					Return(dto.AuthResponse{
						Token: "token",
						User:  dto.User{ID: "u1", Username: "Alice"},
					}, nil).
					Once()
			},
			expectedResp: dto.AuthResponse{
				Token: "token",
				User:  dto.User{ID: "u1", Username: "Alice"},
			},
			expectedErr: nil,
		},
		{
			name:       "Service Error",
			username:   "Bob",
			source:     "discord",
			platformID: "12345",
			mockSetup: func(m *m.MockIdentityService) {
				m.EXPECT().LoginOrRegister(mock.Anything, "Bob", "discord", "12345").
					Return(dto.AuthResponse{}, errors.New("auth error")).
					Once()
			},
			expectedResp: dto.AuthResponse{},
			expectedErr:  errors.New("auth error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl, mockAuth, _, _, _ := setupControllerTest(t)
			tt.mockSetup(mockAuth)

			resp, err := ctrl.Login(context.Background(), tt.username, tt.source, tt.platformID)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResp, resp)
		})
	}
}

func TestLobbyActions(t *testing.T) {
	t.Parallel()

	t.Run("HostGameAction", func(t *testing.T) {
		t.Parallel()
		ctrl, _, mockLobby, _, _ := setupControllerTest(t)
		mockLobby.EXPECT().CreateMatch(mock.Anything, "p1").Return("match-1", nil).Once()

		id, err := ctrl.HostGameAction(context.Background(), "p1")
		assert.NoError(t, err)
		assert.Equal(t, "match-1", id)
	})

	t.Run("HostGameAction Error", func(t *testing.T) {
		t.Parallel()
		ctrl, _, mockLobby, _, _ := setupControllerTest(t)
		mockLobby.EXPECT().CreateMatch(mock.Anything, "p1").Return("", errors.New("fail")).Once()

		_, err := ctrl.HostGameAction(context.Background(), "p1")
		assert.Error(t, err)
	})

	t.Run("ListGamesAction", func(t *testing.T) {
		t.Parallel()
		ctrl, _, mockLobby, _, _ := setupControllerTest(t)
		expected := []dto.MatchSummary{{ID: "m1"}}
		mockLobby.EXPECT().ListMatches(mock.Anything).Return(expected, nil).Once()

		list, err := ctrl.ListGamesAction(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, list)
	})

	t.Run("JoinGameAction", func(t *testing.T) {
		t.Parallel()
		ctrl, _, mockLobby, _, _ := setupControllerTest(t)
		expected := dto.GameView{State: "SETUP"}
		mockLobby.EXPECT().JoinMatch(mock.Anything, "m1", "p2").Return(expected, nil).Once()

		view, err := ctrl.JoinGameAction(context.Background(), "m1", "p2")
		assert.NoError(t, err)
		assert.Equal(t, expected, view)
	})
}

func TestGameActions(t *testing.T) {
	t.Parallel()

	t.Run("PlaceShipAction", func(t *testing.T) {
		t.Parallel()
		ctrl, _, _, mockGame, _ := setupControllerTest(t)
		expected := dto.GameView{State: "SETUP"}
		mockGame.EXPECT().PlaceShip(mock.Anything, "m1", "p1", 3, 0, 0, true).
			Return(expected, nil).Once()

		view, err := ctrl.PlaceShipAction(context.Background(), "m1", "p1", 3, 0, 0, true)
		assert.NoError(t, err)
		assert.Equal(t, expected, view)
	})

	t.Run("AttackAction", func(t *testing.T) {
		t.Parallel()
		ctrl, _, _, mockGame, _ := setupControllerTest(t)
		expected := dto.GameView{State: "PLAYING"}
		mockGame.EXPECT().Attack(mock.Anything, "m1", "p1", 5, 5).
			Return(expected, nil).Once()

		view, err := ctrl.AttackAction(context.Background(), "m1", "p1", 5, 5)
		assert.NoError(t, err)
		assert.Equal(t, expected, view)
	})

	t.Run("GetGameStateAction", func(t *testing.T) {
		t.Parallel()
		ctrl, _, _, mockGame, _ := setupControllerTest(t)
		expected := dto.GameView{State: "FINISHED"}
		mockGame.EXPECT().GetState(mock.Anything, "m1", "p1").
			Return(expected, nil).Once()

		view, err := ctrl.GetGameStateAction(context.Background(), "m1", "p1")
		assert.NoError(t, err)
		assert.Equal(t, expected, view)
	})
}
