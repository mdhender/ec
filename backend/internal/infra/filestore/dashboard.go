// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package filestore

import (
	"fmt"
	"sort"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/cerr"
	"github.com/mdhender/ec/internal/domain"
)

// compile-time interface check
var _ app.DashboardStore = (*Store)(nil)

// GetDashboardSummary reads game.json and cluster.json and returns summary
// counts for the given empire's colonies, ships, and planets.
func (s *Store) GetDashboardSummary(empireNo int) (app.DashboardSummary, error) {
	game, err := s.ReadGame(s.dataPath)
	if err != nil {
		return app.DashboardSummary{}, fmt.Errorf("getDashboardSummary: %w", err)
	}

	var empire *domain.Empire
	for i := range game.Empires {
		if int(game.Empires[i].ID) == empireNo {
			empire = &game.Empires[i]
			break
		}
	}
	if empire == nil {
		return app.DashboardSummary{}, fmt.Errorf("getDashboardSummary: empire %d not found: %w", empireNo, cerr.ErrNotFound)
	}

	cluster, err := s.ReadCluster(s.dataPath)
	if err != nil {
		return app.DashboardSummary{}, fmt.Errorf("getDashboardSummary: %w", err)
	}

	colonyByID := make(map[domain.ColonyID]domain.Colony, len(cluster.Colonies))
	for _, col := range cluster.Colonies {
		colonyByID[col.ID] = col
	}

	planetByID := make(map[domain.PlanetID]domain.Planet, len(cluster.Planets))
	for _, planet := range cluster.Planets {
		planetByID[planet.ID] = planet
	}

	colonyCounts := map[string]int{}
	for _, colonyID := range empire.Colonies {
		col, ok := colonyByID[colonyID]
		if !ok {
			continue // data inconsistency — skip silently
		}
		colonyCounts[col.Kind.String()]++
	}

	seenPlanet := map[domain.PlanetID]bool{}
	planetCounts := map[string]int{}
	for _, colonyID := range empire.Colonies {
		col, ok := colonyByID[colonyID]
		if !ok {
			continue
		}
		if seenPlanet[col.Planet] {
			continue
		}
		seenPlanet[col.Planet] = true
		planet, ok := planetByID[col.Planet]
		if !ok {
			continue
		}
		planetCounts[planet.Kind.String()]++
	}

	colonyKinds := toKindCounts(colonyCounts)
	planetKinds := toKindCounts(planetCounts)

	colonyTotal := 0
	for _, kc := range colonyKinds {
		colonyTotal += kc.Count
	}

	return app.DashboardSummary{
		ColonyCount: colonyTotal,
		ColonyKinds: colonyKinds,
		ShipCount:   0,
		PlanetCount: len(seenPlanet),
		PlanetKinds: planetKinds,
	}, nil
}

// toKindCounts converts a map of kind→count into a sorted []KindCount,
// omitting entries with count 0.
func toKindCounts(m map[string]int) []app.KindCount {
	var result []app.KindCount
	for kind, count := range m {
		if count == 0 {
			continue
		}
		result = append(result, app.KindCount{Kind: kind, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Kind < result[j].Kind
	})
	return result
}
