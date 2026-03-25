// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package domain_test

import (
	"math"
	"testing"

	"github.com/mdhender/ec/internal/domain"
)

func TestCoordsDistance(t *testing.T) {
	tests := []struct {
		name string
		a, b domain.Coords
		want float64
	}{
		{"zero distance", domain.Coords{0, 0, 0}, domain.Coords{0, 0, 0}, 0},
		{"x-axis", domain.Coords{0, 0, 0}, domain.Coords{3, 0, 0}, 3},
		{"y-axis", domain.Coords{0, 0, 0}, domain.Coords{0, 4, 0}, 4},
		{"z-axis", domain.Coords{0, 0, 0}, domain.Coords{0, 0, 5}, 5},
		{"diagonal 3-4-5", domain.Coords{0, 0, 0}, domain.Coords{3, 4, 0}, 5},
		{"3d diagonal", domain.Coords{1, 2, 3}, domain.Coords{4, 6, 3}, 5},
		{"symmetric", domain.Coords{10, 10, 10}, domain.Coords{0, 0, 0}, math.Sqrt(300)},
		{"negative coords", domain.Coords{-1, -2, -3}, domain.Coords{1, 2, 3}, math.Sqrt(56)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Distance(tt.b)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Distance(%v, %v) = %f, want %f", tt.a, tt.b, got, tt.want)
			}
			// Verify symmetry
			rev := tt.b.Distance(tt.a)
			if math.Abs(rev-got) > 1e-9 {
				t.Errorf("Distance is not symmetric: %f vs %f", got, rev)
			}
		})
	}
}
