// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"

	"github.com/mdhender/ec/internal/cerr"
)

// Claims holds the JWT registered claims for EC tokens.
type Claims struct {
	jwt.RegisteredClaims
}

// JWTManager issues and validates JWT tokens using HMAC-SHA256.
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTManager creates a new JWT manager with the given HMAC secret and token TTL.
func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), ttl: ttl}
}

// Issue creates and signs a JWT for the given empire number.
func (m *JWTManager) Issue(empireNo int) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ec",
			Subject:   strconv.Itoa(empireNo),
			Audience:  jwt.ClaimStrings{"ec-web"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Validate parses and validates a JWT, returning the empire number from the sub claim.
// Returns cerr.ErrInvalidToken on any validation failure.
func (m *JWTManager) Validate(tokenStr string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, cerr.ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithAudience("ec-web"), jwt.WithIssuer("ec"))
	if err != nil || !token.Valid {
		return 0, cerr.ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return 0, cerr.ErrInvalidToken
	}
	empireNo, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, cerr.ErrInvalidToken
	}
	return empireNo, nil
}

// Middleware returns an Echo JWT middleware configured for this manager.
func (m *JWTManager) Middleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c *echo.Context) jwt.Claims { return &Claims{} },
		SigningKey:    m.secret,
	})
}

// FromContext extracts the empire number from the validated JWT stored in an Echo context.
func FromContext(c *echo.Context) (empireNo int, ok bool) {
	token, ok2 := c.Get("user").(*jwt.Token)
	if !ok2 || token == nil {
		return 0, false
	}
	claims, ok3 := token.Claims.(*Claims)
	if !ok3 {
		return 0, false
	}
	n, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, false
	}
	return n, true
}
