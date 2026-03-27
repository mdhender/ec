// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import "github.com/mdhender/ec/internal/domain"

// OrderParser parses raw order text into typed domain orders and diagnostics.
// Implementations live in infra; this port is the app-layer contract.
type OrderParser interface {
	Parse(text string) (orders []domain.Order, diagnostics []ParseDiagnostic, err error)
}

// ParseDiagnostic describes a parse error or warning for a single input line.
type ParseDiagnostic struct {
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ParseResult holds the outcome of a parse pass.
type ParseResult struct {
	Orders      []domain.Order   `json:"-"`
	Diagnostics []ParseDiagnostic `json:"diagnostics"`
}
