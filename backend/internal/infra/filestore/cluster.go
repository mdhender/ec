// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package filestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mdhender/ec/internal/domain"
)

// ReadCluster reads a JSON file containing a domain.Cluster.
func (s *Store) ReadCluster(path string) (domain.Cluster, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Cluster{}, fmt.Errorf("reading cluster file: %w", err)
	}
	var cluster domain.Cluster
	if err := json.Unmarshal(data, &cluster); err != nil {
		return domain.Cluster{}, fmt.Errorf("parsing cluster file: %w", err)
	}
	return cluster, nil
}

// WriteCluster writes a domain.Cluster as a JSON file.
// If overwrite is false and the file already exists, an error is returned.
func (s *Store) WriteCluster(path string, cluster domain.Cluster, overwrite bool) error {
	dir := filepath.Dir(path)
	if sb, err := os.Stat(dir); err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	} else if !sb.IsDir() {
		return fmt.Errorf("invalid directory: %s", dir)
	}
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("save file exists: %q (use --overwrite to replace)", path)
		}
	}
	data, err := json.MarshalIndent(cluster, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cluster: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing cluster file: %w", err)
	}
	return nil
}

// WriteGame writes a Game as a JSON file.
// If overwrite is false and the file already exists, an error is returned.
func (s *Store) WriteGame(path string, game *domain.Game, overwrite bool) error {
	dir := filepath.Dir(path)
	if sb, err := os.Stat(dir); err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	} else if !sb.IsDir() {
		return fmt.Errorf("invalid directory: %s", dir)
	}
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("save file exists: %q (use --overwrite to replace)", path)
		}
	}
	data, err := json.MarshalIndent(game, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling game: %w", err)
	}
	if err := os.WriteFile(filepath.Join(path, "game.json"), data, 0o644); err != nil {
		return fmt.Errorf("writing game file: %w", err)
	}
	return nil
}
