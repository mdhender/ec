// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package http

import (
	"github.com/labstack/echo/v5"

	"github.com/mdhender/ec/internal/app"
)

// AddRoutes registers all API routes on the given Echo instance.
func AddRoutes(
	e *echo.Echo,
	jwtMiddleware echo.MiddlewareFunc,
	authStore app.AuthStore,
	tokenSigner app.TokenSigner,
	orderStore app.OrderStore,
	reportStore app.ReportStore,
	shutdownKey string,
	shutdownCh chan struct{},
) {
	// Public routes
	e.GET("/api/health", GetHealth())
	e.POST("/api/login/:magicLink", PostLogin(authStore, tokenSigner))
	e.POST("/api/logout", PostLogout())

	// Protected routes (require valid JWT).
	// If no middleware is provided (e.g. in tests), use a pass-through.
	if jwtMiddleware == nil {
		jwtMiddleware = func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}
	protected := e.Group("", jwtMiddleware)
	protected.GET("/api/:empireNo/orders", GetOrders(orderStore))
	protected.POST("/api/:empireNo/orders", PostOrders(orderStore))
	protected.GET("/api/:empireNo/reports", GetReports(reportStore))
	protected.GET("/api/:empireNo/reports/:turnYear/:turnQuarter", GetReport(reportStore))

	// Shutdown route (only registered if shutdownKey is set)
	if shutdownKey != "" {
		e.POST("/api/shutdown/:key", PostShutdown(shutdownKey, shutdownCh))
	}
}
