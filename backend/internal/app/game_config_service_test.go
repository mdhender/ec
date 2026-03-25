// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app_test

import (
	"errors"
	"testing"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/domain"
)

// mockGameConfigStore is an in-memory GameConfigStore for testing.
type mockGameConfigStore struct {
	gameConfigs   map[string]domain.GameConfig
	authConfigs   map[string]domain.AuthConfig
	forceWriteErr error
}

func newMockStore() *mockGameConfigStore {
	return &mockGameConfigStore{
		gameConfigs: map[string]domain.GameConfig{},
		authConfigs: map[string]domain.AuthConfig{},
	}
}

func (m *mockGameConfigStore) ValidateDir(path string) error {
	return nil
}

func (m *mockGameConfigStore) GameConfigExists(dirPath string) (bool, error) {
	_, ok := m.gameConfigs[dirPath]
	return ok, nil
}

func (m *mockGameConfigStore) AuthConfigExists(dirPath string) (bool, error) {
	_, ok := m.authConfigs[dirPath]
	return ok, nil
}

func (m *mockGameConfigStore) ReadGameConfig(path string) (domain.GameConfig, error) {
	cfg, ok := m.gameConfigs[path]
	if !ok {
		return domain.GameConfig{}, errors.New("game.json not found")
	}
	return cfg, nil
}

func (m *mockGameConfigStore) WriteGameConfig(path string, cfg domain.GameConfig) error {
	if m.forceWriteErr != nil {
		return m.forceWriteErr
	}
	m.gameConfigs[path] = cfg
	return nil
}

func (m *mockGameConfigStore) ReadAuthConfig(path string) (domain.AuthConfig, error) {
	cfg, ok := m.authConfigs[path]
	if !ok {
		return domain.AuthConfig{}, errors.New("auth.json not found")
	}
	return cfg, nil
}

func (m *mockGameConfigStore) WriteAuthConfig(path string, cfg domain.AuthConfig) error {
	if m.forceWriteErr != nil {
		return m.forceWriteErr
	}
	m.authConfigs[path] = cfg
	return nil
}

func (m *mockGameConfigStore) CreateEmpireDir(dirPath string, empireNo int) error {
	return nil
}

// --- TestCreateGame ---

func TestCreateGame(t *testing.T) {
	t.Run("writes empty game and auth configs", func(t *testing.T) {
		store := newMockStore()
		svc := &app.GameConfigService{Store: store}
		dir := "/test/dir"

		if err := svc.CreateGame(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		gc, ok := store.gameConfigs[dir]
		if !ok {
			t.Fatal("WriteGameConfig was not called")
		}
		if gc.Empires == nil || len(gc.Empires) != 0 {
			t.Errorf("expected empty empires slice, got %v", gc.Empires)
		}

		ac, ok := store.authConfigs[dir]
		if !ok {
			t.Fatal("WriteAuthConfig was not called")
		}
		if ac.MagicLinks == nil || len(ac.MagicLinks) != 0 {
			t.Errorf("expected empty magic links map, got %v", ac.MagicLinks)
		}
	})

	t.Run("fails if game.json already exists", func(t *testing.T) {
		store := newMockStore()
		store.gameConfigs["/test/dir"] = domain.GameConfig{}
		svc := &app.GameConfigService{Store: store}

		if err := svc.CreateGame("/test/dir"); err == nil {
			t.Fatal("expected error when game.json exists, got nil")
		}
	})

	t.Run("fails if auth.json already exists", func(t *testing.T) {
		store := newMockStore()
		store.authConfigs["/test/dir"] = domain.AuthConfig{}
		svc := &app.GameConfigService{Store: store}

		if err := svc.CreateGame("/test/dir"); err == nil {
			t.Fatal("expected error when auth.json exists, got nil")
		}
	})
}

// --- TestAddEmpire ---

func TestAddEmpire(t *testing.T) {
	t.Run("auto-assigns 1 for empty list", func(t *testing.T) {
		store := newMockStore()
		svc := &app.GameConfigService{Store: store}
		dir := "/test/dir"
		store.gameConfigs[dir] = domain.GameConfig{Empires: []domain.EmpireEntry{}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}

		n, uuid, err := svc.AddEmpire(dir, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 1 {
			t.Errorf("expected empire 1, got %d", n)
		}
		if uuid == "" {
			t.Error("expected non-empty uuid")
		}
	})

	t.Run("auto-assigns max+1 for non-empty list", func(t *testing.T) {
		store := newMockStore()
		svc := &app.GameConfigService{Store: store}
		dir := "/test/dir"
		store.gameConfigs[dir] = domain.GameConfig{Empires: []domain.EmpireEntry{
			{Empire: 3, Active: true},
			{Empire: 7, Active: true},
		}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}

		n, _, err := svc.AddEmpire(dir, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 8 {
			t.Errorf("expected empire 8, got %d", n)
		}
	})

	t.Run("fails on duplicate empire", func(t *testing.T) {
		store := newMockStore()
		svc := &app.GameConfigService{Store: store}
		dir := "/test/dir"
		store.gameConfigs[dir] = domain.GameConfig{Empires: []domain.EmpireEntry{
			{Empire: 5, Active: true},
		}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}

		if _, _, err := svc.AddEmpire(dir, 5); err == nil {
			t.Fatal("expected error for duplicate empire, got nil")
		}
	})

	t.Run("generates magic link UUID", func(t *testing.T) {
		store := newMockStore()
		svc := &app.GameConfigService{Store: store}
		dir := "/test/dir"
		store.gameConfigs[dir] = domain.GameConfig{Empires: []domain.EmpireEntry{}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}

		_, uuid, err := svc.AddEmpire(dir, 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// UUID should be in xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx format (36 chars)
		if len(uuid) != 36 {
			t.Errorf("expected UUID of length 36, got %d: %q", len(uuid), uuid)
		}
		// Verify it's recorded in auth config
		ac := store.authConfigs[dir]
		link, ok := ac.MagicLinks[uuid]
		if !ok {
			t.Fatalf("magic link %q not found in auth config", uuid)
		}
		if link.Empire != 42 {
			t.Errorf("expected empire 42, got %d", link.Empire)
		}
	})
}

// --- TestRemoveEmpire ---

func TestRemoveEmpire(t *testing.T) {
	t.Run("sets active to false", func(t *testing.T) {
		store := newMockStore()
		svc := &app.GameConfigService{Store: store}
		dir := "/test/dir"
		store.gameConfigs[dir] = domain.GameConfig{Empires: []domain.EmpireEntry{
			{Empire: 10, Active: true},
		}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{
			"some-uuid": {Empire: 10},
		}}

		if err := svc.RemoveEmpire(dir, 10); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		gc := store.gameConfigs[dir]
		if gc.Empires[0].Active {
			t.Error("expected empire to be inactive")
		}
	})

	t.Run("removes magic link", func(t *testing.T) {
		store := newMockStore()
		svc := &app.GameConfigService{Store: store}
		dir := "/test/dir"
		store.gameConfigs[dir] = domain.GameConfig{Empires: []domain.EmpireEntry{
			{Empire: 10, Active: true},
		}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{
			"some-uuid": {Empire: 10},
		}}

		if err := svc.RemoveEmpire(dir, 10); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ac := store.authConfigs[dir]
		if _, ok := ac.MagicLinks["some-uuid"]; ok {
			t.Error("expected magic link to be removed")
		}
	})

	t.Run("no error when magic link missing", func(t *testing.T) {
		store := newMockStore()
		svc := &app.GameConfigService{Store: store}
		dir := "/test/dir"
		store.gameConfigs[dir] = domain.GameConfig{Empires: []domain.EmpireEntry{
			{Empire: 10, Active: true},
		}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}

		if err := svc.RemoveEmpire(dir, 10); err != nil {
			t.Fatalf("unexpected error when no magic link: %v", err)
		}
	})

	t.Run("fails when empire not found", func(t *testing.T) {
		store := newMockStore()
		svc := &app.GameConfigService{Store: store}
		dir := "/test/dir"
		store.gameConfigs[dir] = domain.GameConfig{Empires: []domain.EmpireEntry{}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}

		if err := svc.RemoveEmpire(dir, 99); err == nil {
			t.Fatal("expected error for missing empire, got nil")
		}
	})
}
