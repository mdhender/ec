// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import (
	"crypto/rand"
	"fmt"

	"github.com/mdhender/ec/internal/domain"
)

// GameConfigService orchestrates game and auth config use cases.
type GameConfigService struct {
	Store GameConfigStore
}

// CreateGame initializes an empty game.json and auth.json in dirPath.
// Returns an error if either file already exists or if dirPath is not a directory.
func (s *GameConfigService) CreateGame(dirPath string) error {
	if err := s.Store.ValidateDir(dirPath); err != nil {
		return fmt.Errorf("createGame: %w", err)
	}
	if exists, err := s.Store.GameConfigExists(dirPath); err != nil {
		return fmt.Errorf("createGame: %w", err)
	} else if exists {
		return fmt.Errorf("createGame: game.json already exists in %q", dirPath)
	}
	if exists, err := s.Store.AuthConfigExists(dirPath); err != nil {
		return fmt.Errorf("createGame: %w", err)
	} else if exists {
		return fmt.Errorf("createGame: auth.json already exists in %q", dirPath)
	}

	if err := s.Store.WriteGameConfig(dirPath, domain.GameConfig{Empires: []domain.EmpireEntry{}}); err != nil {
		return fmt.Errorf("createGame: %w", err)
	}
	if err := s.Store.WriteAuthConfig(dirPath, domain.AuthConfig{MagicLinks: map[string]domain.AuthLink{}}); err != nil {
		return fmt.Errorf("createGame: %w", err)
	}
	return nil
}

// AddEmpire adds a new empire to game.json and a magic link to auth.json.
// If empireNo is 0, the next empire number is auto-assigned.
// Returns the assigned empire number and the generated magic link UUID.
func (s *GameConfigService) AddEmpire(dirPath string, empireNo int) (int, string, error) {
	cfg, err := s.Store.ReadGameConfig(dirPath)
	if err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}

	if empireNo == 0 {
		max := 0
		for _, e := range cfg.Empires {
			if e.Empire > max {
				max = e.Empire
			}
		}
		empireNo = max + 1
	}

	for _, e := range cfg.Empires {
		if e.Empire == empireNo {
			return 0, "", fmt.Errorf("addEmpire: empire %d already exists", empireNo)
		}
	}

	cfg.Empires = append(cfg.Empires, domain.EmpireEntry{Empire: empireNo, Active: true})
	if err := s.Store.WriteGameConfig(dirPath, cfg); err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}

	// Create the empire's data directory.
	if err := s.Store.CreateEmpireDir(dirPath, empireNo); err != nil {
		return 0, "", fmt.Errorf("addEmpire: %w", err)
	}

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
	cfg, err := s.Store.ReadGameConfig(dirPath)
	if err != nil {
		return fmt.Errorf("removeEmpire: %w", err)
	}

	found := false
	for i, e := range cfg.Empires {
		if e.Empire == empireNo {
			cfg.Empires[i].Active = false
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("removeEmpire: empire %d not found", empireNo)
	}
	if err := s.Store.WriteGameConfig(dirPath, cfg); err != nil {
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

// newUUID generates a random v4 UUID using crypto/rand.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("newUUID: %w", err)
	}
	// Set version 4 and variant bits.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
