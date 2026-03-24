// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package server

import (
	"time"

	"github.com/labstack/echo/v5"

	"github.com/mdhender/ec/internal/app"
)

// Option is a functional option for configuring a Server.
type Option func(*Server) error

// WithHost sets the server's listen host.
func WithHost(host string) Option {
	return func(s *Server) error { s.host = host; return nil }
}

// WithPort sets the server's listen port.
func WithPort(port string) Option {
	return func(s *Server) error { s.port = port; return nil }
}

// WithShutdownAfter sets an automatic shutdown timer (0 disables it).
func WithShutdownAfter(d time.Duration) Option {
	return func(s *Server) error { s.shutdownAfter = d; return nil }
}

// WithShutdownKey sets the secret key for the HTTP shutdown endpoint.
func WithShutdownKey(key string) Option {
	return func(s *Server) error { s.shutdownKey = key; return nil }
}

// WithJWTMiddleware sets the JWT middleware function.
func WithJWTMiddleware(mw echo.MiddlewareFunc) Option {
	return func(s *Server) error { s.jwtMiddleware = mw; return nil }
}

// WithAuthStore sets the authentication store.
func WithAuthStore(store app.AuthStore) Option {
	return func(s *Server) error { s.authStore = store; return nil }
}

// WithTokenSigner sets the JWT token signer.
func WithTokenSigner(signer app.TokenSigner) Option {
	return func(s *Server) error { s.tokenSigner = signer; return nil }
}

// WithOrderStore sets the order store.
func WithOrderStore(store app.OrderStore) Option {
	return func(s *Server) error { s.orderStore = store; return nil }
}

// WithReportStore sets the report store.
func WithReportStore(store app.ReportStore) Option {
	return func(s *Server) error { s.reportStore = store; return nil }
}
