// Package main is the entry point of the server.
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/callegarimattia/battleship/internal/api"
	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	app := Application{}
	if err := app.Run(); err != nil {
		panic(err)
	}
}

type Application struct {
	E *echo.Echo
}

// Setup initializes the Echo instance and routes.
// It is separate from Run so that tests can initialize without starting the listener.
func (a *Application) Setup() {
	memEngine := service.NewMemoryService()
	authService := service.NewIdentityService()
	appCtrl := controller.NewAppController(authService, memEngine, memEngine)

	a.E = echo.New()

	// Middleware
	a.E.Use(middleware.RequestLogger())
	a.E.Use(middleware.Recover())
	a.E.Use(middleware.Secure())
	a.E.Use(middleware.CORS())
	a.E.Use(middleware.BodyLimit("1M"))
	a.E.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	h := api.NewEchoHandler(appCtrl)

	a.E.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	a.E.Static("/docs", "docs")

	a.E.POST("/login", h.Login)

	g := a.E.Group("/matches")
	g.GET("", h.ListMatches)
	g.POST("", h.HostMatch)

	g.POST("/:id/join", h.JoinMatch)
	g.GET("/:id", h.GetState)
	g.POST("/:id/place", h.PlaceShip)
	g.POST("/:id/attack", h.Attack)
}

// Run calls Setup and then starts the server.
func (a *Application) Run() error {
	a.Setup()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := &http.Server{
		Addr:              ":" + port,
		Handler:           a.E,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	return s.ListenAndServe()
}
