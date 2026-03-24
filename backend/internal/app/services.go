// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import (
	"context"

	"github.com/mdhender/ec/internal/cerr"
)

// LoginService handles the magic-link-to-JWT authentication flow.
type LoginService struct {
	Auth  AuthStore
	Token TokenSigner
}

// Login validates a magic link and issues a JWT token.
// Returns cerr.ErrInvalidMagicLink if the link is unknown.
func (s *LoginService) Login(ctx context.Context, magicLink string) (token string, err error) {
	empireNo, ok, err := s.Auth.ValidateMagicLink(ctx, magicLink)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", cerr.ErrInvalidMagicLink
	}
	return s.Token.Issue(empireNo)
}
