// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package auth_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/ec/internal/infra/auth"
)

const testAuthJSON = `{
  "magic-links": {
    "81ce2bb6-42fe-49b2-80c5-0558787c8471": {"empire": 1812},
    "37e81785-84ee-4fee-850b-160e373a4539": {"empire": 42}
  }
}`

// writeTempAuthFile creates a temporary auth JSON file and returns its path.
func writeTempAuthFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeTempAuthFile: %v", err)
	}
	return path
}

func TestMagicLinkValid(t *testing.T) {
	path := writeTempAuthFile(t, testAuthJSON)

	store, err := auth.NewMagicLinkStore(path)
	if err != nil {
		t.Fatalf("NewMagicLinkStore: unexpected error: %v", err)
	}

	ctx := context.Background()

	empireNo, ok, err := store.ValidateMagicLink(ctx, "81ce2bb6-42fe-49b2-80c5-0558787c8471")
	if err != nil {
		t.Fatalf("ValidateMagicLink: unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("ValidateMagicLink: expected ok==true for known link")
	}
	if empireNo != 1812 {
		t.Fatalf("ValidateMagicLink: expected empireNo 1812, got %d", empireNo)
	}

	empireNo, ok, err = store.ValidateMagicLink(ctx, "37e81785-84ee-4fee-850b-160e373a4539")
	if err != nil {
		t.Fatalf("ValidateMagicLink: unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("ValidateMagicLink: expected ok==true for known link")
	}
	if empireNo != 42 {
		t.Fatalf("ValidateMagicLink: expected empireNo 42, got %d", empireNo)
	}
}

func TestMagicLinkInvalid(t *testing.T) {
	path := writeTempAuthFile(t, testAuthJSON)

	store, err := auth.NewMagicLinkStore(path)
	if err != nil {
		t.Fatalf("NewMagicLinkStore: unexpected error: %v", err)
	}

	ctx := context.Background()
	empireNo, ok, err := store.ValidateMagicLink(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("ValidateMagicLink: expected err==nil for unknown link, got: %v", err)
	}
	if ok {
		t.Fatal("ValidateMagicLink: expected ok==false for unknown link")
	}
	if empireNo != 0 {
		t.Fatalf("ValidateMagicLink: expected empireNo 0 for unknown link, got %d", empireNo)
	}
}

func TestMagicLinkBadFile(t *testing.T) {
	_, err := auth.NewMagicLinkStore("/nonexistent/path/auth.json")
	if err == nil {
		t.Fatal("NewMagicLinkStore: expected error for nonexistent file, got nil")
	}
}
