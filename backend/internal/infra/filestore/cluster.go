// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package filestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mdhender/ec/internal/domain"
)

// ReadCluster reads cluster.json from dataPath directory and returns the parsed Cluster.
func (s *Store) ReadCluster(dataPath string) (domain.Cluster, error) {
	path := filepath.Join(dataPath, "cluster.json")
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

// WriteCluster writes a domain.Cluster as cluster.json inside dataPath directory.
// If overwrite is false and the file already exists, an error is returned.
func (s *Store) WriteCluster(dataPath string, cluster domain.Cluster, overwrite bool) error {
	if sb, err := os.Stat(dataPath); err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	} else if !sb.IsDir() {
		return fmt.Errorf("invalid directory: %s", dataPath)
	}
	path := filepath.Join(dataPath, "cluster.json")
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
