package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequirePlayerID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupContext   func(c echo.Context)
		expectedStatus int
		expectedID     string
		expectError    bool
	}{
		{
			name: "Success - Valid Token",
			setupContext: func(c echo.Context) {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub": "player-123",
				})
				c.Set("user", token)
			},
			expectedStatus: http.StatusOK,
			expectedID:     "player-123",
			expectError:    false,
		},
		{
			name: "Failure - Missing Token",
			setupContext: func(c echo.Context) {
				// No token set in context
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name: "Failure - Invalid Token Type",
			setupContext: func(c echo.Context) {
				c.Set("user", "not-a-jwt-token")
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name: "Failure - Invalid Claims Type",
			setupContext: func(c echo.Context) {
				token := &jwt.Token{
					Claims: &jwt.RegisteredClaims{},
				}
				c.Set("user", token)
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name: "Failure - Missing Subject Claim",
			setupContext: func(c echo.Context) {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"iss": "battleship",
				})
				c.Set("user", token)
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name: "Failure - Empty Subject Claim",
			setupContext: func(c echo.Context) {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub": "",
				})
				c.Set("user", token)
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			// Mock next handler to verify context setting
			next := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			handler := RequirePlayerID(next)
			err := handler(c)

			if tt.expectError {
				require.Error(t, err)
				he := &echo.HTTPError{}
				ok := errors.As(err, &he)
				require.True(t, ok)
				assert.Equal(t, tt.expectedStatus, he.Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, c.Get("player_id"))
			}
		})
	}
}
