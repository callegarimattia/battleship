// Package api contains the http handlers
package api

import (
	"net/http"

	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/labstack/echo/v4"
)

// EchoHandler has the handlers for the http.Server
type EchoHandler struct{ ctrl *controller.AppController }

// NewEchoHandler creates a new http handler using echo
func NewEchoHandler(c *controller.AppController) *EchoHandler {
	return &EchoHandler{ctrl: c}
}

// Login handles the user login request.
// POST /login
func (h *EchoHandler) Login(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
	}

	user, err := h.ctrl.Login(c.Request().Context(), req.Username, "web", req.Username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

// ListMatches retrieves a list of all available matches.
// GET /matches
func (h *EchoHandler) ListMatches(c echo.Context) error {
	matches, err := h.ctrl.ListGamesAction(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, matches)
}

// HostMatch allows a player to host a new match.
// POST /matches
func (h *EchoHandler) HostMatch(c echo.Context) error {
	playerID := c.Get("player_id").(string)

	matchID, err := h.ctrl.HostGameAction(c.Request().Context(), playerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"match_id": matchID})
}

// JoinMatch allows a player to join an existing match.
// POST /matches/:id/join
func (h *EchoHandler) JoinMatch(c echo.Context) error {
	matchID := c.Param("id")
	playerID := c.Get("player_id").(string)

	view, err := h.ctrl.JoinGameAction(c.Request().Context(), matchID, playerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, view)
}

// GetState retrieves the current state of a match.
// GET /matches/:id
func (h *EchoHandler) GetState(c echo.Context) error {
	matchID := c.Param("id")
	playerID := c.Get("player_id").(string)

	view, err := h.ctrl.GetGameStateAction(c.Request().Context(), matchID, playerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, view)
}

// PlaceShip allows a player to place a ship on their board.
// POST /matches/:id/place
func (h *EchoHandler) PlaceShip(c echo.Context) error {
	var req struct {
		Size     int  `json:"size"`
		X        int  `json:"x"`
		Y        int  `json:"y"`
		Vertical bool `json:"vertical"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
	}

	matchID := c.Param("id")
	playerID := c.Get("player_id").(string)

	view, err := h.ctrl.PlaceShipAction(
		c.Request().Context(),
		matchID,
		playerID,
		req.Size,
		req.X,
		req.Y,
		req.Vertical,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, view)
}

// Attack allows a player to attack the opponent's board.
// POST /matches/:id/attack
func (h *EchoHandler) Attack(c echo.Context) error {
	var req struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
	}

	matchID := c.Param("id")
	playerID := c.Get("player_id").(string)

	view, err := h.ctrl.AttackAction(c.Request().Context(), matchID, playerID, req.X, req.Y)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, view)
}
