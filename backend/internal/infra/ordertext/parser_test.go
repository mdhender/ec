// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package ordertext_test

import (
	"testing"

	"github.com/mdhender/ec/internal/domain"
	"github.com/mdhender/ec/internal/infra/ordertext"
)

func TestParser_ParseMVPOrders(t *testing.T) {
	p := ordertext.NewParser()

	tests := []struct {
		name     string
		input    string
		wantKind domain.OrderKind
	}{
		{
			name:     "build change",
			input:    "build change 16 8 factory",
			wantKind: domain.OrderKindBuildChange,
		},
		{
			name:     "mining change",
			input:    "mining change 348 18 92",
			wantKind: domain.OrderKindMiningChange,
		},
		{
			name:     "transfer",
			input:    "transfer 22 29 spy 10",
			wantKind: domain.OrderKindTransfer,
		},
		{
			name:     "assemble",
			input:    "assemble 91 factory 54000",
			wantKind: domain.OrderKindAssemble,
		},
		{
			name:     "move",
			input:    "move 77 orbit 6",
			wantKind: domain.OrderKindMove,
		},
		{
			name:     "draft",
			input:    "draft 13 soldier 3600",
			wantKind: domain.OrderKindDraft,
		},
		{
			name:     "pay",
			input:    "pay 38 professional 5",
			wantKind: domain.OrderKindPay,
		},
		{
			name:     "ration",
			input:    "ration 16 50",
			wantKind: domain.OrderKindRation,
		},
		{
			name:     "name planet",
			input:    "name planet 5 New Terra",
			wantKind: domain.OrderKindName,
		},
		{
			name:     "name ship",
			input:    "name ship 39 Dragonfire",
			wantKind: domain.OrderKindName,
		},
		{
			name:     "name colony",
			input:    "name colony 7 Outpost Beta",
			wantKind: domain.OrderKindName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders, diags, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(diags) != 0 {
				t.Errorf("expected 0 diagnostics, got %d: %+v", len(diags), diags)
			}
			if len(orders) != 1 {
				t.Fatalf("expected 1 order, got %d", len(orders))
			}
			if orders[0].Kind() != tt.wantKind {
				t.Errorf("expected order kind %v, got %v", tt.wantKind, orders[0].Kind())
			}
		})
	}
}

func TestParser_IgnoresBlankAndCommentLines(t *testing.T) {
	p := ordertext.NewParser()

	input := "\n   \n# this is a comment\n"
	orders, diags, err := p.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orders))
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d: %+v", len(diags), diags)
	}
}

func TestParser_RejectsUnsupportedAndMalformedLines(t *testing.T) {
	p := ordertext.NewParser()

	tests := []struct {
		name     string
		input    string
		wantCode string
	}{
		{
			name:     "not_implemented: bombard",
			input:    "bombard 1 2",
			wantCode: "not_implemented",
		},
		{
			name:     "bad_syntax: build change too few args",
			input:    "build change 1",
			wantCode: "bad_syntax",
		},
		{
			name:     "bad_value: ration percentage over 100",
			input:    "ration 1 150",
			wantCode: "bad_value",
		},
		{
			name:     "bad_value: ration non-integer colony ID",
			input:    "ration abc 50",
			wantCode: "bad_value",
		},
		{
			name:     "unknown_command: frobnicator",
			input:    "frobnicator 1 2",
			wantCode: "unknown_command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders, diags, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(orders) != 0 {
				t.Errorf("expected 0 orders, got %d", len(orders))
			}
			if len(diags) != 1 {
				t.Fatalf("expected 1 diagnostic, got %d: %+v", len(diags), diags)
			}
			if diags[0].Code != tt.wantCode {
				t.Errorf("expected diagnostic code %q, got %q (message: %s)", tt.wantCode, diags[0].Code, diags[0].Message)
			}
		})
	}
}
