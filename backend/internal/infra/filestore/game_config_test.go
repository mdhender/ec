// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package filestore_test

import (
	"testing"

	"github.com/mdhender/ec/internal/domain"
	"github.com/mdhender/ec/internal/infra/filestore"
)

func TestGameConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore("")

	original := domain.GameConfig{
		Empires: []domain.EmpireEntry{
			{Empire: 42, Active: true},
			{Empire: 1812, Active: false},
		},
	}

	if err := store.WriteGameConfig(dir, original); err != nil {
		t.Fatalf("WriteGameConfig: %v", err)
	}

	got, err := store.ReadGameConfig(dir)
	if err != nil {
		t.Fatalf("ReadGameConfig: %v", err)
	}

	if len(got.Empires) != len(original.Empires) {
		t.Fatalf("expected %d empires, got %d", len(original.Empires), len(got.Empires))
	}
	for i, e := range original.Empires {
		if got.Empires[i].Empire != e.Empire || got.Empires[i].Active != e.Active {
			t.Errorf("empire[%d]: expected %+v, got %+v", i, e, got.Empires[i])
		}
	}
}

func TestAuthConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore("")

	original := domain.AuthConfig{
		MagicLinks: map[string]domain.AuthLink{
			"37e81785-84ee-4fee-850b-160e373a4539": {Empire: 42},
			"81ce2bb6-42fe-49b2-80c5-0558787c8471": {Empire: 1812},
		},
	}

	if err := store.WriteAuthConfig(dir, original); err != nil {
		t.Fatalf("WriteAuthConfig: %v", err)
	}

	got, err := store.ReadAuthConfig(dir)
	if err != nil {
		t.Fatalf("ReadAuthConfig: %v", err)
	}

	if len(got.MagicLinks) != len(original.MagicLinks) {
		t.Fatalf("expected %d magic links, got %d", len(original.MagicLinks), len(got.MagicLinks))
	}
	for uuid, link := range original.MagicLinks {
		gotLink, ok := got.MagicLinks[uuid]
		if !ok {
			t.Errorf("magic link %q not found", uuid)
			continue
		}
		if gotLink.Empire != link.Empire {
			t.Errorf("magic link %q: expected empire %d, got %d", uuid, link.Empire, gotLink.Empire)
		}
	}
}

func TestReadGameConfigMissing(t *testing.T) {
	dir := t.TempDir()
	store := filestore.NewStore("")

	if _, err := store.ReadGameConfig(dir); err == nil {
		t.Fatal("expected error reading missing game.json, got nil")
	}
}
