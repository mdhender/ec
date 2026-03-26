// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package filestore_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/ec/internal/domain"
	"github.com/mdhender/ec/internal/infra/filestore"
)

func TestReadHomeworldTemplate(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore("")

	t.Run("round-trip", func(t *testing.T) {
		original := domain.HomeworldTemplate{
			Habitability: 30,
			Deposits: []domain.DepositTemplate{
				{Resource: domain.METALLICS, YieldPct: 60, QuantityRemaining: 1000},
				{Resource: domain.NONMETALLICS, YieldPct: 40, QuantityRemaining: 800},
			},
		}

		data, err := json.MarshalIndent(original, "", "  ")
		if err != nil {
			t.Fatalf("MarshalIndent: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "homeworld-template.json"), data, 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		got, err := store.ReadHomeworldTemplate(dir)
		if err != nil {
			t.Fatalf("ReadHomeworldTemplate: %v", err)
		}

		if got.Habitability != original.Habitability {
			t.Errorf("Habitability: got %d, want %d", got.Habitability, original.Habitability)
		}
		if len(got.Deposits) != len(original.Deposits) {
			t.Fatalf("Deposits length: got %d, want %d", len(got.Deposits), len(original.Deposits))
		}
		for i, d := range original.Deposits {
			if got.Deposits[i].Resource != d.Resource {
				t.Errorf("Deposits[%d].Resource: got %v, want %v", i, got.Deposits[i].Resource, d.Resource)
			}
			if got.Deposits[i].YieldPct != d.YieldPct {
				t.Errorf("Deposits[%d].YieldPct: got %d, want %d", i, got.Deposits[i].YieldPct, d.YieldPct)
			}
			if got.Deposits[i].QuantityRemaining != d.QuantityRemaining {
				t.Errorf("Deposits[%d].QuantityRemaining: got %d, want %d", i, got.Deposits[i].QuantityRemaining, d.QuantityRemaining)
			}
		}
	})

	t.Run("missing-file", func(t *testing.T) {
		emptyDir := t.TempDir()
		if _, err := store.ReadHomeworldTemplate(emptyDir); err == nil {
			t.Fatal("expected error reading missing homeworld-template.json, got nil")
		}
	})
}

func TestReadColonyTemplate(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore("")

	t.Run("round-trip", func(t *testing.T) {
		original := domain.ColonyTemplate{
			Kind:      domain.OpenAir,
			TechLevel: 1,
			Inventory: []domain.Inventory{
				{Unit: domain.Farm, TechLevel: 1, QuantityAssembled: 10},
			},
		}

		data, err := json.MarshalIndent(original, "", "  ")
		if err != nil {
			t.Fatalf("MarshalIndent: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "colony-template.json"), data, 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		got, err := store.ReadColonyTemplate(dir)
		if err != nil {
			t.Fatalf("ReadColonyTemplate: %v", err)
		}

		if got.Kind != original.Kind {
			t.Errorf("Kind: got %v, want %v", got.Kind, original.Kind)
		}
		if got.TechLevel != original.TechLevel {
			t.Errorf("TechLevel: got %d, want %d", got.TechLevel, original.TechLevel)
		}
		if len(got.Inventory) != len(original.Inventory) {
			t.Fatalf("Inventory length: got %d, want %d", len(got.Inventory), len(original.Inventory))
		}
		for i, inv := range original.Inventory {
			if got.Inventory[i].Unit != inv.Unit {
				t.Errorf("Inventory[%d].Unit: got %v, want %v", i, got.Inventory[i].Unit, inv.Unit)
			}
			if got.Inventory[i].TechLevel != inv.TechLevel {
				t.Errorf("Inventory[%d].TechLevel: got %d, want %d", i, got.Inventory[i].TechLevel, inv.TechLevel)
			}
			if got.Inventory[i].QuantityAssembled != inv.QuantityAssembled {
				t.Errorf("Inventory[%d].QuantityAssembled: got %d, want %d", i, got.Inventory[i].QuantityAssembled, inv.QuantityAssembled)
			}
			if got.Inventory[i].QuantityDisassembled != inv.QuantityDisassembled {
				t.Errorf("Inventory[%d].QuantityDisassembled: got %d, want %d", i, got.Inventory[i].QuantityDisassembled, inv.QuantityDisassembled)
			}
		}
	})

	t.Run("missing-file", func(t *testing.T) {
		emptyDir := t.TempDir()
		if _, err := store.ReadColonyTemplate(emptyDir); err == nil {
			t.Fatal("expected error reading missing colony-template.json, got nil")
		}
	})
}
