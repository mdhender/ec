// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package filestore_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/ec/internal/cerr"
	"github.com/mdhender/ec/internal/infra/filestore"
)

func TestOrdersRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := filestore.NewStore(dir)
	ctx := context.Background()

	const empireNo = 42
	const body = "move fleet to sector 7\ncolonize planet 3\n"

	if err := s.PutOrders(ctx, empireNo, body); err != nil {
		t.Fatalf("PutOrders: %v", err)
	}

	got, err := s.GetOrders(ctx, empireNo)
	if err != nil {
		t.Fatalf("GetOrders: %v", err)
	}
	if got != body {
		t.Errorf("content mismatch: got %q, want %q", got, body)
	}
}

func TestOrdersNotFound(t *testing.T) {
	dir := t.TempDir()
	s := filestore.NewStore(dir)
	ctx := context.Background()

	_, err := s.GetOrders(ctx, 99999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, cerr.ErrNotFound) {
		t.Errorf("expected cerr.ErrNotFound, got %v", err)
	}
}

func TestListReports(t *testing.T) {
	dir := t.TempDir()
	s := filestore.NewStore(dir)
	ctx := context.Background()

	const empireNo = 7
	empireDir := filepath.Join(dir, fmt.Sprintf("%d", empireNo))
	if err := os.MkdirAll(empireDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Create report files out of order to verify sorting.
	for _, name := range []string{"1.2.json", "1.1.json", "2.1.json"} {
		path := filepath.Join(empireDir, name)
		if err := os.WriteFile(path, []byte(`{}`), 0644); err != nil {
			t.Fatalf("WriteFile %s: %v", name, err)
		}
	}

	reports, err := s.ListReports(ctx, empireNo)
	if err != nil {
		t.Fatalf("ListReports: %v", err)
	}

	want := [][2]int{{1, 1}, {1, 2}, {2, 1}}
	if len(reports) != len(want) {
		t.Fatalf("got %d reports, want %d", len(reports), len(want))
	}
	for i, w := range want {
		if reports[i].TurnYear != w[0] || reports[i].TurnQuarter != w[1] {
			t.Errorf("reports[%d] = {%d,%d}, want {%d,%d}",
				i, reports[i].TurnYear, reports[i].TurnQuarter, w[0], w[1])
		}
	}
}

func TestGetReportNotFound(t *testing.T) {
	dir := t.TempDir()
	s := filestore.NewStore(dir)
	ctx := context.Background()

	_, err := s.GetReport(ctx, 1, 1, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, cerr.ErrNotFound) {
		t.Errorf("expected cerr.ErrNotFound, got %v", err)
	}
}
