// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package filestore_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/ec/internal/cerr"
	"github.com/mdhender/ec/internal/domain"
	"github.com/mdhender/ec/internal/infra/filestore"
)

// writeGameJSON writes a domain.Game as game.json into dir.
func writeGameJSON(t *testing.T, dir string, game domain.Game) {
	t.Helper()
	data, err := json.MarshalIndent(game, "", "  ")
	if err != nil {
		t.Fatalf("marshal game.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "game.json"), data, 0o644); err != nil {
		t.Fatalf("write game.json: %v", err)
	}
}

// writeClusterJSON writes a domain.Cluster as cluster.json into dir.
func writeClusterJSON(t *testing.T, dir string, cluster domain.Cluster) {
	t.Helper()
	data, err := json.MarshalIndent(cluster, "", "  ")
	if err != nil {
		t.Fatalf("marshal cluster.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cluster.json"), data, 0o644); err != nil {
		t.Fatalf("write cluster.json: %v", err)
	}
}

func TestGetDashboardSummary_OneColony(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore(dir)

	writeGameJSON(t, dir, domain.Game{
		Empires: []domain.Empire{
			{ID: 1, Colonies: []domain.ColonyID{1}},
		},
	})
	writeClusterJSON(t, dir, domain.Cluster{
		Colonies: []domain.Colony{
			{ID: 1, Planet: 10, Kind: domain.OpenAir},
		},
		Planets: []domain.Planet{
			{ID: 10, Kind: domain.Terrestrial},
		},
	})

	summary, err := store.GetDashboardSummary(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.ColonyCount != 1 {
		t.Errorf("ColonyCount: expected 1, got %d", summary.ColonyCount)
	}
	if len(summary.ColonyKinds) != 1 || summary.ColonyKinds[0].Kind != "Open Air" || summary.ColonyKinds[0].Count != 1 {
		t.Errorf("ColonyKinds: expected [{Open Air 1}], got %v", summary.ColonyKinds)
	}
	if summary.ShipCount != 0 {
		t.Errorf("ShipCount: expected 0, got %d", summary.ShipCount)
	}
	if summary.PlanetCount != 1 {
		t.Errorf("PlanetCount: expected 1, got %d", summary.PlanetCount)
	}
	if len(summary.PlanetKinds) != 1 || summary.PlanetKinds[0].Kind != "Terrestrial" || summary.PlanetKinds[0].Count != 1 {
		t.Errorf("PlanetKinds: expected [{Terrestrial 1}], got %v", summary.PlanetKinds)
	}
}

func TestGetDashboardSummary_MultipleKinds(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore(dir)

	writeGameJSON(t, dir, domain.Game{
		Empires: []domain.Empire{
			{ID: 1, Colonies: []domain.ColonyID{1, 2}},
		},
	})
	writeClusterJSON(t, dir, domain.Cluster{
		Colonies: []domain.Colony{
			{ID: 1, Planet: 10, Kind: domain.OpenAir},
			{ID: 2, Planet: 20, Kind: domain.Orbital},
		},
		Planets: []domain.Planet{
			{ID: 10, Kind: domain.Terrestrial},
			{ID: 20, Kind: domain.GasGiant},
		},
	})

	summary, err := store.GetDashboardSummary(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.ColonyCount != 2 {
		t.Errorf("ColonyCount: expected 2, got %d", summary.ColonyCount)
	}
	if len(summary.ColonyKinds) != 2 {
		t.Fatalf("ColonyKinds: expected 2 entries, got %d: %v", len(summary.ColonyKinds), summary.ColonyKinds)
	}
	// Sorted by Kind ascending: "Open Air" < "Orbital"
	if summary.ColonyKinds[0].Kind != "Open Air" {
		t.Errorf("ColonyKinds[0]: expected Open Air, got %q", summary.ColonyKinds[0].Kind)
	}
	if summary.ColonyKinds[1].Kind != "Orbital" {
		t.Errorf("ColonyKinds[1]: expected Orbital, got %q", summary.ColonyKinds[1].Kind)
	}

	if len(summary.PlanetKinds) != 2 {
		t.Fatalf("PlanetKinds: expected 2 entries, got %d: %v", len(summary.PlanetKinds), summary.PlanetKinds)
	}
	// Sorted by Kind ascending: "Gas Giant" < "Terrestrial"
	if summary.PlanetKinds[0].Kind != "Gas Giant" {
		t.Errorf("PlanetKinds[0]: expected Gas Giant, got %q", summary.PlanetKinds[0].Kind)
	}
	if summary.PlanetKinds[1].Kind != "Terrestrial" {
		t.Errorf("PlanetKinds[1]: expected Terrestrial, got %q", summary.PlanetKinds[1].Kind)
	}
}

func TestGetDashboardSummary_DeduplicatesPlanets(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore(dir)

	writeGameJSON(t, dir, domain.Game{
		Empires: []domain.Empire{
			{ID: 1, Colonies: []domain.ColonyID{1, 2}},
		},
	})
	writeClusterJSON(t, dir, domain.Cluster{
		Colonies: []domain.Colony{
			{ID: 1, Planet: 10, Kind: domain.OpenAir},
			{ID: 2, Planet: 10, Kind: domain.Enclosed},
		},
		Planets: []domain.Planet{
			{ID: 10, Kind: domain.Terrestrial},
		},
	})

	summary, err := store.GetDashboardSummary(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.PlanetCount != 1 {
		t.Errorf("PlanetCount: expected 1 (deduplicated), got %d", summary.PlanetCount)
	}
}

func TestGetDashboardSummary_EmpireNotFound(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore(dir)

	writeGameJSON(t, dir, domain.Game{
		Empires: []domain.Empire{
			{ID: 1, Colonies: []domain.ColonyID{}},
		},
	})
	writeClusterJSON(t, dir, domain.Cluster{})

	_, err := store.GetDashboardSummary(99)
	if err == nil {
		t.Fatal("expected error for missing empire, got nil")
	}
	if !errors.Is(err, cerr.ErrNotFound) {
		t.Errorf("expected cerr.ErrNotFound, got %v", err)
	}
}

func TestGetDashboardSummary_NoColonies(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore(dir)

	writeGameJSON(t, dir, domain.Game{
		Empires: []domain.Empire{
			{ID: 1, Colonies: []domain.ColonyID{}},
		},
	})
	writeClusterJSON(t, dir, domain.Cluster{})

	summary, err := store.GetDashboardSummary(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.ColonyCount != 0 {
		t.Errorf("ColonyCount: expected 0, got %d", summary.ColonyCount)
	}
	if len(summary.ColonyKinds) != 0 {
		t.Errorf("ColonyKinds: expected empty, got %v", summary.ColonyKinds)
	}
	if summary.ShipCount != 0 {
		t.Errorf("ShipCount: expected 0, got %d", summary.ShipCount)
	}
	if summary.PlanetCount != 0 {
		t.Errorf("PlanetCount: expected 0, got %d", summary.PlanetCount)
	}
}
