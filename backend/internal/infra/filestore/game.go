// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package filestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mdhender/ec/internal/domain"
)

// ValidateDir checks that dirPath exists and is a directory.
func (s *Store) ValidateDir(dirPath string) error {
	fi, err := os.Stat(dirPath)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%q is not a directory", dirPath)
	}
	return nil
}

// GameExists reports whether game.json exists in dirPath.
func (s *Store) GameExists(dirPath string) (bool, error) {
	_, err := os.Stat(filepath.Join(dirPath, "game.json"))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// AuthConfigExists reports whether auth.json exists in dirPath.
func (s *Store) AuthConfigExists(dirPath string) (bool, error) {
	_, err := os.Stat(filepath.Join(dirPath, "auth.json"))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ReadGame reads game.json from dirPath and returns the parsed Game.
func (s *Store) ReadGame(dirPath string) (domain.Game, error) {
	path := filepath.Join(dirPath, "game.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Game{}, fmt.Errorf("reading game.json: %w", err)
	}
	var game domain.Game
	if err := json.Unmarshal(data, &game); err != nil {
		return domain.Game{}, fmt.Errorf("parsing game.json: %w", err)
	}
	return game, nil
}

// WriteGame marshals game as indented JSON and writes it to game.json in dirPath.
func (s *Store) WriteGame(dirPath string, game domain.Game) error {
	data, err := json.MarshalIndent(game, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling game: %w", err)
	}
	data = append(data, '\n')
	path := filepath.Join(dirPath, "game.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing game.json: %w", err)
	}
	return nil
}

// ReadAuthConfig reads auth.json from dirPath and returns the parsed AuthConfig.
func (s *Store) ReadAuthConfig(dirPath string) (domain.AuthConfig, error) {
	path := filepath.Join(dirPath, "auth.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.AuthConfig{}, fmt.Errorf("reading auth.json: %w", err)
	}
	var cfg domain.AuthConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return domain.AuthConfig{}, fmt.Errorf("parsing auth.json: %w", err)
	}
	return cfg, nil
}

// WriteAuthConfig marshals cfg as indented JSON and writes it to auth.json in dirPath.
func (s *Store) WriteAuthConfig(dirPath string, cfg domain.AuthConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling auth config: %w", err)
	}
	data = append(data, '\n')
	path := filepath.Join(dirPath, "auth.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing auth.json: %w", err)
	}
	return nil
}

// CreateEmpireDir creates the data directory for an empire inside dirPath.
func (s *Store) CreateEmpireDir(dirPath string, empireNo int) error {
	empireDir := filepath.Join(dirPath, strconv.Itoa(empireNo))
	if err := os.MkdirAll(empireDir, 0o755); err != nil {
		return fmt.Errorf("creating empire directory: %w", err)
	}
	return nil
}
