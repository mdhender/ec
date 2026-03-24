// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package auth implements authentication adapters for the infra layer.
package auth

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"os"
)

// MagicLinkStore implements app.AuthStore by loading magic links from a JSON
// file at startup and validating them in memory.
type MagicLinkStore struct {
	links map[string]int
}

// magicLinksFile is the on-disk JSON structure for the auth file.
type magicLinksFile struct {
	MagicLinks map[string]struct {
		Empire int `json:"empire"`
	} `json:"magic-links"`
}

// NewMagicLinkStore reads and parses the JSON file at path and returns a
// MagicLinkStore with the links loaded into memory. Returns an error if the
// file does not exist or contains invalid JSON.
func NewMagicLinkStore(path string) (*MagicLinkStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var parsed magicLinksFile
	if err = json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	links := make(map[string]int, len(parsed.MagicLinks))
	for key, val := range parsed.MagicLinks {
		links[key] = val.Empire
	}

	return &MagicLinkStore{links: links}, nil
}

// ValidateMagicLink checks whether magicLink matches any stored link using a
// constant-time comparison. Returns the empire number and ok==true on match,
// or 0 and ok==false when no match is found. err is always nil.
func (s *MagicLinkStore) ValidateMagicLink(_ context.Context, magicLink string) (empireNo int, ok bool, err error) {
	candidate := []byte(magicLink)
	for storedKey, empire := range s.links {
		if subtle.ConstantTimeCompare(candidate, []byte(storedKey)) == 1 {
			return empire, true, nil
		}
	}
	return 0, false, nil
}
