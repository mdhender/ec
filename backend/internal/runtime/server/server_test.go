// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/runtime/server"
)

type mockAuthStore struct{}

func (m *mockAuthStore) ValidateMagicLink(_ context.Context, _ string) (int, bool, error) {
	return 0, false, nil
}

type mockTokenSigner struct{}

func (m *mockTokenSigner) Issue(_ int) (string, error)    { return "", nil }
func (m *mockTokenSigner) Validate(_ string) (int, error) { return 0, nil }

type mockOrderStore struct{}

func (m *mockOrderStore) GetOrders(_ context.Context, _ int) (string, error) { return "", nil }
func (m *mockOrderStore) PutOrders(_ context.Context, _ int, _ string) error { return nil }

type mockReportStore struct{}

func (m *mockReportStore) ListReports(_ context.Context, _ int) ([]app.ReportMeta, error) {
	return nil, nil
}
func (m *mockReportStore) GetReport(_ context.Context, _, _, _ int) ([]byte, error) {
	return nil, nil
}

func allDeps() []server.Option {
	return []server.Option{
		server.WithAuthStore(&mockAuthStore{}),
		server.WithTokenSigner(&mockTokenSigner{}),
		server.WithOrderStore(&mockOrderStore{}),
		server.WithReportStore(&mockReportStore{}),
	}
}

func TestNewMissingDeps(t *testing.T) {
	if _, err := server.New(); err == nil {
		t.Fatal("expected error when no stores provided")
	}
	if _, err := server.New(server.WithAuthStore(&mockAuthStore{})); err == nil {
		t.Fatal("expected error when tokenSigner missing")
	}
	if _, err := server.New(
		server.WithAuthStore(&mockAuthStore{}),
		server.WithTokenSigner(&mockTokenSigner{}),
	); err == nil {
		t.Fatal("expected error when orderStore missing")
	}
	if _, err := server.New(allDeps()...); err != nil {
		t.Fatalf("unexpected error with all deps: %v", err)
	}
}

func TestAutoShutdown(t *testing.T) {
	opts := append(allDeps(),
		server.WithHost("localhost"),
		server.WithPort("18081"),
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
