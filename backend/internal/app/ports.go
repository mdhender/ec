// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import "context"

// AuthStore loads and validates magic links.
type AuthStore interface {
	ValidateMagicLink(ctx context.Context, magicLink string) (empireNo int, ok bool, err error)
}

// TokenSigner issues and validates JWT tokens.
type TokenSigner interface {
	Issue(empireNo int) (token string, err error)
	Validate(token string) (empireNo int, err error)
}

// OrderStore reads and writes empire order files.
type OrderStore interface {
	GetOrders(ctx context.Context, empireNo int) (string, error)
	PutOrders(ctx context.Context, empireNo int, body string) error
}

// ReportStore lists and reads empire turn reports.
type ReportStore interface {
	ListReports(ctx context.Context, empireNo int) ([]ReportMeta, error)
	GetReport(ctx context.Context, empireNo int, turnYear, turnQuarter int) ([]byte, error)
}

// ReportMeta is metadata about a turn report (for listing).
type ReportMeta struct {
	TurnYear    int `json:"turn_year"`
	TurnQuarter int `json:"turn_quarter"`
}
