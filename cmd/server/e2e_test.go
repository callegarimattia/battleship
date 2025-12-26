package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestE2E_FullGameScenario(t *testing.T) {
	// Disable rate limiting for E2E tests
	os.Setenv("RATE_LIMIT", "1000")
	defer os.Unsetenv("RATE_LIMIT")

	t.Parallel()

	app := &Application{}
	app.Setup()

	// Use a real HTTP server
	ts := httptest.NewServer(app.E)
	defer ts.Close()

	// 1. Players Login
	aliceClient := &testClient{t: t, baseURL: ts.URL, client: ts.Client()}
	alice := aliceClient.login("Alice")

	bobClient := &testClient{t: t, baseURL: ts.URL, client: ts.Client()}
	_ = bobClient.login("Bob")

	// 2. Host and Join Match
	matchID := aliceClient.createMatch()
	bobClient.joinMatch(matchID)

	// 3. Place Ships
	// Configuration: Size 5, 4, 3, 3, 2. All horizontal for simplicity.
	// Player 1 (Alice) placement
	aliceClient.placeShip(matchID, 5, 0, 0, false)
	aliceClient.placeShip(matchID, 4, 0, 1, false)
	aliceClient.placeShip(matchID, 3, 0, 2, false)
	aliceClient.placeShip(matchID, 3, 0, 3, false)
	aliceClient.placeShip(matchID, 2, 0, 4, false)

	// Player 2 (Bob) placement
	bobClient.placeShip(matchID, 5, 0, 0, false)
	bobClient.placeShip(matchID, 4, 0, 1, false)
	bobClient.placeShip(matchID, 3, 0, 2, false)
	bobClient.placeShip(matchID, 3, 0, 3, false)
	bobClient.placeShip(matchID, 2, 0, 4, false)

	// 4. Verify Game Started
	state := aliceClient.getMatchState(matchID)
	require.Equal(t, dto.StatePlaying, state.State)
	require.Equal(t, alice.ID, state.Turn, "Alice should start")

	// 5. Game Loop: Alice destroys Bob's fleet
	// Ships are at Y=0..4, X=0..(Size-1)
	targets := []struct{ x, y int }{
		{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, // Size 5
		{0, 1}, {1, 1}, {2, 1}, {3, 1}, // Size 4
		{0, 2}, {1, 2}, {2, 2}, // Size 3
		{0, 3}, {1, 3}, {2, 3}, // Size 3
		{0, 4}, {1, 4}, // Size 2
	}

	for i, target := range targets {
		// Alice attacks
		state = aliceClient.attack(matchID, target.x, target.y)

		if state.State == dto.StateFinished {
			break
		}

		// Bob misses (always attacks unique empty spots)
		// We use the loop index to generate unique coordinates (9-row, col)
		// ensuring we don't hit the same spot twice.
		bobClient.attack(matchID, 9-(i/10), i%10)
	}

	// 6. Verify Game Over
	finalState := aliceClient.getMatchState(matchID)
	require.Equal(t, dto.StateFinished, finalState.State)
	require.Equal(t, alice.ID, finalState.Winner)
}

// --- Test Helper ---

type testClient struct {
	t       *testing.T
	baseURL string
	client  *http.Client
	token   string
}

type testResponse struct {
	Code int
	Body *bytes.Buffer
}

func (c *testClient) do(
	method, path string,
	body interface{},
	headers map[string]string, //nolint:unparam
) *testResponse {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(c.t, err, "failed to marshal request body")
		reqBody = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	require.NoError(c.t, err, "failed to create request")

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if c.token != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+c.token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	require.NoError(c.t, err, "failed to execute request")
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(c.t, err, "failed to read response body")

	return &testResponse{
		Code: resp.StatusCode,
		Body: bytes.NewBuffer(respBody),
	}
}

func (c *testClient) login(username string) dto.User {
	rec := c.do(http.MethodPost, "/login", map[string]string{"username": username}, nil)
	require.Equal(c.t, http.StatusOK, rec.Code)

	var resp dto.AuthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(c.t, err)

	c.token = resp.Token
	return resp.User
}

func (c *testClient) createMatch() string {
	rec := c.do(http.MethodPost, "/matches", nil, nil)
	require.Equal(c.t, http.StatusOK, rec.Code)

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(c.t, err)
	return resp["match_id"]
}

func (c *testClient) joinMatch(matchID string) {
	rec := c.do(
		http.MethodPost,
		"/matches/"+matchID+"/join",
		nil,
		nil,
	)
	require.Equal(c.t, http.StatusOK, rec.Code)
}

func (c *testClient) placeShip(
	matchID string,
	size, x, y int, //nolint:unparam
	vertical bool, //nolint:unparam
) {
	payload := map[string]interface{}{
		"size":     size,
		"x":        x,
		"y":        y,
		"vertical": vertical,
	}
	rec := c.do(
		http.MethodPost,
		"/matches/"+matchID+"/place",
		payload,
		nil,
	)
	require.Equal(
		c.t,
		http.StatusOK,
		rec.Code,
		fmt.Sprintf("placeShip failed for size %d at %d,%d", size, x, y),
	)
}

func (c *testClient) getMatchState(matchID string) dto.GameView {
	rec := c.do(
		http.MethodGet,
		"/matches/"+matchID,
		nil,
		nil,
	)
	require.Equal(c.t, http.StatusOK, rec.Code)

	var state dto.GameView
	err := json.Unmarshal(rec.Body.Bytes(), &state)
	require.NoError(c.t, err)
	return state
}

func (c *testClient) attack(matchID string, x, y int) dto.GameView {
	payload := map[string]interface{}{"x": x, "y": y}
	rec := c.do(
		http.MethodPost,
		"/matches/"+matchID+"/attack",
		payload,
		nil,
	)
	require.Equal(c.t, http.StatusOK, rec.Code, fmt.Sprintf("attack failed at %d,%d", x, y))

	var state dto.GameView
	err := json.Unmarshal(rec.Body.Bytes(), &state)
	require.NoError(c.t, err)
	return state
}
