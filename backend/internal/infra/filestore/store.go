// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package filestore implements file-based order and report storage.
package filestore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/cerr"
)

// Store implements app.OrderStore and app.ReportStore using the local filesystem.
type Store struct {
	dataPath string
}

// NewStore returns a new Store rooted at dataPath.
func NewStore(dataPath string) *Store {
	return &Store{dataPath: dataPath}
}

// GetOrders reads the orders file for the given empire.
// Returns cerr.ErrNotFound (wrapped) if the file does not exist.
func (s *Store) GetOrders(_ context.Context, empireNo int) (string, error) {
	path := filepath.Join(s.dataPath, fmt.Sprintf("%d", empireNo), "orders.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("GetOrders %d: %w", empireNo, cerr.ErrNotFound)
		}
		return "", err
	}
	return string(data), nil
}

// PutOrders writes the orders file for the given empire atomically.
// The empire directory is created if it does not exist.
func (s *Store) PutOrders(_ context.Context, empireNo int, body string) error {
	dir := filepath.Join(s.dataPath, fmt.Sprintf("%d", empireNo))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "orders-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.WriteString(body); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	dest := filepath.Join(dir, "orders.txt")
	if err := os.Rename(tmpName, dest); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

// ListReports returns metadata for all turn reports available for the given empire,
// sorted by TurnYear ascending, then TurnQuarter ascending.
// Returns an empty slice (not an error) if the directory doesn't exist or no files match.
func (s *Store) ListReports(_ context.Context, empireNo int) ([]app.ReportMeta, error) {
	pattern := filepath.Join(s.dataPath, fmt.Sprintf("%d", empireNo), "*.*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// TODO: When the reports engine is added, it will create a reports.json
	// manifest. At that point, switch from glob-based discovery to reading
	// the manifest. For now, validate filenames defensively.
	var results []app.ReportMeta
	for _, match := range matches {
		base := filepath.Base(match)
		// expect "{year}.{quarter}.json" where year is 0–9999 and quarter is 0–4.
		name := strings.TrimSuffix(base, ".json")
		parts := strings.SplitN(name, ".", 2)
		if len(parts) != 2 {
			continue
		}
		turnYear, err := strconv.Atoi(parts[0])
		if err != nil || turnYear < 0 || turnYear > 9999 {
			continue
		}
		turnQuarter, err := strconv.Atoi(parts[1])
		if err != nil || turnQuarter < 0 || turnQuarter > 4 {
			continue
		}
		// year 0 is the pre-game state; only quarter 0 is valid.
		if turnYear == 0 && turnQuarter != 0 {
			continue
		}
		results = append(results, app.ReportMeta{TurnYear: turnYear, TurnQuarter: turnQuarter})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].TurnYear != results[j].TurnYear {
			return results[i].TurnYear < results[j].TurnYear
		}
		return results[i].TurnQuarter < results[j].TurnQuarter
	})

	return results, nil
}

// GetReport reads a turn report for the given empire and turn.
// Returns cerr.ErrNotFound (wrapped) if the file does not exist.
func (s *Store) GetReport(_ context.Context, empireNo, turnYear, turnQuarter int) ([]byte, error) {
	filename := fmt.Sprintf("%d.%d.json", turnYear, turnQuarter)
	path := filepath.Join(s.dataPath, fmt.Sprintf("%d", empireNo), filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("GetReport %d %d.%d: %w", empireNo, turnYear, turnQuarter, cerr.ErrNotFound)
		}
		return nil, err
	}
	return data, nil
}
