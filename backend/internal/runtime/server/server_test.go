// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package server_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mdhender/ec/internal/runtime/server"
)

// writeTestAuthFile creates a minimal auth.json for test startup.
func writeTestAuthFile(t *testing.T, dir string) {
	t.Helper()
	data, _ := json.Marshal(map[string]any{
		"magic-links": map[string]any{},
	})
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), data, 0600); err != nil {
		t.Fatalf("writeTestAuthFile: %v", err)
	}
}

func testOpts(t *testing.T) []server.Option {
	t.Helper()
	dir := t.TempDir()
	writeTestAuthFile(t, dir)
	return []server.Option{
		server.WithDataPath(dir),
		server.WithJWTSecret("test-secret"),
	}
}

func TestNewMissingDeps(t *testing.T) {
	if _, err := server.New(); err == nil {
		t.Fatal("expected error when no config provided")
	}
	if _, err := server.New(server.WithDataPath("/tmp")); err == nil {
		t.Fatal("expected error when jwtSecret missing")
	}
	if _, err := server.New(server.WithJWTSecret("s")); err == nil {
		t.Fatal("expected error when dataPath missing")
	}
	opts := testOpts(t)
	if _, err := server.New(opts...); err != nil {
		t.Fatalf("unexpected error with all config: %v", err)
	}
}

func TestAutoShutdown(t *testing.T) {
	opts := append(testOpts(t),
		server.WithHost("localhost"),
		server.WithPort("0"),
		server.WithShutdownAfter(200*time.Millisecond),
	)
	s, err := server.New(opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	errCh := make(chan error, 1)
	go func() { errCh <- s.Start() }()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down within 3 seconds")
	}
}
