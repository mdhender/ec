// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package cerr implements constant errors.
package cerr

// Error implements errors.Error
type Error string

func (err Error) Error() string {
	return string(err)
}

const (
	ErrUnauthorized     = Error("unauthorized")
	ErrForbidden        = Error("forbidden")
	ErrNotFound         = Error("not found")
	ErrInvalidMagicLink = Error("invalid magic link")
	ErrInvalidToken     = Error("invalid token")
	ErrMissingToken     = Error("missing token")
	ErrInvalidEmpire    = Error("invalid empire number")
)
