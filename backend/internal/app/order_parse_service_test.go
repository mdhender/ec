// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app_test

import (
	"errors"
	"testing"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/domain"
)

// stubOrder is a minimal domain.Order implementation for testing.
type stubOrder struct {
	kind  domain.OrderKind
	phase domain.Phase
}

func (o *stubOrder) Kind() domain.OrderKind { return o.kind }
func (o *stubOrder) TurnPhase() domain.Phase { return o.phase }
func (o *stubOrder) Validate() error         { return nil }

// stubParser is a configurable stub implementation of app.OrderParser.
type stubParser struct {
	orders      []domain.Order
	diagnostics []app.ParseDiagnostic
	err         error
}

func (p *stubParser) Parse(_ string) ([]domain.Order, []app.ParseDiagnostic, error) {
	return p.orders, p.diagnostics, p.err
}

func TestParseOrdersService_EmptyInput(t *testing.T) {
	svc := app.NewParseOrdersService(&stubParser{})

	result, err := svc.Parse("")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result.Orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(result.Orders))
	}
	if len(result.Diagnostics) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(result.Diagnostics))
	}
}

func TestParseOrdersService_ReturnsOrdersAndDiagnostics(t *testing.T) {
	orders := []domain.Order{
		&stubOrder{kind: domain.OrderKindMove, phase: 1},
		&stubOrder{kind: domain.OrderKindPay, phase: 2},
	}
	diagnostics := []app.ParseDiagnostic{
		{Line: 3, Code: "E001", Message: "unknown order keyword"},
	}

	svc := app.NewParseOrdersService(&stubParser{
		orders:      orders,
		diagnostics: diagnostics,
	})

	result, err := svc.Parse("some order text")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result.Orders) != 2 {
		t.Errorf("expected 2 orders, got %d", len(result.Orders))
	}
	if len(result.Diagnostics) != 1 {
		t.Errorf("expected 1 diagnostic, got %d", len(result.Diagnostics))
	}

	if result.Orders[0].Kind() != domain.OrderKindMove {
		t.Errorf("expected first order kind Move, got %v", result.Orders[0].Kind())
	}
	if result.Orders[1].Kind() != domain.OrderKindPay {
		t.Errorf("expected second order kind Pay, got %v", result.Orders[1].Kind())
	}
	if result.Diagnostics[0].Line != 3 {
		t.Errorf("expected diagnostic at line 3, got %d", result.Diagnostics[0].Line)
	}
}

func TestParseOrdersService_PropagatesParserFailure(t *testing.T) {
	parserErr := errors.New("unexpected internal parser failure")

	svc := app.NewParseOrdersService(&stubParser{err: parserErr})

	_, err := svc.Parse("some order text")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parserErr) {
		t.Errorf("expected parser error %v, got %v", parserErr, err)
	}
}
