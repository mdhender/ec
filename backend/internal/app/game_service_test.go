// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app_test

import (
	"errors"
	"testing"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/domain"
)

// mockGameStore is an in-memory GameStore for testing.
type mockGameStore struct {
	games         map[string]domain.Game
	authConfigs   map[string]domain.AuthConfig
	forceWriteErr error
}

func newMockStore() *mockGameStore {
	return &mockGameStore{
		games:       map[string]domain.Game{},
		authConfigs: map[string]domain.AuthConfig{},
	}
}

func (m *mockGameStore) ValidateDir(path string) error {
	return nil
}

func (m *mockGameStore) GameExists(dirPath string) (bool, error) {
	_, ok := m.games[dirPath]
	return ok, nil
}

func (m *mockGameStore) AuthConfigExists(dirPath string) (bool, error) {
	_, ok := m.authConfigs[dirPath]
	return ok, nil
}

func (m *mockGameStore) ReadGame(path string) (domain.Game, error) {
	game, ok := m.games[path]
	if !ok {
		return domain.Game{}, errors.New("game.json not found")
	}
	return game, nil
}

func (m *mockGameStore) WriteGame(path string, game domain.Game) error {
	if m.forceWriteErr != nil {
		return m.forceWriteErr
	}
	m.games[path] = game
	return nil
}

func (m *mockGameStore) ReadAuthConfig(path string) (domain.AuthConfig, error) {
	cfg, ok := m.authConfigs[path]
	if !ok {
		return domain.AuthConfig{}, errors.New("auth.json not found")
	}
	return cfg, nil
}

func (m *mockGameStore) WriteAuthConfig(path string, cfg domain.AuthConfig) error {
	if m.forceWriteErr != nil {
		return m.forceWriteErr
	}
	m.authConfigs[path] = cfg
	return nil
}

func (m *mockGameStore) CreateEmpireDir(dirPath string, empireNo int) error {
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
		svc := &app.GameService{Store: store, Cluster: clusterStore}
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
		svc := &app.GameService{Store: store, Cluster: clusterStore}

		if err := svc.CreateGame("/test/dir"); err == nil {
			t.Fatal("expected error when game.json exists, got nil")
		}
	})

	t.Run("fails if auth.json already exists", func(t *testing.T) {
		store := newMockStore()
		store.authConfigs["/test/dir"] = domain.AuthConfig{}
		clusterStore := newMockClusterStore()
		svc := &app.GameService{Store: store, Cluster: clusterStore}

		if err := svc.CreateGame("/test/dir"); err == nil {
			t.Fatal("expected error when auth.json exists, got nil")
		}
	})
}

// --- TestAddEmpire ---

func defaultColonyTemplate() domain.ColonyTemplate {
	return domain.ColonyTemplate{
		Kind:      domain.OpenAir,
		TechLevel: 1,
		Inventory: []domain.Inventory{
			{Unit: domain.Farm, TechLevel: 1, QuantityAssembled: 10},
			{Unit: domain.Mine, TechLevel: 1, QuantityAssembled: 20},
			{Unit: domain.Factory, TechLevel: 1, QuantityAssembled: 5},
		},
	}
}

// makeTestClusterWithDeposits builds a cluster with one system, one star,
// one terrestrial planet that has the given deposits pre-populated.
func makeTestClusterWithDeposits(planetID domain.PlanetID, depositIDs []domain.DepositID) domain.Cluster {
	deposits := make([]domain.Deposit, len(depositIDs))
	for i, id := range depositIDs {
		deposits[i] = domain.Deposit{
			ID:                id,
			Resource:          domain.METALLICS,
			YieldPct:          50,
			QuantityRemaining: 1000,
		}
	}
	return domain.Cluster{
		Systems: []domain.System{
			{ID: 1, Location: domain.Coords{X: 10, Y: 10, Z: 10}},
		},
		Stars: []domain.Star{
			{ID: 1, System: 1, Orbits: [10]domain.PlanetID{planetID}},
		},
		Planets: []domain.Planet{
			{ID: planetID, Kind: domain.Terrestrial, Habitability: 25, Deposits: depositIDs},
		},
		Deposits: deposits,
	}
}

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
		svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}
		store.games[dir] = domain.Game{
			ActiveHomeWorldID: hwPlanetID,
			Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
			Empires:           []domain.Empire{},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
		clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

		n, _, uuid, err := svc.AddEmpire(dir, 0, "TestEmpire", 0)
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
		svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}
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

		n, _, _, err := svc.AddEmpire(dir, 0, "TestEmpire", 0)
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
		svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}
		store.games[dir] = domain.Game{
			ActiveHomeWorldID: hwPlanetID,
			Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID, Empires: []domain.EmpireID{5}}},
			Empires:           []domain.Empire{{ID: 5, Active: true}},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
		clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

		if _, _, _, err := svc.AddEmpire(dir, 5, "TestEmpire", 0); err == nil {
			t.Fatal("expected error for duplicate empire, got nil")
		}
	})

	t.Run("generates magic link UUID", func(t *testing.T) {
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}
		store.games[dir] = domain.Game{
			ActiveHomeWorldID: hwPlanetID,
			Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
			Empires:           []domain.Empire{},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
		clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

		_, _, uuid, err := svc.AddEmpire(dir, 42, "TestEmpire", 0)
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}
	store.games[dir] = domain.Game{
		Empires: []domain.Empire{},
		Races:   []domain.Race{},
	}
	store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
	clusterStore.clusters[dir] = domain.Cluster{}

	if _, _, _, err := svc.AddEmpire(dir, 0, "TestEmpire", 0); err == nil {
		t.Fatal("expected error when no active homeworld, got nil")
	}
}

func TestAddEmpire_HomeWorldOverride(t *testing.T) {
	const dir = "/test/dir"
	const hwPlanetID domain.PlanetID = 200
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}
	store.games[dir] = domain.Game{
		ActiveHomeWorldID: 999, // different active homeworld
		Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
		Empires:           []domain.Empire{},
	}
	store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
	clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

	n, _, _, err := svc.AddEmpire(dir, 0, "TestEmpire", hwPlanetID)
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}
	store.games[dir] = domain.Game{
		Races:   []domain.Race{},
		Empires: []domain.Empire{},
	}
	store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
	clusterStore.clusters[dir] = domain.Cluster{}

	// Pass a homeWorldID that doesn't exist in game.Races
	if _, _, _, err := svc.AddEmpire(dir, 0, "TestEmpire", 999); err == nil {
		t.Fatal("expected error when homeworld not in game.Races, got nil")
	}
}

func TestAddEmpire_HomeWorldFull(t *testing.T) {
	const dir = "/test/dir"
	const hwPlanetID domain.PlanetID = 300
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}

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

	if _, _, _, err := svc.AddEmpire(dir, 0, "TestEmpire", 0); err == nil {
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
			svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}
			store.games[dir] = domain.Game{
				ActiveHomeWorldID: hwPlanetID,
				Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
				Empires:           []domain.Empire{},
			}
			store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
			clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

			n, _, _, err := svc.AddEmpire(dir, 0, tt.input, 0)
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
		svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()}}
		store.games[dir] = domain.Game{
			ActiveHomeWorldID: hwPlanetID,
			Races:             []domain.Race{{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID}},
			Empires:           []domain.Empire{},
		}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}
		clusterStore.clusters[dir] = makeTestCluster(hwPlanetID)

		if _, _, _, err := svc.AddEmpire(dir, 0, "!@#$", 0); err == nil {
			t.Fatal("expected error for empty scrubbed name, got nil")
		}
	})
}

// --- TestRemoveEmpire ---

func TestRemoveEmpire(t *testing.T) {
	t.Run("sets active to false", func(t *testing.T) {
		store := newMockStore()
		clusterStore := newMockClusterStore()
		svc := &app.GameService{Store: store, Cluster: clusterStore}
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
		svc := &app.GameService{Store: store, Cluster: clusterStore}
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
		svc := &app.GameService{Store: store, Cluster: clusterStore}
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
		svc := &app.GameService{Store: store, Cluster: clusterStore}
		dir := "/test/dir"
		store.games[dir] = domain.Game{Empires: []domain.Empire{}}
		store.authConfigs[dir] = domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}

		if err := svc.RemoveEmpire(dir, 99); err == nil {
			t.Fatal("expected error for missing empire, got nil")
		}
	})
}

// mockTemplateStore is an in-memory TemplateStore for testing.
type mockTemplateStore struct {
	homeworldTemplate domain.HomeworldTemplate
	colonyTemplate    domain.ColonyTemplate
	forceErr          error
}

func (m *mockTemplateStore) ReadHomeworldTemplate(dataPath string) (domain.HomeworldTemplate, error) {
	if m.forceErr != nil {
		return domain.HomeworldTemplate{}, m.forceErr
	}
	return m.homeworldTemplate, nil
}

func (m *mockTemplateStore) ReadColonyTemplate(dataPath string) (domain.ColonyTemplate, error) {
	if m.forceErr != nil {
		return domain.ColonyTemplate{}, m.forceErr
	}
	return m.colonyTemplate, nil
}

func defaultHomeworldTemplate() domain.HomeworldTemplate {
	return domain.HomeworldTemplate{
		Habitability: 25,
		Deposits: []domain.DepositTemplate{
			{Resource: domain.FUEL, YieldPct: 50, QuantityRemaining: 100},
		},
	}
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}}
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}}
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}}
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}}
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}}
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}}
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}}
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
	svc := &app.GameService{Store: store, Cluster: clusterStore, Templates: &mockTemplateStore{homeworldTemplate: defaultHomeworldTemplate()}}
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

func TestCreateHomeWorld_AppliesTemplate(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()

	tmpl := domain.HomeworldTemplate{
		Habitability: 30,
		Deposits: []domain.DepositTemplate{
			{Resource: domain.METALLICS, YieldPct: 60, QuantityRemaining: 500},
			{Resource: domain.NONMETALLICS, YieldPct: 40, QuantityRemaining: 300},
		},
	}
	svc := &app.GameService{
		Store:     store,
		Cluster:   clusterStore,
		Templates: &mockTemplateStore{homeworldTemplate: tmpl},
	}

	// Planet 1 with two pre-existing deposits (IDs 10, 11)
	cluster := domain.Cluster{
		Systems: []domain.System{{ID: 1, Location: domain.Coords{X: 0, Y: 0, Z: 0}}},
		Stars:   []domain.Star{{ID: 1, System: 1, Orbits: [10]domain.PlanetID{1}}},
		Planets: []domain.Planet{
			{ID: 1, Kind: domain.Terrestrial, Habitability: 0, Deposits: []domain.DepositID{10, 11}},
		},
		Deposits: []domain.Deposit{
			{ID: 10, Resource: domain.GOLD, YieldPct: 10, QuantityRemaining: 50},
			{ID: 11, Resource: domain.FUEL, YieldPct: 20, QuantityRemaining: 75},
		},
	}
	store.games[dir] = domain.Game{}
	clusterStore.clusters[dir] = cluster

	_, err := svc.CreateHomeWorld(dir, 1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	saved := clusterStore.clusters[dir]

	// Find planet 1
	var p domain.Planet
	for _, pl := range saved.Planets {
		if pl.ID == 1 {
			p = pl
			break
		}
	}

	if p.Habitability != 30 {
		t.Errorf("Habitability: got %d, want 30", p.Habitability)
	}
	if len(p.Deposits) != 2 {
		t.Fatalf("planet.Deposits length: got %d, want 2", len(p.Deposits))
	}
	// New deposit IDs must be greater than the old max (11)
	for _, did := range p.Deposits {
		if int(did) <= 11 {
			t.Errorf("expected new deposit ID > 11, got %d", did)
		}
	}
	// Old deposit records (IDs 10, 11) must be gone
	for _, d := range saved.Deposits {
		if d.ID == 10 || d.ID == 11 {
			t.Errorf("expected old deposit ID %d to be removed, but it still exists", d.ID)
		}
	}
	// New deposits must match template
	if len(saved.Deposits) != 2 {
		t.Fatalf("cluster.Deposits length: got %d, want 2", len(saved.Deposits))
	}
	if saved.Deposits[0].Resource != domain.METALLICS || saved.Deposits[0].YieldPct != 60 {
		t.Errorf("first deposit: got %v %d%%, want METALLICS 60%%", saved.Deposits[0].Resource, saved.Deposits[0].YieldPct)
	}
	if saved.Deposits[1].Resource != domain.NONMETALLICS || saved.Deposits[1].YieldPct != 40 {
		t.Errorf("second deposit: got %v %d%%, want NONMETALLICS 40%%", saved.Deposits[1].Resource, saved.Deposits[1].YieldPct)
	}
}

func TestCreateHomeWorld_TemplateError(t *testing.T) {
	const dir = "/test/dir"
	store := newMockStore()
	clusterStore := newMockClusterStore()
	svc := &app.GameService{
		Store:     store,
		Cluster:   clusterStore,
		Templates: &mockTemplateStore{forceErr: errors.New("template read failure")},
	}
	store.games[dir] = domain.Game{}
	clusterStore.clusters[dir] = makeRichCluster()

	if _, err := svc.CreateHomeWorld(dir, 1, 0); err == nil {
		t.Fatal("expected error when template read fails, got nil")
	}
}

func TestAddEmpire_ColonyFromTemplate(t *testing.T) {
	const hwPlanetID = domain.PlanetID(100)
	store := newMockStore()
	_ = store.WriteGame(".", domain.Game{})
	_ = store.WriteAuthConfig(".", domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}})

	clusterStore := newMockClusterStore()
	cluster := makeTestCluster(hwPlanetID)
	_ = clusterStore.WriteCluster(".", cluster, true)

	tmpl := defaultColonyTemplate()
	svc := &app.GameService{
		Store:     store,
		Cluster:   clusterStore,
		Templates: &mockTemplateStore{colonyTemplate: tmpl},
	}

	// Set up a race/homeworld first
	game, _ := store.ReadGame(".")
	game.Races = append(game.Races, domain.Race{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID})
	game.ActiveHomeWorldID = hwPlanetID
	_ = store.WriteGame(".", game)

	_, _, _, err := svc.AddEmpire(".", 0, "Test Empire", hwPlanetID)
	if err != nil {
		t.Fatalf("AddEmpire error: %v", err)
	}

	got, _ := clusterStore.ReadCluster(".")
	if len(got.Colonies) != 1 {
		t.Fatalf("expected 1 colony, got %d", len(got.Colonies))
	}
	c := got.Colonies[0]
	if c.Planet != hwPlanetID {
		t.Errorf("colony.Planet = %d, want %d", c.Planet, hwPlanetID)
	}
	if c.Kind != domain.OpenAir {
		t.Errorf("colony.Kind = %v, want OpenAir", c.Kind)
	}
	if c.TechLevel != 1 {
		t.Errorf("colony.TechLevel = %d, want 1", c.TechLevel)
	}
	if len(c.Inventory) != 3 {
		t.Errorf("colony.Inventory len = %d, want 3", len(c.Inventory))
	}
	// Verify deep copy: modifying returned colony inventory should not affect template
	if len(c.Inventory) > 0 {
		origQty := tmpl.Inventory[0].QuantityAssembled
		c.Inventory[0].QuantityAssembled = 999
		if tmpl.Inventory[0].QuantityAssembled != origQty {
			t.Error("colony.Inventory is not a deep copy — template was modified")
		}
	}
}

func TestAddEmpire_FarmGroup(t *testing.T) {
	const hwPlanetID = domain.PlanetID(100)
	store := newMockStore()
	_ = store.WriteGame(".", domain.Game{})
	_ = store.WriteAuthConfig(".", domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}})

	clusterStore := newMockClusterStore()
	_ = clusterStore.WriteCluster(".", makeTestCluster(hwPlanetID), true)

	tmpl := domain.ColonyTemplate{
		Kind:      domain.OpenAir,
		TechLevel: 1,
		Inventory: []domain.Inventory{
			{Unit: domain.Farm, TechLevel: 1, QuantityAssembled: 10},
			{Unit: domain.Farm, TechLevel: 2, QuantityAssembled: 5},
			{Unit: domain.Mine, TechLevel: 1, QuantityAssembled: 8},
		},
	}
	svc := &app.GameService{
		Store:     store,
		Cluster:   clusterStore,
		Templates: &mockTemplateStore{colonyTemplate: tmpl},
	}

	game, _ := store.ReadGame(".")
	game.Races = append(game.Races, domain.Race{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID})
	game.ActiveHomeWorldID = hwPlanetID
	_ = store.WriteGame(".", game)

	_, _, _, err := svc.AddEmpire(".", 0, "Test Empire", hwPlanetID)
	if err != nil {
		t.Fatalf("AddEmpire error: %v", err)
	}

	got, _ := clusterStore.ReadCluster(".")
	if len(got.Colonies) != 1 {
		t.Fatalf("expected 1 colony, got %d", len(got.Colonies))
	}
	c := got.Colonies[0]
	if len(c.FarmGroups) != 1 {
		t.Fatalf("expected 1 FarmGroup, got %d", len(c.FarmGroups))
	}
	fg := c.FarmGroups[0]
	if fg.ID != 1 {
		t.Errorf("FarmGroup.ID = %d, want 1", fg.ID)
	}
	if len(fg.Units) != 2 {
		t.Fatalf("expected 2 farm sub-groups, got %d", len(fg.Units))
	}
	if fg.Units[0].TechLevel != 1 || fg.Units[0].Quantity != 10 {
		t.Errorf("farm sub-group[0] = {TL:%d Qty:%d}, want {TL:1 Qty:10}", fg.Units[0].TechLevel, fg.Units[0].Quantity)
	}
	if fg.Units[1].TechLevel != 2 || fg.Units[1].Quantity != 5 {
		t.Errorf("farm sub-group[1] = {TL:%d Qty:%d}, want {TL:2 Qty:5}", fg.Units[1].TechLevel, fg.Units[1].Quantity)
	}
}

func TestAddEmpire_FarmGroup_NoFarms(t *testing.T) {
	const hwPlanetID = domain.PlanetID(100)
	store := newMockStore()
	_ = store.WriteGame(".", domain.Game{})
	_ = store.WriteAuthConfig(".", domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}})

	clusterStore := newMockClusterStore()
	_ = clusterStore.WriteCluster(".", makeTestCluster(hwPlanetID), true)

	tmpl := domain.ColonyTemplate{
		Kind:      domain.OpenAir,
		TechLevel: 1,
		Inventory: []domain.Inventory{
			{Unit: domain.Mine, TechLevel: 1, QuantityAssembled: 10},
		},
	}
	svc := &app.GameService{
		Store:     store,
		Cluster:   clusterStore,
		Templates: &mockTemplateStore{colonyTemplate: tmpl},
	}

	game, _ := store.ReadGame(".")
	game.Races = append(game.Races, domain.Race{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID})
	game.ActiveHomeWorldID = hwPlanetID
	_ = store.WriteGame(".", game)

	_, _, _, err := svc.AddEmpire(".", 0, "Test Empire", hwPlanetID)
	if err != nil {
		t.Fatalf("AddEmpire error: %v", err)
	}

	got, _ := clusterStore.ReadCluster(".")
	c := got.Colonies[0]
	if len(c.FarmGroups) != 0 {
		t.Errorf("expected no FarmGroups, got %d", len(c.FarmGroups))
	}
}

func TestBuildMiningGroups(t *testing.T) {
	tests := []struct {
		name       string
		inventory  []domain.Inventory
		depositIDs []domain.DepositID
		wantNil    bool
		wantGroups []struct {
			depositID domain.DepositID
			units     []domain.GroupUnit
		}
	}{
		{
			name: "even split",
			inventory: []domain.Inventory{
				{Unit: domain.Mine, TechLevel: 1, QuantityAssembled: 20},
			},
			depositIDs: []domain.DepositID{1, 2, 3},
			wantGroups: []struct {
				depositID domain.DepositID
				units     []domain.GroupUnit
			}{
				{1, []domain.GroupUnit{{TechLevel: 1, Quantity: 7}}},
				{2, []domain.GroupUnit{{TechLevel: 1, Quantity: 7}}},
				{3, []domain.GroupUnit{{TechLevel: 1, Quantity: 6}}},
			},
		},
		{
			name: "remainder round-robin",
			inventory: []domain.Inventory{
				{Unit: domain.Mine, TechLevel: 1, QuantityAssembled: 10},
			},
			depositIDs: []domain.DepositID{1, 2, 3},
			wantGroups: []struct {
				depositID domain.DepositID
				units     []domain.GroupUnit
			}{
				{1, []domain.GroupUnit{{TechLevel: 1, Quantity: 4}}},
				{2, []domain.GroupUnit{{TechLevel: 1, Quantity: 3}}},
				{3, []domain.GroupUnit{{TechLevel: 1, Quantity: 3}}},
			},
		},
		{
			name: "multi-TL greedy",
			inventory: []domain.Inventory{
				{Unit: domain.Mine, TechLevel: 1, QuantityAssembled: 5},
				{Unit: domain.Mine, TechLevel: 2, QuantityAssembled: 5},
			},
			depositIDs: []domain.DepositID{1, 2},
			wantGroups: []struct {
				depositID domain.DepositID
				units     []domain.GroupUnit
			}{
				{1, []domain.GroupUnit{{TechLevel: 1, Quantity: 5}}},
				{2, []domain.GroupUnit{{TechLevel: 2, Quantity: 5}}},
			},
		},
		{
			name:       "no deposits",
			inventory:  []domain.Inventory{{Unit: domain.Mine, TechLevel: 1, QuantityAssembled: 10}},
			depositIDs: nil,
			wantNil:    true,
		},
		{
			name:       "no mines",
			inventory:  []domain.Inventory{{Unit: domain.Farm, TechLevel: 1, QuantityAssembled: 10}},
			depositIDs: []domain.DepositID{1, 2, 3},
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.BuildMiningGroupsForTest(tt.inventory, tt.depositIDs)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tt.wantGroups) {
				t.Fatalf("expected %d groups, got %d", len(tt.wantGroups), len(got))
			}
			for i, wg := range tt.wantGroups {
				if got[i].Deposit != wg.depositID {
					t.Errorf("group[%d].Deposit = %d, want %d", i, got[i].Deposit, wg.depositID)
				}
				if got[i].ID != domain.MiningGroupID(i+1) {
					t.Errorf("group[%d].ID = %d, want %d", i, got[i].ID, i+1)
				}
				if len(got[i].Units) != len(wg.units) {
					t.Fatalf("group[%d] units len = %d, want %d", i, len(got[i].Units), len(wg.units))
				}
				for j, wu := range wg.units {
					if got[i].Units[j].TechLevel != wu.TechLevel || got[i].Units[j].Quantity != wu.Quantity {
						t.Errorf("group[%d].Units[%d] = {TL:%d Qty:%d}, want {TL:%d Qty:%d}",
							i, j, got[i].Units[j].TechLevel, got[i].Units[j].Quantity,
							wu.TechLevel, wu.Quantity)
					}
				}
			}
		})
	}
}

func TestAddEmpire_MiningGroups(t *testing.T) {
	const hwPlanetID = domain.PlanetID(100)
	depositIDs := []domain.DepositID{10, 11, 12}

	store := newMockStore()
	_ = store.WriteGame(".", domain.Game{})
	_ = store.WriteAuthConfig(".", domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}})

	clusterStore := newMockClusterStore()
	_ = clusterStore.WriteCluster(".", makeTestClusterWithDeposits(hwPlanetID, depositIDs), true)

	tmpl := domain.ColonyTemplate{
		Kind:      domain.OpenAir,
		TechLevel: 1,
		Inventory: []domain.Inventory{
			{Unit: domain.Mine, TechLevel: 1, QuantityAssembled: 9},
		},
	}
	svc := &app.GameService{
		Store:     store,
		Cluster:   clusterStore,
		Templates: &mockTemplateStore{colonyTemplate: tmpl},
	}

	game, _ := store.ReadGame(".")
	game.Races = append(game.Races, domain.Race{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID})
	game.ActiveHomeWorldID = hwPlanetID
	_ = store.WriteGame(".", game)

	_, _, _, err := svc.AddEmpire(".", 0, "Test Empire", hwPlanetID)
	if err != nil {
		t.Fatalf("AddEmpire error: %v", err)
	}

	got, _ := clusterStore.ReadCluster(".")
	c := got.Colonies[0]
	if len(c.MiningGroups) != 3 {
		t.Fatalf("expected 3 MiningGroups, got %d", len(c.MiningGroups))
	}
	for i, depID := range depositIDs {
		mg := c.MiningGroups[i]
		if mg.Deposit != depID {
			t.Errorf("MiningGroup[%d].Deposit = %d, want %d", i, mg.Deposit, depID)
		}
		if mg.ID != domain.MiningGroupID(i+1) {
			t.Errorf("MiningGroup[%d].ID = %d, want %d", i, mg.ID, i+1)
		}
		if len(mg.Units) != 1 || mg.Units[0].TechLevel != 1 || mg.Units[0].Quantity != 3 {
			t.Errorf("MiningGroup[%d].Units = %v, want [{TL:1 Qty:3}]", i, mg.Units)
		}
	}
	// Verify colony.Inventory not modified
	if c.Inventory[0].QuantityAssembled != 9 {
		t.Errorf("colony.Inventory[0].QuantityAssembled = %d, want 9 (must not be mutated)", c.Inventory[0].QuantityAssembled)
	}
}

func TestAddEmpire_MiningGroups_NoDeposits(t *testing.T) {
	const hwPlanetID = domain.PlanetID(100)

	store := newMockStore()
	_ = store.WriteGame(".", domain.Game{})
	_ = store.WriteAuthConfig(".", domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}})

	clusterStore := newMockClusterStore()
	_ = clusterStore.WriteCluster(".", makeTestCluster(hwPlanetID), true)

	svc := &app.GameService{
		Store:     store,
		Cluster:   clusterStore,
		Templates: &mockTemplateStore{colonyTemplate: defaultColonyTemplate()},
	}

	game, _ := store.ReadGame(".")
	game.Races = append(game.Races, domain.Race{ID: domain.RaceID(hwPlanetID), HomeWorld: hwPlanetID})
	game.ActiveHomeWorldID = hwPlanetID
	_ = store.WriteGame(".", game)

	_, _, _, err := svc.AddEmpire(".", 0, "Test Empire", hwPlanetID)
	if err != nil {
		t.Fatalf("AddEmpire error: %v", err)
	}

	got, _ := clusterStore.ReadCluster(".")
	c := got.Colonies[0]
	if c.MiningGroups != nil {
		t.Errorf("expected nil MiningGroups, got %v", c.MiningGroups)
	}
}
