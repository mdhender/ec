// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import (
	cryptorand "crypto/rand"
	"fmt"
	mathrand "math/rand/v2"
	"strings"
	"unicode"

	"github.com/mdhender/ec/internal/domain"
)

// GameConfigService orchestrates game and auth config use cases.
type GameConfigService struct {
	Store   GameConfigStore
	Cluster ClusterStore
}

// CreateGame initializes an empty game.json and auth.json in dirPath.
// Returns an error if either file already exists or if dirPath is not a directory.
func (s *GameConfigService) CreateGame(dirPath string) error {
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
// Returns the assigned empire number and the generated magic link UUID.
func (s *GameConfigService) AddEmpire(dirPath string, empireNo int, name string, homeWorldID domain.PlanetID) (int, string, error) {
	game, err := s.Store.ReadGame(dirPath)
	if err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}
	cluster, err := s.Cluster.ReadCluster(dirPath)
	if err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}

	// Resolve homeworld
	if homeWorldID == 0 {
		homeWorldID = game.ActiveHomeWorldID
	}
	if homeWorldID == 0 {
		return 0, "", fmt.Errorf("addEmpire: no active homeworld; use --homeworld or run create homeworld first")
	}

	// Find the race for this homeworld
	raceIdx := -1
	for i, r := range game.Races {
		if r.HomeWorld == homeWorldID {
			raceIdx = i
			break
		}
	}
	if raceIdx == -1 {
		return 0, "", fmt.Errorf("addEmpire: homeworld %d does not exist", homeWorldID)
	}
	race := game.Races[raceIdx]

	// Check empire limit per race
	if len(race.Empires) >= 25 {
		return 0, "", fmt.Errorf("addEmpire: homeworld %d is full (25 empires)", homeWorldID)
	}

	// Scrub empire name
	scrubbedName := scrubEmpireName(name)
	if scrubbedName == "" {
		return 0, "", fmt.Errorf("addEmpire: empire name is required")
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
			return 0, "", fmt.Errorf("addEmpire: empire %d already exists", empireNo)
		}
	}

	// Find homeworld planet in cluster
	var homePlanet *domain.Planet
	for i := range cluster.Planets {
		if cluster.Planets[i].ID == homeWorldID {
			homePlanet = &cluster.Planets[i]
			break
		}
	}
	if homePlanet == nil {
		return 0, "", fmt.Errorf("addEmpire: homeworld planet %d not found in cluster", homeWorldID)
	}

	// Find the system location for the homeworld planet
	systemLocation, err := findSystemForPlanet(cluster, homeWorldID)
	if err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}

	// Create starting colony
	colony := domain.Colony{
		ID:        domain.ColonyID(len(cluster.Colonies) + 1),
		Empire:    domain.EmpireID(empireNo),
		Location:  systemLocation,
		TechLevel: 1,
	}
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
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}

	// Write cluster.json
	if err := s.Cluster.WriteCluster(dirPath, cluster, true); err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}

	// Create empire directory
	if err := s.Store.CreateEmpireDir(dirPath, empireNo); err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}

	// Generate magic link
	authCfg, err := s.Store.ReadAuthConfig(dirPath)
	if err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}
	uuid, err := newUUID()
	if err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}
	if authCfg.MagicLinks == nil {
		authCfg.MagicLinks = map[string]domain.AuthLink{}
	}
	authCfg.MagicLinks[uuid] = domain.AuthLink{Empire: empireNo}
	if err := s.Store.WriteAuthConfig(dirPath, authCfg); err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}

	return empireNo, uuid, nil
}

// RemoveEmpire sets the empire's Active flag to false and removes its magic links.
func (s *GameConfigService) RemoveEmpire(dirPath string, empireNo int) error {
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
func (s *GameConfigService) ShowMagicLink(dirPath string, empireNo int) (string, error) {
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
func (s *GameConfigService) CreateHomeWorld(dataPath string, planetID domain.PlanetID, minDistance int) (domain.PlanetID, error) {
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

	// Set habitability on the chosen planet
	for i := range cluster.Planets {
		if cluster.Planets[i].ID == planetID {
			cluster.Planets[i].Habitability = 25
			break
		}
	}

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
