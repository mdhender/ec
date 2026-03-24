// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package auth implements JWT-based authentication for the EC API server.
package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/mdhender/ec/internal/cerr"
)

// JWTManager issues and validates JWT tokens using HMAC-SHA256.
// It implements app.TokenSigner.
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

// Claims holds the registered JWT claims for EC tokens.
type Claims struct {
	jwt.RegisteredClaims
}

// NewJWTManager returns a JWTManager signing with the given secret and TTL.
func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// Issue creates a signed JWT for the given empire number.
func (m *JWTManager) Issue(empireNo int) (string, error) {
	now := time.Now().UTC()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(empireNo),
			Issuer:    "ec",
			Audience:  jwt.ClaimStrings{"ec-web"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

// Validate parses and validates the token, returning the empire number from the sub claim.
// Returns cerr.ErrInvalidToken on any validation failure.
func (m *JWTManager) Validate(token string) (int, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, cerr.ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !t.Valid {
		return 0, cerr.ErrInvalidToken
	}

	claims, ok := t.Claims.(*Claims)
	if !ok {
		return 0, cerr.ErrInvalidToken
	}

	empireNo, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, cerr.ErrInvalidToken
	}

	return empireNo, nil
}

// Middleware returns Echo middleware that validates JWTs issued by this manager.
// Validated tokens are stored in the Echo context under key "user".
func (m *JWTManager) Middleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c *echo.Context) jwt.Claims {
			return new(Claims)
		},
		SigningKey:    m.secret,
		SigningMethod: "HS256",
		ContextKey:    "user",
	})
}

// FromContext extracts the empire number from the JWT stored in the Echo context.
// Returns 0, false if extraction fails.
func FromContext(c *echo.Context) (empireNo int, ok bool) {
	raw := c.Get("user")
	if raw == nil {
		return 0, false
	}
	t, ok := raw.(*jwt.Token)
	if !ok {
		return 0, false
	}
	claims, ok := t.Claims.(*Claims)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, false
	}
	return n, true
}
