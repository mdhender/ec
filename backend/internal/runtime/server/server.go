// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/mdhender/ec/internal/app"
	deliveryhttp "github.com/mdhender/ec/internal/delivery/http"
	"github.com/mdhender/ec/internal/infra/auth"
	"github.com/mdhender/ec/internal/infra/filestore"
	"github.com/mdhender/ec/internal/infra/ordertext"
)

const maxOrderBytes int64 = 1 << 20 // 1 MiB

// Server manages the EC API HTTP server lifecycle.
type Server struct {
	host          string
	port          string
	shutdownAfter time.Duration
	shutdownKey   string
	dataPath      string
	jwtSecret     string
	jwtTTL        time.Duration
	shutdownCh    chan struct{}
	once          sync.Once
}

// New creates a Server by applying the given options.
// Returns an error if required configuration is missing.
func New(opts ...Option) (*Server, error) {
	s := &Server{
		host:       "localhost",
		port:       "8080",
		jwtTTL:     24 * time.Hour,
		shutdownCh: make(chan struct{}, 1),
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	if s.dataPath == "" {
		return nil, fmt.Errorf("server: dataPath is required")
	}
	if s.jwtSecret == "" {
		return nil, fmt.Errorf("server: jwtSecret is required")
	}
	return s, nil
}

// Start builds infra adapters, wires routes, starts the HTTP server, and blocks until shutdown.
func (s *Server) Start() error {
	// Build infra adapters (runtime owns concrete instantiation).
	authStore, err := auth.NewMagicLinkStore(filepath.Join(s.dataPath, "auth.json"))
	if err != nil {
		return fmt.Errorf("server: load auth: %w", err)
	}

	jwtMgr := auth.NewJWTManager(s.jwtSecret, s.jwtTTL)
	fileStore := filestore.NewStore(s.dataPath)

	// Build app-layer services.
	loginSvc := &app.LoginService{
		Auth:  authStore,
		Token: jwtMgr,
	}

	// Build parse service.
	orderParser := ordertext.NewParser()
	parseOrdersSvc := &app.ParseOrdersService{Parser: orderParser}

	// Empire extractor bridges infra/auth into delivery without a direct import.
	empireExtractor := func(c *echo.Context) (int, bool) {
		return auth.FromContext(c)
	}

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	tokenValidator := func(token string) (int, error) {
		return jwtMgr.Validate(token)
	}

	deliveryhttp.AddRoutes(
		e,
		jwtMgr.Middleware(),
		empireExtractor,
		tokenValidator,
		loginSvc,
		fileStore,        // orderStore
		fileStore,        // reportStore
		fileStore,        // dashboardStore
		s.shutdownKey,
		s.shutdownCh,
		maxOrderBytes,
		parseOrdersSvc,
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
