// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package server

import (
	"time"
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

// WithDataPath sets the path to the data directory.
func WithDataPath(path string) Option {
	return func(s *Server) error { s.dataPath = path; return nil }
}

// WithJWTSecret sets the HMAC secret for JWT signing.
func WithJWTSecret(secret string) Option {
	return func(s *Server) error { s.jwtSecret = secret; return nil }
}

// WithJWTTTL sets the JWT token time-to-live.
func WithJWTTTL(ttl time.Duration) Option {
	return func(s *Server) error { s.jwtTTL = ttl; return nil }
}
