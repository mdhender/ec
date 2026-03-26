// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import (
	"cmp"
	cryptorand "crypto/rand"
	"fmt"
	mathrand "math/rand/v2"
	"slices"
	"strings"
	"unicode"

	"github.com/mdhender/ec/internal/domain"
)

// GameService orchestrates game and auth config use cases.
type GameService struct {
	Store     GameStore
	Cluster   ClusterStore
	Templates TemplateStore
}

// CreateGame initializes an empty game.json and auth.json in dirPath.
// Returns an error if either file already exists or if dirPath is not a directory.
func (s *GameService) CreateGame(dirPath string) error {
	if err := s.Store.ValidateDir(dirPath); err != nil {
		return fmt.Errorf("createGame: %w", err)
	}
	if exists, err := s.Store.GameExists(dirPath); err != nil {
		return fmt.Errorf("createGame: %w", err)
	} else if exists {
		return fmt.Errorf("createGame: game.json already exists in %q", dirPath)
	}
	if exists, err := s.Store.AuthConfigExists(dirPath); err != nil {
		return fmt.Errorf("createGame: %w", err)
	} else if exists {
		return fmt.Errorf("createGame: auth.json already exists in %q", dirPath)
	}

	if err := s.Store.WriteGame(dirPath, domain.Game{Empires: []domain.Empire{}}); err != nil {
		return fmt.Errorf("createGame: %w", err)
	}
	if err := s.Store.WriteAuthConfig(dirPath, domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}); err != nil {
		return fmt.Errorf("createGame: %w", err)
	}
	return nil
}

// AddEmpire adds a new empire to game.json and a magic link to auth.json.
// If empireNo is 0, the next empire number is auto-assigned.
// homeWorldID: if 0, uses game.ActiveHomeWorldID; if still 0, error.
// Returns the assigned empire number, the scrubbed name, and the generated magic link UUID.
func (s *GameService) AddEmpire(dirPath string, empireNo int, name string, homeWorldID domain.PlanetID) (int, string, string, error) {
	game, err := s.Store.ReadGame(dirPath)
	if err != nil {
		return 0, "", "", fmt.Errorf("addEmpire: %w", err)
	}
	cluster, err := s.Cluster.ReadCluster(dirPath)
	if err != nil {
		return 0, "", "", fmt.Errorf("addEmpire: %w", err)
	}

	// Resolve homeworld
	if homeWorldID == 0 {
		homeWorldID = game.ActiveHomeWorldID
	}
	if homeWorldID == 0 {
		return 0, "", "", fmt.Errorf("addEmpire: no active homeworld; use --homeworld or run create homeworld first")
	}

	// Ordering invariant: CreateHomeWorld must have been called for this
	// planet before AddEmpire can assign empires to it. The race lookup
	// below enforces this — if no Race exists with HomeWorld == homeWorldID,
	// the operation fails.
	raceIdx := -1
	for i, r := range game.Races {
		if r.HomeWorld == homeWorldID {
			raceIdx = i
			break
		}
	}
	if raceIdx == -1 {
		return 0, "", "", fmt.Errorf("addEmpire: homeworld %d does not exist", homeWorldID)
	}
	race := game.Races[raceIdx]

	// Check empire limit per race
	if len(race.Empires) >= 25 {
		return 0, "", "", fmt.Errorf("addEmpire: homeworld %d is full (25 empires)", homeWorldID)
	}

	// Scrub empire name
	scrubbedName := scrubEmpireName(name)
	if scrubbedName == "" {
		return 0, "", "", fmt.Errorf("addEmpire: empire name is required")
	}

	// Auto-assign empire number if 0
	if empireNo == 0 {
		max := 0
		for _, e := range game.Empires {
			if int(e.ID) > max {
				max = int(e.ID)
			}
		}
		empireNo = max + 1
	}

	// Check for duplicate empire number
	for _, e := range game.Empires {
		if int(e.ID) == empireNo {
			return 0, "", "", fmt.Errorf("addEmpire: empire %d already exists", empireNo)
		}
	}

	// Read and validate colony template
	colonyTmpl, err := s.Templates.ReadColonyTemplate(dirPath)
	if err != nil {
		return 0, "", "", fmt.Errorf("addEmpire: %w", err)
	}
	if colonyTmpl.Kind < domain.OpenAir || colonyTmpl.Kind > domain.Enclosed {
		return 0, "", "", fmt.Errorf("addEmpire: invalid colony template kind %d", colonyTmpl.Kind)
	}
	if colonyTmpl.TechLevel <= 0 {
		return 0, "", "", fmt.Errorf("addEmpire: invalid colony template tech level %d (must be > 0)", colonyTmpl.TechLevel)
	}
	for i, inv := range colonyTmpl.Inventory {
		if inv.QuantityAssembled < 0 {
			return 0, "", "", fmt.Errorf("addEmpire: colony template inventory %d has negative assembled quantity", i)
		}
	}

	// Create starting colony (use max existing ID + 1 to avoid duplicates if colonies are ever deleted)
	maxColonyID := 0
	for _, c := range cluster.Colonies {
		if int(c.ID) > maxColonyID {
			maxColonyID = int(c.ID)
		}
	}

	// Deep-copy inventory from template so each empire gets independent slices.
	inventory := make([]domain.Inventory, len(colonyTmpl.Inventory))
	copy(inventory, colonyTmpl.Inventory)

	colony := domain.Colony{
		ID:        domain.ColonyID(maxColonyID + 1),
		Empire:    domain.EmpireID(empireNo),
		Planet:    homeWorldID,
		Kind:      colonyTmpl.Kind,
		TechLevel: colonyTmpl.TechLevel,
		Inventory: inventory,
	}

	// Build farm group from assembled Farm units in the colony inventory.
	// Each colony has at most one FarmGroup; sub-groups are by tech level.
	// Aggregate by TechLevel in case multiple inventory rows share a TL.
	farmByTL := make(map[domain.TechLevel]int)
	for _, inv := range colony.Inventory {
		if inv.Unit == domain.Farm && inv.QuantityAssembled > 0 {
			farmByTL[inv.TechLevel] += inv.QuantityAssembled
		}
	}
	var farmUnits []domain.GroupUnit
	for tl, qty := range farmByTL {
		farmUnits = append(farmUnits, domain.GroupUnit{TechLevel: tl, Quantity: qty})
	}
	slices.SortFunc(farmUnits, func(a, b domain.GroupUnit) int {
		return cmp.Compare(a.TechLevel, b.TechLevel)
	})
	if len(farmUnits) > 0 {
		colony.FarmGroups = []domain.FarmGroup{
			{ID: 1, Units: farmUnits},
		}
	}

	// Build mining groups — one per deposit, Mine units split evenly.
	var hwDepositIDs []domain.DepositID
	hwPlanetFound := false
	for _, p := range cluster.Planets {
		if p.ID == homeWorldID {
			hwDepositIDs = p.Deposits
			hwPlanetFound = true
			break
		}
	}
	if !hwPlanetFound {
		return 0, "", "", fmt.Errorf("addEmpire: homeworld planet %d not found in cluster", homeWorldID)
	}
	colony.MiningGroups = buildMiningGroups(colony.Inventory, hwDepositIDs)

	cluster.Colonies = append(cluster.Colonies, colony)

	// Create empire
	empire := domain.Empire{
		ID:        domain.EmpireID(empireNo),
		Name:      scrubbedName,
		Active:    true,
		Race:      race.ID,
		HomeWorld: homeWorldID,
		Colonies:  []domain.ColonyID{colony.ID},
	}
	game.Empires = append(game.Empires, empire)

	// Append empire ID to race
	game.Races[raceIdx].Empires = append(game.Races[raceIdx].Empires, domain.EmpireID(empireNo))

	// Write game.json
	if err := s.Store.WriteGame(dirPath, game); err != nil {
		return 0, "", "", fmt.Errorf("addEmpire: %w", err)
	}

	// Write cluster.json
	if err := s.Cluster.WriteCluster(dirPath, cluster, true); err != nil {
		return 0, "", "", fmt.Errorf("addEmpire: %w", err)
	}

	// Create empire directory
	if err := s.Store.CreateEmpireDir(dirPath, empireNo); err != nil {
		return 0, "", "", fmt.Errorf("addEmpire: %w", err)
	}

	// Generate magic link
	authCfg, err := s.Store.ReadAuthConfig(dirPath)
	if err != nil {
		return 0, "", "", fmt.Errorf("addEmpire: %w", err)
	}
	uuid, err := newUUID()
	if err != nil {
		return 0, "", "", fmt.Errorf("addEmpire: %w", err)
	}
	if authCfg.MagicLinks == nil {
		authCfg.MagicLinks = map[string]domain.AuthLink{}
	}
	authCfg.MagicLinks[uuid] = domain.AuthLink{Empire: empireNo}
	if err := s.Store.WriteAuthConfig(dirPath, authCfg); err != nil {
		return 0, "", "", fmt.Errorf("addEmpire: %w", err)
	}

	return empireNo, scrubbedName, uuid, nil
}

// RemoveEmpire sets the empire's Active flag to false and removes its magic links.
func (s *GameService) RemoveEmpire(dirPath string, empireNo int) error {
	game, err := s.Store.ReadGame(dirPath)
	if err != nil {
		return fmt.Errorf("removeEmpire: %w", err)
	}

	found := false
	for i, e := range game.Empires {
		if int(e.ID) == empireNo {
			game.Empires[i].Active = false
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("removeEmpire: empire %d not found", empireNo)
	}
	if err := s.Store.WriteGame(dirPath, game); err != nil {
		return fmt.Errorf("removeEmpire: %w", err)
	}

	authCfg, err := s.Store.ReadAuthConfig(dirPath)
	if err != nil {
		return fmt.Errorf("removeEmpire: %w", err)
	}
	for uuid, link := range authCfg.MagicLinks {
		if link.Empire == empireNo {
			delete(authCfg.MagicLinks, uuid)
		}
	}
	if err := s.Store.WriteAuthConfig(dirPath, authCfg); err != nil {
		return fmt.Errorf("removeEmpire: %w", err)
	}
	return nil
}

// ShowMagicLink returns the magic link UUID for the given empire.
func (s *GameService) ShowMagicLink(dirPath string, empireNo int) (string, error) {
	authCfg, err := s.Store.ReadAuthConfig(dirPath)
	if err != nil {
		return "", fmt.Errorf("showMagicLink: %w", err)
	}
	for uuid, link := range authCfg.MagicLinks {
		if link.Empire == empireNo {
			return uuid, nil
		}
	}
	return "", fmt.Errorf("showMagicLink: no magic link found for empire %d", empireNo)
}

// CreateHomeWorld selects or validates a homeworld planet and records a new Race.
// If planetID != 0, that specific planet is used.
// If planetID == 0, a terrestrial planet is auto-selected that is at least minDistance from existing homeworlds.
func (s *GameService) CreateHomeWorld(dataPath string, planetID domain.PlanetID, minDistance int) (domain.PlanetID, error) {
	game, err := s.Store.ReadGame(dataPath)
	if err != nil {
		return 0, fmt.Errorf("createHomeWorld: %w", err)
	}
	cluster, err := s.Cluster.ReadCluster(dataPath)
	if err != nil {
		return 0, fmt.Errorf("createHomeWorld: %w", err)
	}

	// Collect existing homeworld planet IDs
	existingHomeWorlds := make(map[domain.PlanetID]bool)
	for _, r := range game.Races {
		existingHomeWorlds[r.HomeWorld] = true
	}

	if planetID != 0 {
		// Explicit planet: validate it
		found := false
		for _, p := range cluster.Planets {
			if p.ID == planetID {
				found = true
				if p.Kind != domain.Terrestrial {
					return 0, fmt.Errorf("createHomeWorld: planet %d is not terrestrial", planetID)
				}
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("createHomeWorld: planet %d not found", planetID)
		}
		if existingHomeWorlds[planetID] {
			return 0, fmt.Errorf("createHomeWorld: planet %d is already a homeworld", planetID)
		}
	} else {
		// Auto-select: collect candidate terrestrial planets not already homeworlds
		type candidate struct {
			planet   domain.Planet
			location domain.Coords
		}
		var candidates []candidate
		for _, p := range cluster.Planets {
			if p.Kind != domain.Terrestrial {
				continue
			}
			if existingHomeWorlds[p.ID] {
				continue
			}
			loc, err := findSystemForPlanet(cluster, p.ID)
			if err != nil {
				continue
			}
			// Check min distance from all existing homeworlds
			tooClose := false
			for hwID := range existingHomeWorlds {
				hwLoc, err := findSystemForPlanet(cluster, hwID)
				if err != nil {
					continue
				}
				if loc.Distance(hwLoc) < float64(minDistance) {
					tooClose = true
					break
				}
			}
			if !tooClose {
				candidates = append(candidates, candidate{planet: p, location: loc})
			}
		}
		if len(candidates) == 0 {
			return 0, fmt.Errorf("createHomeWorld: no available terrestrial planets meet the distance requirement")
		}
		chosen := candidates[mathrand.IntN(len(candidates))]
		planetID = chosen.planet.ID
	}

	// Load homeworld template and apply it
	tmpl, err := s.Templates.ReadHomeworldTemplate(dataPath)
	if err != nil {
		return 0, fmt.Errorf("createHomeWorld: %w", err)
	}
	if tmpl.Habitability < 0 || tmpl.Habitability > 100 {
		return 0, fmt.Errorf("createHomeWorld: invalid template habitability %d (must be 0–100)", tmpl.Habitability)
	}
	if len(tmpl.Deposits) == 0 {
		return 0, fmt.Errorf("createHomeWorld: template has no deposits")
	}
	for i, dt := range tmpl.Deposits {
		if dt.Resource < domain.GOLD || dt.Resource > domain.NONMETALLICS {
			return 0, fmt.Errorf("createHomeWorld: template deposit %d has invalid resource %d", i, dt.Resource)
		}
		if dt.YieldPct < 0 || dt.YieldPct > 100 {
			return 0, fmt.Errorf("createHomeWorld: template deposit %d has invalid yield %d%% (must be 0–100)", i, dt.YieldPct)
		}
		if dt.QuantityRemaining < 0 {
			return 0, fmt.Errorf("createHomeWorld: template deposit %d has negative quantity", i)
		}
	}

	// Find max deposit ID across all existing deposits
	maxDepositID := 0
	for _, d := range cluster.Deposits {
		if int(d.ID) > maxDepositID {
			maxDepositID = int(d.ID)
		}
	}

	// Find planet index (must use index to mutate in place)
	planetIdx := -1
	for i := range cluster.Planets {
		if cluster.Planets[i].ID == planetID {
			planetIdx = i
			break
		}
	}
	if planetIdx == -1 {
		return 0, fmt.Errorf("createHomeWorld: planet %d not found in cluster", planetID)
	}

	// Remove old deposits for this planet
	oldIDs := make(map[domain.DepositID]bool, len(cluster.Planets[planetIdx].Deposits))
	for _, did := range cluster.Planets[planetIdx].Deposits {
		oldIDs[did] = true
	}
	var filtered []domain.Deposit
	for _, d := range cluster.Deposits {
		if !oldIDs[d.ID] {
			filtered = append(filtered, d)
		}
	}
	cluster.Deposits = filtered
	cluster.Planets[planetIdx].Deposits = nil

	// Add template deposits with fresh IDs
	for _, dt := range tmpl.Deposits {
		maxDepositID++
		cluster.Deposits = append(cluster.Deposits, domain.Deposit{
			ID:                domain.DepositID(maxDepositID),
			Resource:          dt.Resource,
			YieldPct:          dt.YieldPct,
			QuantityRemaining: dt.QuantityRemaining,
		})
		cluster.Planets[planetIdx].Deposits = append(cluster.Planets[planetIdx].Deposits, domain.DepositID(maxDepositID))
	}

	// Set habitability from template
	cluster.Planets[planetIdx].Habitability = tmpl.Habitability

	// Add a new Race
	race := domain.Race{
		ID:        domain.RaceID(planetID),
		HomeWorld: planetID,
		Empires:   nil,
	}
	game.Races = append(game.Races, race)
	game.ActiveHomeWorldID = planetID

	// Write updated files
	if err := s.Store.WriteGame(dataPath, game); err != nil {
		return 0, fmt.Errorf("createHomeWorld: %w", err)
	}
	if err := s.Cluster.WriteCluster(dataPath, cluster, true); err != nil {
		return 0, fmt.Errorf("createHomeWorld: %w", err)
	}

	return planetID, nil
}

// findSystemForPlanet walks cluster.Stars to find which star has planetID in its Orbits,
// then looks up the system by the star's System field.
func findSystemForPlanet(cluster domain.Cluster, planetID domain.PlanetID) (domain.Coords, error) {
	// Find the star that has this planet in its orbits
	var starSystemID domain.SystemID
	found := false
	for _, star := range cluster.Stars {
		for _, orbitPlanetID := range star.Orbits {
			if orbitPlanetID == planetID {
				starSystemID = star.System
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return domain.Coords{}, fmt.Errorf("planet %d not found in any star's orbits", planetID)
	}

	// Look up the system by ID
	for _, sys := range cluster.Systems {
		if sys.ID == starSystemID {
			return sys.Location, nil
		}
	}
	return domain.Coords{}, fmt.Errorf("system %d not found for planet %d", starSystemID, planetID)
}

// scrubEmpireName removes disallowed characters, compresses spaces, and trims.
func scrubEmpireName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' || r == '.' || r == ',' {
			b.WriteRune(r)
		}
	}
	// Compress runs of spaces
	parts := strings.Fields(b.String())
	return strings.TrimSpace(strings.Join(parts, " "))
}

// buildMiningGroups creates one MiningGroup per deposit, distributing
// assembled Mine units as evenly as possible. Remainder units are
// assigned round-robin. Sub-groups within each group are by tech level.
// The returned slice is nil if depositIDs is empty or there are no Mine units.
//
// Note: this is intentionally the simplest assignment that could possibly
// work. Future sprints will rework the algorithm.
func buildMiningGroups(inventory []domain.Inventory, depositIDs []domain.DepositID) []domain.MiningGroup {
	type mineEntry struct {
		TechLevel domain.TechLevel
		Quantity  int
	}
	// Aggregate by TechLevel in case multiple inventory rows share a TL.
	mineByTL := make(map[domain.TechLevel]int)
	for _, inv := range inventory {
		if inv.Unit == domain.Mine && inv.QuantityAssembled > 0 {
			mineByTL[inv.TechLevel] += inv.QuantityAssembled
		}
	}
	if len(depositIDs) == 0 || len(mineByTL) == 0 {
		return nil
	}
	var pool []mineEntry
	for tl, qty := range mineByTL {
		pool = append(pool, mineEntry{tl, qty})
	}
	slices.SortFunc(pool, func(a, b mineEntry) int {
		return cmp.Compare(a.TechLevel, b.TechLevel)
	})

	total := 0
	for _, e := range pool {
		total += e.Quantity
	}
	n := len(depositIDs)
	base := total / n
	remainder := total % n

	var groups []domain.MiningGroup
	poolIdx := 0
	poolRemaining := pool[0].Quantity

	for i, depositID := range depositIDs {
		groupQty := base
		if i < remainder {
			groupQty = base + 1
		}
		if groupQty == 0 {
			groups = append(groups, domain.MiningGroup{
				ID:      domain.MiningGroupID(i + 1),
				Deposit: depositID,
			})
			continue
		}

		var units []domain.GroupUnit
		remaining := groupQty
		for remaining > 0 && poolIdx < len(pool) {
			take := remaining
			if take > poolRemaining {
				take = poolRemaining
			}
			units = append(units, domain.GroupUnit{
				TechLevel: pool[poolIdx].TechLevel,
				Quantity:  take,
			})
			remaining -= take
			poolRemaining -= take
			if poolRemaining == 0 {
				poolIdx++
				if poolIdx < len(pool) {
					poolRemaining = pool[poolIdx].Quantity
				}
			}
		}
		groups = append(groups, domain.MiningGroup{
			ID:      domain.MiningGroupID(i + 1),
			Deposit: depositID,
			Units:   units,
		})
	}
	return groups
}

// newUUID generates a random v4 UUID using crypto/rand.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		return "", fmt.Errorf("newUUID: %w", err)
	}
	// Set version 4 and variant bits.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
