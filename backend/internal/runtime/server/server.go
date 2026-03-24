// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/mdhender/ec/internal/app"
	deliveryhttp "github.com/mdhender/ec/internal/delivery/http"
)

// Server manages the EC API HTTP server lifecycle.
type Server struct {
	host          string
	port          string
	shutdownAfter time.Duration
	shutdownKey   string
	jwtMiddleware echo.MiddlewareFunc
	authStore     app.AuthStore
	tokenSigner   app.TokenSigner
	orderStore    app.OrderStore
	reportStore   app.ReportStore
	shutdownCh    chan struct{}
	once          sync.Once
}

// New creates a Server by applying the given options.
// Returns an error if required stores are missing.
func New(opts ...Option) (*Server, error) {
	s := &Server{
		host:       "localhost",
		port:       "8080",
		shutdownCh: make(chan struct{}, 1),
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	if s.authStore == nil {
		return nil, fmt.Errorf("server: authStore is required")
	}
	if s.tokenSigner == nil {
		return nil, fmt.Errorf("server: tokenSigner is required")
	}
	if s.orderStore == nil {
		return nil, fmt.Errorf("server: orderStore is required")
	}
	if s.reportStore == nil {
		return nil, fmt.Errorf("server: reportStore is required")
	}
	return s, nil
}

// Start wires routes, starts the HTTP server, and blocks until shutdown.
func (s *Server) Start() error {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	deliveryhttp.AddRoutes(
		e,
		s.jwtMiddleware,
		s.authStore,
		s.tokenSigner,
		s.orderStore,
		s.reportStore,
		s.shutdownKey,
		s.shutdownCh,
	)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	srv := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	// Start listener in background.
	listenErr := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			listenErr <- err
		}
		close(listenErr)
	}()

	// Signal handling.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Optional auto-shutdown timer.
	var timer <-chan time.Time
	if s.shutdownAfter > 0 {
		timer = time.After(s.shutdownAfter)
	}

	// Wait for a shutdown trigger.
	select {
	case err := <-listenErr:
		if err != nil {
			return fmt.Errorf("server listen: %w", err)
		}
		return nil
	case sig := <-sigCh:
		slog.Info("signal received", "signal", sig)
	case <-s.shutdownCh:
		slog.Info("shutdown requested via API")
	case <-timer:
		slog.Info("auto-shutdown timer fired")
	}

	// Graceful shutdown with a 5-second grace period.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	slog.Info("server stopped")
	return nil
}

// Shutdown triggers a graceful shutdown via the internal channel.
func (s *Server) Shutdown() {
	s.once.Do(func() {
		select {
		case s.shutdownCh <- struct{}{}:
		default:
		}
	})
}
