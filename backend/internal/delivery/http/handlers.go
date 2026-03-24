// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package http

import (
	"crypto/subtle"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/cerr"
	"github.com/mdhender/ec/internal/infra/auth"
)

// Todo returns a 501 Not Implemented handler.
func Todo() func(c *echo.Context) error {
	return func(c *echo.Context) error {
		return c.JSON(http.StatusNotImplemented, map[string]any{"error": "not implemented"})
	}
}

// GetHealth returns a 200 JSON response with ok and current UTC time.
func GetHealth() func(c *echo.Context) error {
	return func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"ok":   true,
			"time": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// PostLogin validates a magic link and issues a JWT token.
func PostLogin(authStore app.AuthStore, tokenSigner app.TokenSigner) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		magicLink := c.Param("magicLink")
		ctx := c.Request().Context()

		empireNo, ok, err := authStore.ValidateMagicLink(ctx, magicLink)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		}

		token, err := tokenSigner.Issue(empireNo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"access_token": token,
			"token_type":   "Bearer",
		})
	}
}

// PostLogout returns a 200 ok response.
func PostLogout() func(c *echo.Context) error {
	return func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

// GetOrders returns the orders for an empire (JWT-protected, empire must match token).
func GetOrders(orderStore app.OrderStore) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		empireNo, err := strconv.Atoi(c.Param("empireNo"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid empire number"})
		}

		jwtEmpireNo, ok := auth.FromContext(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		}

		if empireNo != jwtEmpireNo {
			return c.JSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
		}

		ctx := c.Request().Context()
		body, err := orderStore.GetOrders(ctx, empireNo)
		if err != nil {
			if errors.Is(err, cerr.ErrNotFound) {
				return c.JSON(http.StatusNotFound, map[string]any{"error": "not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		return c.String(http.StatusOK, body)
	}
}

// PostOrders stores the orders for an empire (JWT-protected, empire must match token).
func PostOrders(orderStore app.OrderStore) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		empireNo, err := strconv.Atoi(c.Param("empireNo"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid empire number"})
		}

		jwtEmpireNo, ok := auth.FromContext(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		}

		if empireNo != jwtEmpireNo {
			return c.JSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
		}

		rawBody, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		ctx := c.Request().Context()
		if err := orderStore.PutOrders(ctx, empireNo, string(rawBody)); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		return c.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

// GetReports returns the list of reports for an empire (JWT-protected, empire must match token).
func GetReports(reportStore app.ReportStore) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		empireNo, err := strconv.Atoi(c.Param("empireNo"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid empire number"})
		}

		jwtEmpireNo, ok := auth.FromContext(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		}

		if empireNo != jwtEmpireNo {
			return c.JSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
		}

		ctx := c.Request().Context()
		reports, err := reportStore.ListReports(ctx, empireNo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		return c.JSON(http.StatusOK, reports)
	}
}

// GetReport returns a specific turn report for an empire (JWT-protected, empire must match token).
func GetReport(reportStore app.ReportStore) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		empireNo, err := strconv.Atoi(c.Param("empireNo"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid empire number"})
		}

		jwtEmpireNo, ok := auth.FromContext(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		}

		if empireNo != jwtEmpireNo {
			return c.JSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
		}

		turnYear, err := strconv.Atoi(c.Param("turnYear"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid turn year"})
		}

		turnQuarter, err := strconv.Atoi(c.Param("turnQuarter"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid turn quarter"})
		}

		ctx := c.Request().Context()
		data, err := reportStore.GetReport(ctx, empireNo, turnYear, turnQuarter)
		if err != nil {
			if errors.Is(err, cerr.ErrNotFound) {
				return c.JSON(http.StatusNotFound, map[string]any{"error": "not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		return c.JSONBlob(http.StatusOK, data)
	}
}

// PostShutdown triggers a graceful shutdown using a shared secret key.
// If key is empty, always returns 501.
func PostShutdown(key string, shutdownCh chan struct{}) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		if key == "" {
			return c.JSON(http.StatusNotImplemented, map[string]any{"error": "not implemented"})
		}

		provided := c.Param("key")
		if subtle.ConstantTimeCompare([]byte(key), []byte(provided)) == 1 {
			select {
			case shutdownCh <- struct{}{}:
			default:
			}
			return c.JSON(http.StatusOK, map[string]any{"ok": true})
		}

		return c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
	}
}
