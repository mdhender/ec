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
	empireExtractor EmpireExtractor,
	tokenValidator TokenValidator,
	loginSvc *app.LoginService,
	orderStore app.OrderStore,
	reportStore app.ReportStore,
	dashboardStore app.DashboardStore,
	shutdownKey string,
	shutdownCh chan struct{},
	maxOrderBytes int64,
) {
	// Public routes
	e.GET("/api/health", GetHealth())
	e.GET("/api/me", GetMe(tokenValidator))
	e.POST("/api/login/:magicLink", PostLogin(loginSvc))
	e.POST("/api/logout", PostLogout())

	// Protected routes (require valid JWT + empire ownership).
	if jwtMiddleware == nil {
		jwtMiddleware = func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}
	empireAuth := EmpireAuthMiddleware(empireExtractor)
	protected := e.Group("", jwtMiddleware, empireAuth)
	protected.GET("/api/:empireNo/orders", GetOrders(orderStore))
	protected.POST("/api/:empireNo/orders", PostOrders(orderStore, maxOrderBytes))
	protected.GET("/api/:empireNo/reports", GetReports(reportStore))
	protected.GET("/api/:empireNo/reports/:turnYear/:turnQuarter", GetReport(reportStore))
	protected.GET("/api/:empireNo/dashboard", GetDashboard(dashboardStore))

	// Shutdown route (only registered if shutdownKey is set)
	if shutdownKey != "" {
		e.POST("/api/shutdown/:key", PostShutdown(shutdownKey, shutdownCh))
	}
}
