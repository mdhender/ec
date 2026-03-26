// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package http

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/cerr"
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

// PostLogin validates a magic link and issues a JWT token via the LoginService.
func PostLogin(loginSvc *app.LoginService) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		magicLink := c.Param("magicLink")
		ctx := c.Request().Context()

		token, err := loginSvc.Login(ctx, magicLink)
		if err != nil {
			if errors.Is(err, cerr.ErrInvalidMagicLink) {
				return c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"access_token": token,
			"token_type":   "Bearer",
		})
	}
}

// TokenValidator validates a raw JWT string and returns the empire number.
// Returns 0 and an error if the token is invalid.
type TokenValidator func(token string) (empireNo int, err error)

// GetMe returns the identity of the current user.
// If the request carries a valid Bearer token, empire is the token's subject and
// name is "Empire <n>". Otherwise, empire is 0 and name is "guest".
func GetMe(validate TokenValidator) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		empire := 0
		authenticated := false
		name := "guest"

		if h := c.Request().Header.Get("Authorization"); len(h) > 7 && h[:7] == "Bearer " {
			if empireNo, err := validate(h[7:]); err == nil && empireNo > 0 {
				empire = empireNo
				authenticated = true
				name = fmt.Sprintf("Empire %d", empireNo)
			}
		}

		return c.JSON(http.StatusOK, map[string]any{
			"empire":        empire,
			"authenticated": authenticated,
			"name":          name,
		})
	}
}

// PostLogout returns a 200 ok response.
func PostLogout() func(c *echo.Context) error {
	return func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

// GetOrders returns the orders for the authenticated empire.
// Requires EmpireAuthMiddleware to have validated ownership.
func GetOrders(orderStore app.OrderStore) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		empireNo, _ := EmpireFromCtx(c)

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

// PostOrders stores the orders for the authenticated empire.
// Requires EmpireAuthMiddleware to have validated ownership.
func PostOrders(orderStore app.OrderStore, maxBodyBytes int64) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		empireNo, _ := EmpireFromCtx(c)

		rawBody, err := io.ReadAll(io.LimitReader(c.Request().Body, maxBodyBytes+1))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}
		if int64(len(rawBody)) > maxBodyBytes {
			return c.JSON(http.StatusRequestEntityTooLarge, map[string]any{"error": "request body too large"})
		}

		ctx := c.Request().Context()
		if err := orderStore.PutOrders(ctx, empireNo, string(rawBody)); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		return c.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

// GetReports returns the list of reports for the authenticated empire.
// Requires EmpireAuthMiddleware to have validated ownership.
func GetReports(reportStore app.ReportStore) func(c *echo.Context) error {
	type reportItem struct {
		TurnYear    int    `json:"turn_year"`
		TurnQuarter int    `json:"turn_quarter"`
		Link        string `json:"link"`
	}
	return func(c *echo.Context) error {
		empireNo, _ := EmpireFromCtx(c)

		ctx := c.Request().Context()
		reports, err := reportStore.ListReports(ctx, empireNo)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		items := make([]reportItem, len(reports))
		for i, r := range reports {
			items[i] = reportItem{
				TurnYear:    r.TurnYear,
				TurnQuarter: r.TurnQuarter,
				Link:        fmt.Sprintf("/api/%d/reports/%d/%d", empireNo, r.TurnYear, r.TurnQuarter),
			}
		}

		return c.JSON(http.StatusOK, items)
	}
}

// GetReport returns a specific turn report for the authenticated empire.
// Requires EmpireAuthMiddleware to have validated ownership.
func GetReport(reportStore app.ReportStore) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		empireNo, _ := EmpireFromCtx(c)

		turnYear, turnQuarter, err := parseTurnParams(c)
		if err != nil {
			return err
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

// parseTurnParams extracts and validates turnYear and turnQuarter path params.
func parseTurnParams(c *echo.Context) (int, int, error) {
	turnYear, err := parseIntParam(c, "turnYear")
	if err != nil {
		return 0, 0, c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid turn year"})
	}
	turnQuarter, err := parseIntParam(c, "turnQuarter")
	if err != nil {
		return 0, 0, c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid turn quarter"})
	}
	return turnYear, turnQuarter, nil
}

// parseIntParam extracts a named path parameter as an int.
func parseIntParam(c *echo.Context, name string) (int, error) {
	v := c.Param(name)
	if v == "" {
		return 0, errors.New("missing parameter")
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// GetDashboard returns colony, ship, and planet summary counts for the
// authenticated empire.
// Requires EmpireAuthMiddleware to have validated ownership.
func GetDashboard(store app.DashboardStore) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		empireNo, _ := EmpireFromCtx(c)

		summary, err := store.GetDashboardSummary(empireNo)
		if err != nil {
			if errors.Is(err, cerr.ErrNotFound) {
				return c.JSON(http.StatusNotFound, map[string]any{"error": "not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}

		return c.JSON(http.StatusOK, summary)
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
