// Package main is the entry point of the server.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/callegarimattia/battleship/internal/api"
	"github.com/callegarimattia/battleship/internal/controller"
	"github.com/callegarimattia/battleship/internal/env"
	"github.com/callegarimattia/battleship/internal/events"
	"github.com/callegarimattia/battleship/internal/service"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
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
	cfg, err := env.LoadServerConfig()
	if err != nil {
		log.Fatalf("Failed to load server config: %v", err)
	}

	// Initialize event bus
	eventBus := events.NewMemoryEventBus()
	// Note: defer eventBus.Close() is typically used when the bus has resources to clean up
	// and the function's scope is the lifetime of the bus. For a server, the bus might
	// need to live for the entire application lifecycle, so deferring here might close it
	// prematurely if Setup is called and then the server runs.
	// For now, keeping it as per instruction, but it might need adjustment based on bus implementation.

	// Initialize services, passing the event bus
	memEngine := service.NewMemoryService(eventBus) // Modified to pass eventBus
	authService := service.NewIdentityService(cfg.JWTSecret)
	appCtrl := controller.NewAppController(authService, memEngine, memEngine)

	a.E = echo.New()

	// Middleware
	a.E.Use(middleware.RequestLogger())
	a.E.Use(middleware.Recover())
	a.E.Use(middleware.Secure())
	a.E.Use(middleware.CORS())
	a.E.Use(middleware.BodyLimit("1M"))
	a.E.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(cfg.RateLimit))))

	h := api.NewEchoHandler(appCtrl)

	a.E.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	a.E.Static("/docs", "docs")

	a.E.POST("/login", h.Login)

	g := a.E.Group("/matches")
	g.GET("", h.ListMatches)

	// Protected routes
	protected := g.Group("")
	protected.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(cfg.JWTSecret),
	}))
	protected.Use(api.RequirePlayerID)

	protected.POST("", h.HostMatch)
	protected.POST("/:id/join", h.JoinMatch)
	protected.GET("/:id", h.GetState)
	protected.POST("/:id/place", h.PlaceShip)
	protected.POST("/:id/attack", h.Attack)
}

// Run calls Setup and then starts the server.
func (a *Application) Run() error {
	a.Setup()

	cfg, err := env.LoadServerConfig()
	if err != nil {
		log.Fatalf("Failed to load server config: %v", err)
	}

	s := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           a.E,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	return s.ListenAndServe()
}
