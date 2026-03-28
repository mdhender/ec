// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package http

import (
	"github.com/labstack/echo/v5"

	"github.com/mdhender/ec/internal/app"
)

// RouteDeps holds all dependencies needed to register API routes.
type RouteDeps struct {
	Echo            *echo.Echo
	JWTMiddleware   echo.MiddlewareFunc
	EmpireExtractor EmpireExtractor
	TokenValidator  TokenValidator
	LoginSvc        *app.LoginService
	OrderStore      app.OrderStore
	ReportStore     app.ReportStore
	DashboardStore  app.DashboardStore
	ShutdownKey     string
	ShutdownCh      chan struct{}
	MaxOrderBytes   int64
	ParseOrdersSvc  *app.ParseOrdersService
}

// AddRoutes registers all API routes on the given Echo instance.
func AddRoutes(deps RouteDeps) {
	e := deps.Echo
	jwtMiddleware := deps.JWTMiddleware
	empireExtractor := deps.EmpireExtractor
	tokenValidator := deps.TokenValidator
	loginSvc := deps.LoginSvc
	orderStore := deps.OrderStore
	reportStore := deps.ReportStore
	dashboardStore := deps.DashboardStore
	shutdownKey := deps.ShutdownKey
	shutdownCh := deps.ShutdownCh
	maxOrderBytes := deps.MaxOrderBytes
	parseOrdersSvc := deps.ParseOrdersSvc
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
	protected.POST("/api/:empireNo/orders/parse", PostParseOrders(parseOrdersSvc, maxOrderBytes))
	protected.GET("/api/:empireNo/reports", GetReports(reportStore))
	protected.GET("/api/:empireNo/reports/:turnYear/:turnQuarter", GetReport(reportStore))
	protected.GET("/api/:empireNo/dashboard", GetDashboard(dashboardStore))

	// Shutdown route (only registered if shutdownKey is set)
	if shutdownKey != "" {
		e.POST("/api/shutdown/:key", PostShutdown(shutdownKey, shutdownCh))
	}
}
