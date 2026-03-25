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
	games         map[string]domain.Game
	authConfigs   map[string]domain.AuthConfig
	forceWriteErr error
}

func newMockStore() *mockGameConfigStore {
	return &mockGameConfigStore{
		games:       map[string]domain.Game{},
		authConfigs: map[string]domain.AuthConfig{},
	}
}

func (m *mockGameConfigStore) ValidateDir(path string) error {
	return nil
}

func (m *mockGameConfigStore) GameExists(dirPath string) (bool, error) {
	_, ok := m.games[dirPath]
	return ok, nil
}

func (m *mockGameConfigStore) AuthConfigExists(dirPath string) (bool, error) {
	_, ok := m.authConfigs[dirPath]
	return ok, nil
}

func (m *mockGameConfigStore) ReadGame(path string) (domain.Game, error) {
	game, ok := m.games[path]
	if !ok {
		return domain.Game{}, errors.New("game.json not found")
	}
	return game, nil
}

func (m *mockGameConfigStore) WriteGame(path string, game domain.Game) error {
	if m.forceWriteErr != nil {
		return m.forceWriteErr
	}
	m.games[path] = game
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

// mockClusterStore is an in-memory ClusterStore for testing.
type mockClusterStore struct {
	clusters      map[string]domain.Cluster
	forceWriteErr error
}

func newMockClusterStore() *mockClusterStore {
	return &mockClusterStore{
		clusters: map[string]domain.Cluster{},
	}
}

func (m *mockClusterStore) ReadCluster(dataPath string) (domain.Cluster, error) {
	c, ok := m.clusters[dataPath]
	if !ok {
		return domain.Cluster{}, errors.New("cluster.json not found")
	}
	return c, nil
}

func (m *mockClusterStore) WriteCluster(dataPath string, cluster domain.Cluster, overwrite bool) error {
	if m.forceWriteErr != nil {
		return m.forceWriteErr
	}
	m.clusters[dataPath] = cluster
	return nil
}

// --- TestCreateGame ---

func TestCreateGame(t *testing.T) {
	t.Run("writes empty game and auth configs", func(t *testing.T) {
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		dir := "/test/dir"

		if err := svc.CreateGame(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		game, ok := store.games[dir]
		if !ok {
			t.Fatal("WriteGame was not called")
		}
		if game.Empires == nil || len(game.Empires) != 0 {
			t.Errorf("expected empty empires slice, got %v", game.Empires)
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
		store.games["/test/dir"] = domain.Game{}
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}

		if err := svc.CreateGame("/test/dir"); err == nil {
			t.Fatal("expected error when game.json exists, got nil")
		}
	})

	t.Run("fails if auth.json already exists", func(t *testing.T) {
		store := newMockStore()
		store.authConfigs["/test/dir"] = domain.AuthConfig{}
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}

		if err := svc.CreateGame("/test/dir"); err == nil {
			t.Fatal("expected error when auth.json exists, got nil")
		}
	})
}

// --- TestAddEmpire ---

// makeTestCluster builds a minimal cluster with one system, one star, and a terrestrial planet.
func makeTestCluster(planetID domain.PlanetID) domain.Cluster {
	return domain.Cluster{
		Systems: []domain.System{
			{ID: 1, Location: domain.Coords{X: 10, Y: 10, Z: 10}},
		},
		Stars: []domain.Star{
			{ID: 1, System: 1, Orbits: [10]domain.PlanetID{planetID}},
		},
		Planets: []domain.Planet{
			{ID: planetID, Kind: domain.Terrestrial, Habitability: 25},
		},
	}
}

func TestAddEmpire(t *testing.T) {
	const dir = "/test/dir"
	const hwPlanetID domain.PlanetID = 100

	t.Run("auto-assigns 1 for empty list", func(t *testing.T) {
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		store.games[dir] = domain.Game{
			ActiveHomeWorldID: hwPlanetID,
			Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
			Empires:           []domain.Empire{},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
		clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

		n, uuid, err := svc.AddEmpire(dir, 0, "TestEmpire", 0)
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
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		store.games[dir] = domain.Game{
			ActiveHomeWorldID: hwPlanetID,
			Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID, Empires: []domain.EmpireID{3, 7}}},
			Empires: []domain.Empire{
				{ID: 3, Active: true},
				{ID: 7, Active: true},
			},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
		clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

		n, _, err := svc.AddEmpire(dir, 0, "TestEmpire", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != 8 {
			t.Errorf("expected empire 8, got %d", n)
		}
	})

	t.Run("fails on duplicate empire", func(t *testing.T) {
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		store.games[dir] = domain.Game{
			ActiveHomeWorldID: hwPlanetID,
			Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID, Empires: []domain.EmpireID{5}}},
			Empires:           []domain.Empire{{ID: 5, Active: true}},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
		clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

		if _, _, err := svc.AddEmpire(dir, 5, "TestEmpire", 0); err == nil {
			t.Fatal("expected error for duplicate empire, got nil")
		}
	})

	t.Run("generates magic link UUID", func(t *testing.T) {
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		store.games[dir] = domain.Game{
			ActiveHomeWorldID: hwPlanetID,
			Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
			Empires:           []domain.Empire{},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
		clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

		_, uuid, err := svc.AddEmpire(dir, 42, "TestEmpire", 0)
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

func TestAddEmpire_RequiresHomeWorld(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{
		Empires: []domain.Empire{},
		Races:   []domain.Race{},
	}
	store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
	clusterStore.clusters[dir] = domain.Cluster{}

	if _, _, err := svc.AddEmpire(dir, 0, "TestEmpire", 0); err == nil {
		t.Fatal("expected error when no active homeworld, got nil")
	}
}

func TestAddEmpire_HomeWorldOverride(t *testing.T) {
	const dir = "/test/dir"
	const hwPlanetID domain.PlanetID = 200
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{
		ActiveHomeWorldID: 999, // different active homeworld
		Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
		Empires:           []domain.Empire{},
	}
	store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
	clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

	n, _, err := svc.AddEmpire(dir, 0, "TestEmpire", hwPlanetID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero empire number")
	}
	// Verify empire was assigned to the correct race
	game := store.games[dir]
	if game.Empires[0].HomeWorld != hwPlanetID {
		t.Errorf("expected empire homeworld %d, got %d", hwPlanetID, game.Empires[0].HomeWorld)
	}
}

func TestAddEmpire_HomeWorldNotFound(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{
		Races:   []domain.Race{},
		Empires: []domain.Empire{},
	}
	store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
	clusterStore.clusters[dir] = domain.Cluster{}

	// Pass a homeWorldID that doesn't exist in game.Races
	if _, _, err := svc.AddEmpire(dir, 0, "TestEmpire", 999); err == nil {
		t.Fatal("expected error when homeworld not in game.Races, got nil")
	}
}

func TestAddEmpire_HomeWorldFull(t *testing.T) {
	const dir = "/test/dir"
	const hwPlanetID domain.PlanetID = 300
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}

	// Create a race with 25 empires
	empireIDs := make([]domain.EmpireID, 25)
	empires := make([]domain.Empire, 25)
	for i := range empireIDs {
		empireIDs[i] = domain.EmpireID(i + 1)
		empires[i] = domain.Empire{ID: domain.EmpireID(i + 1), Active: true}
	}
	store.games[dir] = domain.Game{
		ActiveHomeWorldID: hwPlanetID,
		Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID, Empires: empireIDs}},
		Empires:           empires,
	}
	store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
	clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

	if _, _, err := svc.AddEmpire(dir, 0, "TestEmpire", 0); err == nil {
		t.Fatal("expected error when homeworld is full (25 empires), got nil")
	}
}

func TestScrubEmpireName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Alpha Empire", "Alpha Empire"},
		{"  Alpha  Empire  ", "Alpha Empire"},
		{"Alpha<>Empire", "AlphaEmpire"},
		{"Alpha!@#Empire", "AlphaEmpire"},
		{"Alpha-Empire", "Alpha-Empire"},
		{"Alpha_Empire", "Alpha_Empire"},
		{"Alpha.Empire", "Alpha.Empire"},
		{"Alpha,Empire", "Alpha,Empire"},
		{"", ""},
		{"   ", ""},
		{"Ré public", "Ré public"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// We test via AddEmpire indirectly, but we can also test scrubEmpireName
			// by verifying empire names stored after AddEmpire.
			// Since scrubEmpireName is unexported, we test behavior through AddEmpire.
			if tt.want == "" {
				return // Empty result tested by checking that AddEmpire fails
			}
			const dir = "/test/dir"
			const hwPlanetID domain.PlanetID = 400
			store := newMockStore()
			clusterStore := newMockClusterStore()
			svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
			store.games[dir] = domain.Game{
				ActiveHomeWorldID: hwPlanetID,
				Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
				Empires:           []domain.Empire{},
			}
			store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
			clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

			n, _, err := svc.AddEmpire(dir, 0, tt.input, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			game := store.games[dir]
			var gotName string
			for _, e := range game.Empires {
				if int(e.ID) == n {
					gotName = e.Name
					break
				}
			}
			if gotName != tt.want {
				t.Errorf("input=%q: expected %q, got %q", tt.input, tt.want, gotName)
			}
		})
	}

	// Test empty name fails
	t.Run("empty name fails", func(t *testing.T) {
		const dir = "/test/dir"
		const hwPlanetID domain.PlanetID = 400
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		store.games[dir] = domain.Game{
			ActiveHomeWorldID: hwPlanetID,
			Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
			Empires:           []domain.Empire{},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
		clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

		if _, _, err := svc.AddEmpire(dir, 0, "!@#$", 0); err == nil {
			t.Fatal("expected error for empty scrubbed name, got nil")
		}
	})
}

// --- TestRemoveEmpire ---

func TestRemoveEmpire(t *testing.T) {
	t.Run("sets active to false", func(t *testing.T) {
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		dir := "/test/dir"
		store.games[dir] = domain.Game{
			Empires: []domain.Empire{
				{ID: domain.EmpireID(10), Active: true},
			},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{
			"some-uuid": {Empire: 10},
		}}

		if err := svc.RemoveEmpire(dir, 10); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		game := store.games[dir]
		if game.Empires[0].Active {
			t.Error("expected empire to be inactive")
		}
	})

	t.Run("removes magic link", func(t *testing.T) {
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		dir := "/test/dir"
		store.games[dir] = domain.Game{
			Empires: []domain.Empire{
				{ID: domain.EmpireID(10), Active: true},
			},
		}
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
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		dir := "/test/dir"
		store.games[dir] = domain.Game{
			Empires: []domain.Empire{
				{ID: domain.EmpireID(10), Active: true},
			},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}

		if err := svc.RemoveEmpire(dir, 10); err != nil {
			t.Fatalf("unexpected error when no magic link: %v", err)
		}
	})

	t.Run("fails when empire not found", func(t *testing.T) {
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
		dir := "/test/dir"
		store.games[dir] = domain.Game{Empires: []domain.Empire{}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}

		if err := svc.RemoveEmpire(dir, 99); err == nil {
			t.Fatal("expected error for missing empire, got nil")
		}
	})
}

// --- TestCreateHomeWorld ---

func makeRichCluster() domain.Cluster {
	// Systems at various distances
	return domain.Cluster{
		Systems: []domain.System{
			{ID: 1, Location: domain.Coords{X: 0, Y: 0, Z: 0}},
			{ID: 2, Location: domain.Coords{X: 10, Y: 0, Z: 0}},
			{ID: 3, Location: domain.Coords{X: 1, Y: 0, Z: 0}}, // close to system 1
		},
		Stars: []domain.Star{
			{ID: 1, System: 1, Orbits: [10]domain.PlanetID{1}},
			{ID: 2, System: 2, Orbits: [10]domain.PlanetID{2}},
			{ID: 3, System: 3, Orbits: [10]domain.PlanetID{3}},
		},
		Planets: []domain.Planet{
			{ID: 1, Kind: domain.Terrestrial, Habitability: 0},
			{ID: 2, Kind: domain.Terrestrial, Habitability: 0},
			{ID: 3, Kind: domain.Terrestrial, Habitability: 0},
		},
	}
}

func TestCreateHomeWorld_AutoSelect(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{}
	clusterStore.clusters[dir] = makeRichCluster()

	planetID, err := svc.CreateHomeWorld(dir, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if planetID == 0 {
		t.Error("expected non-zero planetID")
	}
	game := store.games[dir]
	if len(game.Races) != 1 {
		t.Fatalf("expected 1 race, got %d", len(game.Races))
	}
	if game.ActiveHomeWorldID != planetID {
		t.Errorf("expected ActiveHomeWorldID %d, got %d", planetID, game.ActiveHomeWorldID)
	}
}

func TestCreateHomeWorld_PlanetFlag(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{}
	clusterStore.clusters[dir] = makeRichCluster()

	planetID, err := svc.CreateHomeWorld(dir, 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if planetID != 2 {
		t.Errorf("expected planetID 2, got %d", planetID)
	}
}

func TestCreateHomeWorld_PlanetFlagSkipsDistance(t *testing.T) {
	// A specified planet should succeed even if it's close to existing homeworlds
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	// Planet 3 is at system 3 (distance 1 from system 1), which would fail minDistance=3
	store.games[dir] = domain.Game{
		Races: []domain.Race{{ID: 1, HomeWorld: 1}},
	}
	clusterStore.clusters[dir] = makeRichCluster()

	// Planet 3 is close to planet 1's system, but explicitly specifying it should work
	planetID, err := svc.CreateHomeWorld(dir, 3, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if planetID != 3 {
		t.Errorf("expected planetID 3, got %d", planetID)
	}
}

func TestCreateHomeWorld_PlanetNotFound(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{}
	clusterStore.clusters[dir] = makeRichCluster()

	if _, err := svc.CreateHomeWorld(dir, 999, 0); err == nil {
		t.Fatal("expected error for non-existent planet, got nil")
	}
}

func TestCreateHomeWorld_NotTerrestrial(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{}
	c := makeRichCluster()
	c.Planets[0].Kind = domain.GasGiant // planet ID 1 is now a gas giant
	clusterStore.clusters[dir] = c

	if _, err := svc.CreateHomeWorld(dir, 1, 0); err == nil {
		t.Fatal("expected error for non-terrestrial planet, got nil")
	}
}

func TestCreateHomeWorld_AlreadyHomeWorld(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{
		Races: []domain.Race{{ID: 1, HomeWorld: 1}},
	}
	clusterStore.clusters[dir] = makeRichCluster()

	if _, err := svc.CreateHomeWorld(dir, 1, 0); err == nil {
		t.Fatal("expected error for already-homeworld planet, got nil")
	}
}

func TestCreateHomeWorld_MinDistance(t *testing.T) {
	// Planet 1 is at system 1 (0,0,0); planet 3 is at system 3 (1,0,0) — distance 1
	// Planet 2 is at system 2 (10,0,0) — distance 10 from system 1
	// With homeworld at planet 1, and minDistance=3:
	//   planet 3 should be rejected (distance 1 < 3)
	//   planet 2 should be accepted (distance 10 >= 3)
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{
		Races: []domain.Race{{ID: 1, HomeWorld: 1}},
	}
	clusterStore.clusters[dir] = makeRichCluster()

	planetID, err := svc.CreateHomeWorld(dir, 0, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if planetID == 3 {
		t.Error("expected planet 3 to be rejected (too close), but it was chosen")
	}
	if planetID != 2 {
		t.Errorf("expected planet 2 to be chosen, got %d", planetID)
	}
}

func TestCreateHomeWorld_NoTerrestrials(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameConfigService{Store: store, Cluster: clusterStore}
	store.games[dir] = domain.Game{}
	c := makeRichCluster()
	// Make all planets non-terrestrial
	for i := range c.Planets {
		c.Planets[i].Kind = domain.GasGiant
	}
	clusterStore.clusters[dir] = c

	if _, err := svc.CreateHomeWorld(dir, 0, 0); err == nil {
		t.Fatal("expected error when no terrestrials available, got nil")
	}
}
