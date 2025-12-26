package server

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// RequirePlayerID extracts the user ID from the JWT and validates it.
// It sets "player_id" in the context.
func RequirePlayerID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing token")
		}

		claims, ok := user.Claims.(jwt.MapClaims)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
		}

		id, ok := claims["sub"].(string)
		if !ok || id == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID in token")
		}

		// Inject ID into context for handlers
		c.Set("player_id", id)

		return next(c)
	}
}
