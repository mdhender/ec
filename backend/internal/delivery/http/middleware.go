// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

const empireCtxKey = "ec.empireNo"

// EmpireExtractor extracts the authenticated empire number from a request context.
// Returns the empire number and true on success, or 0 and false if unavailable.
type EmpireExtractor func(c *echo.Context) (empireNo int, ok bool)

// EmpireAuthMiddleware returns middleware that validates empire ownership.
// It extracts the :empireNo path parameter, calls the extractor to get the
// authenticated empire number, and compares them. On success it stores the
// validated empire number in context for handlers to retrieve via EmpireFromCtx.
func EmpireAuthMiddleware(extract EmpireExtractor) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			empireNo, err := strconv.Atoi(c.Param("empireNo"))
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid empire number"})
			}

			jwtEmpireNo, ok := extract(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			}

			if empireNo != jwtEmpireNo {
				return c.JSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
			}

			c.Set(empireCtxKey, empireNo)
			return next(c)
		}
	}
}

// EmpireFromCtx returns the validated empire number set by EmpireAuthMiddleware.
func EmpireFromCtx(c *echo.Context) (int, bool) {
	v := c.Get(empireCtxKey)
	if v == nil {
		return 0, false
	}
	n, ok := v.(int)
	return n, ok
}
