package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/dto"
	mocks "github.com/callegarimattia/battleship/internal/mocks/controller"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Test Helpers ---

func setupTest(
	t *testing.T,
) (*echo.Echo, *EchoHandler, *mocks.MockIdentityService, *mocks.MockLobbyService, *mocks.MockGameService, *mocks.MockNotificationService) {
	e := echo.New()
	mockAuth := mocks.NewMockIdentityService(t)
	mockLobby := mocks.NewMockLobbyService(t)
	mockGame := mocks.NewMockGameService(t)
	mockNotifier := mocks.NewMockNotificationService(t)
	ctrl := controller.NewAppController(mockAuth, mockLobby, mockGame, mockNotifier)
	h := NewEchoHandler(ctrl)
	return e, h, mockAuth, mockLobby, mockGame, mockNotifier
}

func makeRequest(
	method, path string,
	body any,
	headers map[string]string,
) (*http.Request, *httptest.ResponseRecorder) {
	var bodyReader *bytes.Buffer
	if body != nil {
		if s, ok := body.(string); ok {
			bodyReader = bytes.NewBufferString(s)
		} else {
			jsonBytes, _ := json.Marshal(body)
			bodyReader = bytes.NewBuffer(jsonBytes)
		}
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	return req, rec
}

// --- Tests ---

func TestLogin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		reqBody        any
		mockSetup      func(*mocks.MockIdentityService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			reqBody: map[string]string{"username": "Alice"},
			mockSetup: func(m *mocks.MockIdentityService) {
				m.EXPECT().LoginOrRegister(mock.Anything, "Alice", "web", "Alice").
					Return(dto.AuthResponse{
						Token: "t1",
						User:  dto.User{ID: "user-123", Username: "Alice"},
					}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "user-123",
		},
		{
			name:           "Invalid JSON",
			reqBody:        "{invalid-json", // passing string directly
			mockSetup:      func(m *mocks.MockIdentityService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name:    "Service Error",
			reqBody: map[string]string{"username": "ErrorUser"},
			mockSetup: func(m *mocks.MockIdentityService) {
				m.EXPECT().LoginOrRegister(mock.Anything, "ErrorUser", "web", "ErrorUser").
					Return(dto.AuthResponse{}, errors.New("db down")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "db down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e, h, mockAuth, _, _, _ := setupTest(t)
			tt.mockSetup(mockAuth)

			req, rec := makeRequest(http.MethodPost, "/login", tt.reqBody, nil)
			c := e.NewContext(req, rec)

			err := h.Login(c)
			if err != nil {
				// Echo returns error for 4xx/5xx, so we need to check that too
				he := &echo.HTTPError{}
				ok := errors.As(err, &he)
				if assert.True(t, ok) {
					assert.Equal(t, tt.expectedStatus, he.Code)
					assert.Contains(t, he.Message, tt.expectedBody)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestListMatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockLobbyService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			mockSetup: func(m *mocks.MockLobbyService) {
				m.EXPECT().ListMatches(mock.Anything).
					Return([]dto.MatchSummary{
						{ID: "m1", HostName: "H1", PlayerCount: 1, CreatedAt: time.Now()},
					}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "m1",
		},
		{
			name: "Service Error",
			mockSetup: func(m *mocks.MockLobbyService) {
				m.EXPECT().ListMatches(mock.Anything).
					Return(nil, errors.New("db fail")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "db fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e, h, _, mockLobby, _, _ := setupTest(t)
			tt.mockSetup(mockLobby)

			req, rec := makeRequest(http.MethodGet, "/matches", nil, nil)
			c := e.NewContext(req, rec)

			err := h.ListMatches(c)
			if err != nil {
				he := &echo.HTTPError{}
				ok := errors.As(err, &he)
				if assert.True(t, ok) {
					assert.Equal(t, tt.expectedStatus, he.Code)
					assert.Contains(t, he.Message, tt.expectedBody)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestHostMatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		headers        map[string]string
		mockSetup      func(*mocks.MockLobbyService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			headers: map[string]string{"X-Player-ID": "user-123"},
			mockSetup: func(m *mocks.MockLobbyService) {
				m.EXPECT().CreateMatch(mock.Anything, "user-123").
					Return("match-new-id", nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "match-new-id",
		},
		{
			name:    "Service Error",
			headers: map[string]string{"X-Player-ID": "user-123"},
			mockSetup: func(m *mocks.MockLobbyService) {
				m.EXPECT().CreateMatch(mock.Anything, "user-123").
					Return("", errors.New("create fail")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "create fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e, h, _, mockLobby, _, _ := setupTest(t)
			tt.mockSetup(mockLobby)

			req, rec := makeRequest(http.MethodPost, "/matches", nil, tt.headers)
			c := e.NewContext(req, rec)
			if id := tt.headers["X-Player-ID"]; id != "" {
				c.Set("player_id", id)
			}

			err := h.HostMatch(c)
			if err != nil {
				he := &echo.HTTPError{}
				ok := errors.As(err, &he)
				if assert.True(t, ok) {
					assert.Equal(t, tt.expectedStatus, he.Code)
					assert.Contains(t, he.Message, tt.expectedBody)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestJoinMatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		headers        map[string]string
		paramID        string
		mockSetup      func(*mocks.MockLobbyService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			headers: map[string]string{"X-Player-ID": "p2"},
			paramID: "m1",
			mockSetup: func(m *mocks.MockLobbyService) {
				m.EXPECT().JoinMatch(mock.Anything, "m1", "p2").
					Return(dto.GameView{State: "SETUP"}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "SETUP",
		},
		{
			name:    "Service Error",
			headers: map[string]string{"X-Player-ID": "p2"},
			paramID: "m1",
			mockSetup: func(m *mocks.MockLobbyService) {
				m.EXPECT().JoinMatch(mock.Anything, "m1", "p2").
					Return(dto.GameView{}, errors.New("game full")).
					Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "game full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e, h, _, mockLobby, _, _ := setupTest(t)
			tt.mockSetup(mockLobby)

			req, rec := makeRequest(
				http.MethodPost,
				"/matches/"+tt.paramID+"/join",
				nil,
				tt.headers,
			)
			c := e.NewContext(req, rec)
			if id := tt.headers["X-Player-ID"]; id != "" {
				c.Set("player_id", id)
			}
			c.SetParamNames("id")
			c.SetParamValues(tt.paramID)

			err := h.JoinMatch(c)
			if err != nil {
				he := &echo.HTTPError{}
				ok := errors.As(err, &he)
				if assert.True(t, ok) {
					assert.Equal(t, tt.expectedStatus, he.Code)
					assert.Contains(t, he.Message, tt.expectedBody)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestGetState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		headers        map[string]string
		paramID        string
		mockSetup      func(*mocks.MockGameService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			headers: map[string]string{"X-Player-ID": "p1"},
			paramID: "m1",
			mockSetup: func(m *mocks.MockGameService) {
				m.EXPECT().GetState(mock.Anything, "m1", "p1").
					Return(dto.GameView{State: "PLAYING"}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "PLAYING",
		},
		{
			name:    "Service Error",
			headers: map[string]string{"X-Player-ID": "p1"},
			paramID: "m1",
			mockSetup: func(m *mocks.MockGameService) {
				m.EXPECT().GetState(mock.Anything, "m1", "p1").
					Return(dto.GameView{}, errors.New("not found")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e, h, _, _, mockGame, _ := setupTest(t)
			tt.mockSetup(mockGame)

			req, rec := makeRequest(http.MethodGet, "/matches/"+tt.paramID, nil, tt.headers)
			c := e.NewContext(req, rec)
			if id := tt.headers["X-Player-ID"]; id != "" {
				c.Set("player_id", id)
			}
			c.SetParamNames("id")
			c.SetParamValues(tt.paramID)

			err := h.GetState(c)
			if err != nil {
				he := &echo.HTTPError{}
				ok := errors.As(err, &he)
				if assert.True(t, ok) {
					assert.Equal(t, tt.expectedStatus, he.Code)
					assert.Contains(t, he.Message, tt.expectedBody)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestPlaceShip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		headers        map[string]string
		reqBody        any
		mockSetup      func(*mocks.MockGameService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Success",
			headers: map[string]string{"X-Player-ID": "p1"},
			reqBody: map[string]any{"size": 3, "x": 0, "y": 0, "vertical": true},
			mockSetup: func(m *mocks.MockGameService) {
				m.EXPECT().PlaceShip(mock.Anything, "m1", "p1", 3, 0, 0, true).
					Return(dto.GameView{State: "SETUP"}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "SETUP",
		},
		{
			name:           "Invalid JSON",
			headers:        map[string]string{"X-Player-ID": "p1"},
			reqBody:        "{bad-json",
			mockSetup:      func(m *mocks.MockGameService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name:    "Service Error",
			headers: map[string]string{"X-Player-ID": "p1"},
			reqBody: map[string]any{"size": 3, "x": 0, "y": 0, "vertical": true},
			mockSetup: func(m *mocks.MockGameService) {
				m.EXPECT().PlaceShip(mock.Anything, "m1", "p1", 3, 0, 0, true).
					Return(dto.GameView{}, errors.New("overlap")).
					Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "overlap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e, h, _, _, mockGame, _ := setupTest(t)
			tt.mockSetup(mockGame)

			req, rec := makeRequest(http.MethodPost, "/matches/m1/place", tt.reqBody, tt.headers)
			c := e.NewContext(req, rec)
			if id := tt.headers["X-Player-ID"]; id != "" {
				c.Set("player_id", id)
			}
			c.SetParamNames("id")
			c.SetParamValues("m1")

			err := h.PlaceShip(c)
			if err != nil {
				he := &echo.HTTPError{}
				ok := errors.As(err, &he)
				if assert.True(t, ok) {
					assert.Equal(t, tt.expectedStatus, he.Code)
					assert.Contains(t, he.Message, tt.expectedBody)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestAttack(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		headers        map[string]string
		reqBody        any
		mockSetup      func(*mocks.MockGameService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Hit",
			headers: map[string]string{"X-Player-ID": "p1"},
			reqBody: map[string]any{"x": 5, "y": 5},
			mockSetup: func(m *mocks.MockGameService) {
				m.EXPECT().Attack(mock.Anything, "m1", "p1", 5, 5).
					Return(dto.GameView{State: "playing", Turn: "p2"}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "playing",
		},
		{
			name:           "Invalid JSON",
			headers:        map[string]string{"X-Player-ID": "p1"},
			reqBody:        "{bad",
			mockSetup:      func(m *mocks.MockGameService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
		{
			name:    "Service Error",
			headers: map[string]string{"X-Player-ID": "p1"},
			reqBody: map[string]any{"x": 5, "y": 5},
			mockSetup: func(m *mocks.MockGameService) {
				m.EXPECT().Attack(mock.Anything, "m1", "p1", 5, 5).
					Return(dto.GameView{}, errors.New("not your turn")).
					Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "not your turn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e, h, _, _, mockGame, _ := setupTest(t)
			tt.mockSetup(mockGame)

			req, rec := makeRequest(http.MethodPost, "/matches/m1/attack", tt.reqBody, tt.headers)
			c := e.NewContext(req, rec)
			if id := tt.headers["X-Player-ID"]; id != "" {
				c.Set("player_id", id)
			}
			c.SetParamNames("id")
			c.SetParamValues("m1")

			err := h.Attack(c)
			if err != nil {
				he := &echo.HTTPError{}
				ok := errors.As(err, &he)
				if assert.True(t, ok) {
					assert.Equal(t, tt.expectedStatus, he.Code)
					assert.Contains(t, he.Message, tt.expectedBody)
				}
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestStreamMatchEvents(t *testing.T) { //nolint:paralleltest
	e, h, _, _, mockGame, mockNotifier := setupTest(t)

	mockSub := mocks.NewMockSubscription(t)
	mockSub.EXPECT().Unsubscribe().Return().Maybe()

	eventChan := make(chan *dto.GameEvent, 1)

	mockNotifier.EXPECT().Subscribe("m1").
		Return(mockSub, (<-chan *dto.GameEvent)(eventChan)).
		Once()

	initialView := dto.GameView{State: "WAITING", Turn: "p1"}
	mockGame.EXPECT().GetState(mock.Anything, "m1", "p1").
		Return(initialView, nil).
		Once()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := e.NewContext(r, w)
		c.SetPath("/matches/:id/ws")
		c.SetParamNames("id")
		c.SetParamValues("m1")
		c.Set("player_id", "p1")

		err := h.StreamMatchEvents(c)
		assert.NoError(t, err)
	}))
	defer ts.Close()

	wsURL := "ws" + ts.URL[4:] + "/matches/m1/ws"

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	var evt dto.WSEvent
	err = ws.ReadJSON(&evt)
	assert.NoError(t, err)
	assert.Equal(t, "game_update", evt.Type)
	assert.NotNil(t, evt.Payload)
	assert.Equal(t, dto.GameState("WAITING"), evt.Payload.State)

	// Updated view expectations? Maybe redundant if we don't call GetState again
	// Actually StreamMatchEvents fetches fresh state in the loop.

	updatedView := dto.GameView{State: "PLAYING", Turn: "p2"}
	mockGame.EXPECT().GetState(mock.Anything, "m1", "p1").
		Return(updatedView, nil).
		Maybe()

	eventChan <- &dto.GameEvent{Type: dto.EventGameStarted}

	err = ws.ReadJSON(&evt)
	assert.NoError(t, err)
	assert.Equal(t, "game_update", evt.Type)
	assert.NotNil(t, evt.Payload)
	assert.Equal(t, dto.GameState("PLAYING"), evt.Payload.State)
}
