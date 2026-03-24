// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cerr_test

import (
	"testing"

	"github.com/mdhender/ec/internal/cerr"
)

func TestErrorsImplementError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrUnauthorized", cerr.ErrUnauthorized, "unauthorized"},
		{"ErrForbidden", cerr.ErrForbidden, "forbidden"},
		{"ErrNotFound", cerr.ErrNotFound, "not found"},
		{"ErrInvalidMagicLink", cerr.ErrInvalidMagicLink, "invalid magic link"},
		{"ErrInvalidToken", cerr.ErrInvalidToken, "invalid token"},
		{"ErrMissingToken", cerr.ErrMissingToken, "missing token"},
		{"ErrInvalidEmpire", cerr.ErrInvalidEmpire, "invalid empire number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("expected non-nil error")
			}
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}
