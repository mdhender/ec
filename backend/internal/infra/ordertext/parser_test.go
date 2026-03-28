// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package ordertext_test

import (
	"os"
	"path/filepath"
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
			input:    "build change 16 8 factory-6",
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
			name:     "assemble other",
			input:    "assemble 91 factory-6 54000",
			wantKind: domain.OrderKindAssemble,
		},
		{
			name:     "assemble other with commas",
			input:    "assemble 58 missile-launcher-1 6,000",
			wantKind: domain.OrderKindAssemble,
		},
		{
			name:     "assemble factory",
			input:    "assemble 91 factory factory-6 54000 hyper-engine-1",
			wantKind: domain.OrderKindAssemble,
		},
		{
			name:     "assemble factory with commas",
			input:    "assemble 91 factory factory-6 54,000 hyper-engine-1",
			wantKind: domain.OrderKindAssemble,
		},
		{
			name:     "assemble mine",
			input:    "assemble 83 mine mine-2 25680 92",
			wantKind: domain.OrderKindAssemble,
		},
		{
			name:     "assemble mine with commas",
			input:    "assemble 83 mine mine-2 25,680 92",
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
			input:    `name planet 5 "New Terra"`,
			wantKind: domain.OrderKindName,
		},
		{
			name:     "name ship",
			input:    `name ship 39 "Dragonfire"`,
			wantKind: domain.OrderKindName,
		},
		{
			name:     "name colony",
			input:    `name colony 7 "Outpost Beta"`,
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

func TestParser_ScrubbedNames(t *testing.T) {
	p := ordertext.NewParser()

	tests := []struct {
		name     string
		input    string
		wantName string
	}{
		{
			name:     "trim leading and trailing whitespace, collapse internal spaces",
			input:    `name ship 39 " fool  of a took "`,
			wantName: "fool of a took",
		},
		{
			name:     "strip ampersand",
			input:    `name ship 39 "Me & Joe"`,
			wantName: "Me Joe",
		},
		{
			name:     "preserve hash",
			input:    `name ship 39 "Borg #9"`,
			wantName: "Borg #9",
		},
		{
			name:     "strip backslash",
			input:    `name ship 39 "Joy\Division"`,
			wantName: "JoyDivision",
		},
		// allowed punctuation
		{
			name:     "allow forward slash",
			input:    `name ship 39 "Slash // Burn"`,
			wantName: "Slash // Burn",
		},
		{
			name:     "allow comma",
			input:    `name colony 7 "Ready, Aim"`,
			wantName: "Ready, Aim",
		},
		{
			name:     "allow period",
			input:    `name ship 39 "U.S.S. Daring"`,
			wantName: "U.S.S. Daring",
		},
		{
			name:     "allow hyphen",
			input:    `name colony 7 "Alpha-Prime"`,
			wantName: "Alpha-Prime",
		},
		{
			name:     "allow underscore",
			input:    `name ship 39 "Dark_Star"`,
			wantName: "Dark_Star",
		},
		{
			name:     "allow plus",
			input:    `name colony 7 "C++"`,
			wantName: "C++",
		},
		{
			name:     "allow parentheses",
			input:    `name ship 39 "Argo (II)"`,
			wantName: "Argo (II)",
		},
		{
			name:     "allow apostrophe",
			input:    `name colony 7 "O'Brien's World"`,
			wantName: "O'Brien's World",
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
			o, ok := orders[0].(domain.NameOrder)
			if !ok {
				t.Fatalf("expected domain.NameOrder, got %T", orders[0])
			}
			if o.NewName != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, o.NewName)
			}
		})
	}
}

func TestParser_IgnoresBlankAndCommentLines(t *testing.T) {
	p := ordertext.NewParser()

	input := "\n   \n// this is a comment\n"
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

func TestParser_MidLineComment(t *testing.T) {
	p := ordertext.NewParser()

	// "move 77 orbit 6 // move the scout inward" should parse as a valid move order.
	orders, diags, err := p.Parse("move 77 orbit 6 // move the scout inward")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d: %+v", len(diags), diags)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	if orders[0].Kind() != domain.OrderKindMove {
		t.Errorf("expected OrderKindMove, got %v", orders[0].Kind())
	}
}

func TestParser_SlashSlashInsideQuotedName(t *testing.T) {
	p := ordertext.NewParser()

	// `name ship 39 "Slash // Burn"` — the // inside the quoted name must not be
	// treated as a comment; the order must still parse successfully.
	orders, diags, err := p.Parse(`name ship 39 "Slash // Burn"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d: %+v", len(diags), diags)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
}

func TestParser_SetupNotImplemented(t *testing.T) {
	p := ordertext.NewParser()

	orders, diags, err := p.Parse("setup ship from 29")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orders))
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d: %+v", len(diags), diags)
	}
	if diags[0].Code != "not_implemented" {
		t.Errorf("expected diagnostic code %q, got %q", "not_implemented", diags[0].Code)
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
			name:     "syntax: build change too few args",
			input:    "build change 1",
			wantCode: "syntax",
		},
		{
			name:     "invalid_value: ration percentage over 100",
			input:    "ration 1 150",
			wantCode: "invalid_value",
		},
		{
			name:     "invalid_value: ration non-integer colony ID",
			input:    "ration abc 50",
			wantCode: "invalid_value",
		},
		{
			name:     "unterminated_quote: unclosed name",
			input:    `name ship 39 "Unfinished`,
			wantCode: "unterminated_quote",
		},
		{
			name:     "unexpected_end: end without setup",
			input:    "end",
			wantCode: "unexpected_end",
		},
		{
			name:     "syntax: unquoted single-word name",
			input:    "name ship 39 Dragonfire",
			wantCode: "syntax",
		},
		{
			name:     "syntax: unquoted multi-word name",
			input:    `name colony 7 Outpost Beta`,
			wantCode: "syntax",
		},
		{
			name:     "invalid_value: bare factory without tech level",
			input:    "build change 16 8 factory",
			wantCode: "invalid_value",
		},
		{
			name:     "invalid_value: bare hyper-engine without tech level",
			input:    "assemble 91 hyper-engine 10",
			wantCode: "invalid_value",
		},
		{
			name:     "syntax: assemble factory too few args",
			input:    "assemble 91 factory factory-6 54000",
			wantCode: "syntax",
		},
		{
			name:     "syntax: assemble mine too few args",
			input:    "assemble 83 mine mine-2 25680",
			wantCode: "syntax",
		},
		{
			name:     "invalid_value: assemble factory with non-factory unit",
			input:    "assemble 91 factory mine-2 100 hyper-engine-1",
			wantCode: "invalid_value",
		},
		{
			name:     "invalid_value: assemble mine with non-mine unit",
			input:    "assemble 83 mine factory-6 100 92",
			wantCode: "invalid_value",
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

func TestParser_ValidOrderFile(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "valid_orders.txt"))
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	p := ordertext.NewParser()
	orders, diags, err := p.Parse(string(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		for _, d := range diags {
			t.Errorf("unexpected diagnostic at line %d [%s]: %s", d.Line, d.Code, d.Message)
		}
	}
	if len(orders) == 0 {
		t.Fatalf("expected at least one order, got 0")
	}
	t.Logf("parsed %d orders, %d diagnostics", len(orders), len(diags))
	for i, o := range orders {
		t.Logf("  order[%d]: kind=%s phase=%d", i, o.Kind(), o.TurnPhase())
	}
}

func TestParser_ErrorsFile(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "errors_mixed.txt"))
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	p := ordertext.NewParser()
	orders, diags, err := p.Parse(string(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// errors_mixed.txt has 3 valid orders and 8 bad lines
	wantOrders := 3
	wantDiags := 8
	if len(orders) != wantOrders {
		t.Errorf("expected %d orders, got %d", wantOrders, len(orders))
	}
	if len(diags) != wantDiags {
		t.Errorf("expected %d diagnostics, got %d", wantDiags, len(diags))
	}

	t.Logf("parsed %d orders, %d diagnostics", len(orders), len(diags))
	for i, o := range orders {
		t.Logf("  order[%d]: kind=%s phase=%d", i, o.Kind(), o.TurnPhase())
	}
	for i, d := range diags {
		t.Logf("  diag[%d]: line=%d code=%s msg=%s", i, d.Line, d.Code, d.Message)
	}
}
