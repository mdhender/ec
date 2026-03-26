// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

// KindCount pairs a human-readable kind label with a count.
type KindCount struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}

// DashboardSummary is the response payload for GET /api/:empireNo/dashboard.
type DashboardSummary struct {
	ColonyCount int         `json:"colony_count"`
	ColonyKinds []KindCount `json:"colony_kinds"`
	ShipCount   int         `json:"ship_count"`
	PlanetCount int         `json:"planet_count"`
	PlanetKinds []KindCount `json:"planet_kinds"`
}

// DashboardStore computes dashboard summary data for a given empire.
type DashboardStore interface {
	GetDashboardSummary(empireNo int) (DashboardSummary, error)
}
