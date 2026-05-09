package server

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"charm.land/log/v2"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Config holds the server configuration.
type Config struct {
	Host         string
	Port         int
	Instructions string
	APIKey       string
}

// Run starts the OpenAI-compatible API server.
func Run(ctx context.Context, cfg Config) error {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:   true,
		LogURIPath:  true,
		LogStatus:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			log.Info("Request",
				"method", v.Method,
				"path", v.URIPath,
				"status", v.Status,
				"latency", v.Latency,
				"ip", v.RemoteIP,
			)
			return nil
		},
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	if cfg.APIKey != "" {
		e.Use(apiKeyAuth(cfg.APIKey))
	}

	h := &handler{
		instructions: cfg.Instructions,
	}

	v1 := e.Group("/v1")
	v1.GET("/models", h.listModels)
	v1.POST("/chat/completions", h.chatCompletions)

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Info("Starting server", "addr", addr)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", "err", err)
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down server")
	return srv.Close()
}
