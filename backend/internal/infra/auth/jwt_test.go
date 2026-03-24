// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package auth_test

import (
	"testing"
	"time"

	"github.com/mdhender/ec/internal/infra/auth"
)

func TestJWTRoundTrip(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", time.Hour)
	const empireNo = 42

	token, err := mgr.Issue(empireNo)
	if err != nil {
		t.Fatalf("Issue: unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("Issue: expected non-empty token")
	}

	got, err := mgr.Validate(token)
	if err != nil {
		t.Fatalf("Validate: unexpected error: %v", err)
	}
	if got != empireNo {
		t.Fatalf("Validate: got empireNo %d, want %d", got, empireNo)
	}
}

func TestJWTExpired(t *testing.T) {
	mgr := auth.NewJWTManager("test-secret", time.Millisecond)

	token, err := mgr.Issue(42)
	if err != nil {
		t.Fatalf("Issue: unexpected error: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	_, err = mgr.Validate(token)
	if err == nil {
		t.Fatal("Validate: expected error for expired token, got nil")
	}
}
