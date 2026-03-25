// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package clustergen

import (
	"github.com/mdhender/ec/internal/domain"
)

// DepositStats holds aggregate statistics for a resource type.
type DepositStats struct {
	Count    int
	TotalQty int64
	TotalPct int64
}

type ClusterStats struct {
	NumSystems      int
	NumStars        int
	TotalPlanets    int
	PlanetsByKind   map[domain.PlanetKind]int
	HabitableByKind map[domain.PlanetKind]int
	TotalHabByKind  map[domain.PlanetKind]int
	Overall         map[domain.NaturalResource]*DepositStats
	ByPlanetKind    map[domain.PlanetKind]map[domain.NaturalResource]*DepositStats
}

func NewClusterStats() *ClusterStats {
	planetKinds := []domain.PlanetKind{domain.Terrestrial, domain.AsteroidBelt, domain.GasGiant}
	resourceKinds := []domain.NaturalResource{domain.GOLD, domain.FUEL, domain.METALLICS, domain.NONMETALLICS}
	cs := &ClusterStats{
		PlanetsByKind:   map[domain.PlanetKind]int{},
		HabitableByKind: map[domain.PlanetKind]int{},
		TotalHabByKind:  map[domain.PlanetKind]int{},
		Overall:         map[domain.NaturalResource]*DepositStats{},
		ByPlanetKind:    map[domain.PlanetKind]map[domain.NaturalResource]*DepositStats{},
	}
	for _, rk := range resourceKinds {
		cs.Overall[rk] = &DepositStats{}
	}
	for _, pk := range planetKinds {
		cs.ByPlanetKind[pk] = map[domain.NaturalResource]*DepositStats{}
		for _, rk := range resourceKinds {
			cs.ByPlanetKind[pk][rk] = &DepositStats{}
		}
	}
	return cs
}

// Collect aggregates statistics from a normalized domain.Cluster.
func (cs *ClusterStats) Collect(c domain.Cluster) {
	// Build deposit lookup by ID for efficient access.
	depByID := make(map[domain.DepositID]domain.Deposit, len(c.Deposits))
	for _, dep := range c.Deposits {
		depByID[dep.ID] = dep
	}

	cs.NumSystems += len(c.Systems)
	cs.NumStars += len(c.Stars)

	for _, p := range c.Planets {
		cs.TotalPlanets++
		cs.PlanetsByKind[p.Kind]++
		if p.Habitability != 0 {
			cs.HabitableByKind[p.Kind]++
			cs.TotalHabByKind[p.Kind] += p.Habitability
		}
		for _, depID := range p.Deposits {
			dep := depByID[depID]
			cs.Overall[dep.Resource].Count++
			cs.Overall[dep.Resource].TotalQty += int64(dep.QuantityRemaining)
			cs.Overall[dep.Resource].TotalPct += int64(dep.YieldPct)
			cs.ByPlanetKind[p.Kind][dep.Resource].Count++
			cs.ByPlanetKind[p.Kind][dep.Resource].TotalQty += int64(dep.QuantityRemaining)
			cs.ByPlanetKind[p.Kind][dep.Resource].TotalPct += int64(dep.YieldPct)
		}
	}
}
