// Package client provides an HTTP and WebSocket client for the Battleship game.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/callegarimattia/battleship/internal/dto"
	"github.com/gorilla/websocket"
)

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 5 * time.Second},
	}
}

// Helper for authorized requests
func (c *Client) do(method, path string, body, dest any) error {
	var bodyReader *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API Error: %d", resp.StatusCode)
	}

	if dest != nil {
		return json.NewDecoder(resp.Body).Decode(dest)
	}

	return resp.Body.Close()
}

// --- Auth ---

func (c *Client) Login(username string) (*dto.AuthResponse, error) {
	req := map[string]string{"username": username}
	var res dto.AuthResponse
	err := c.do("POST", "/login", req, &res)
	if err == nil {
		c.Token = res.Token // Store token automatically
	}
	return &res, err
}

// --- Lobby ---

func (c *Client) ListMatches() ([]dto.MatchSummary, error) {
	var matches []dto.MatchSummary
	err := c.do("GET", "/matches", nil, &matches)
	return matches, err
}

func (c *Client) CreateMatch() (string, error) {
	var res struct {
		MatchID string `json:"match_id"`
	}
	err := c.do("POST", "/matches", nil, &res)
	return res.MatchID, err
}

func (c *Client) JoinMatch(matchID string) (*dto.GameView, error) {
	var game dto.GameView
	err := c.do("POST", fmt.Sprintf("/matches/%s/join", matchID), nil, &game)
	return &game, err
}

// --- Game ---

func (c *Client) GetGameState(matchID string) (*dto.GameView, error) {
	var game dto.GameView
	err := c.do("GET", fmt.Sprintf("/matches/%s", matchID), nil, &game)
	return &game, err
}

func (c *Client) PlaceShip(matchID string, size, x, y int, vertical bool) (*dto.GameView, error) {
	var game dto.GameView
	req := map[string]any{
		"size":     size,
		"x":        x,
		"y":        y,
		"vertical": vertical,
	}
	err := c.do("POST", fmt.Sprintf("/matches/%s/place", matchID), req, &game)
	return &game, err
}

func (c *Client) Attack(matchID string, x, y int) (*dto.GameView, error) {
	var game dto.GameView
	req := map[string]any{
		"x": x,
		"y": y,
	}
	err := c.do("POST", fmt.Sprintf("/matches/%s/attack", matchID), req, &game)
	return &game, err
}

// SubscribeToMatch connects to the WebSocket endpoint and returns a channel that signals updates.
// SubscribeToMatch connects to the WebSocket endpoint and returns a channel that receives game state updates.
func (c *Client) SubscribeToMatch(matchID string) (<-chan *dto.WSEvent, error) {
	// Determine WS scheme
	scheme := "ws"
	if strings.HasPrefix(c.BaseURL, "https") {
		scheme = "wss"
	}

	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	u.Scheme = scheme
	u.Path = fmt.Sprintf("/matches/%s/ws", matchID)

	header := http.Header{}
	if c.Token != "" {
		header.Set("Authorization", "Bearer "+c.Token)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}

	updateChan := make(chan *dto.WSEvent, 1)

	// Pump
	go func() {
		defer func() { _ = conn.Close() }()
		defer close(updateChan)
		for {
			var evt dto.WSEvent
			if err := conn.ReadJSON(&evt); err != nil {
				return
			}
			// Signal update
			select {
			case updateChan <- &evt:
			default:
			}
		}
	}()

	return updateChan, nil
}
