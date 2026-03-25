// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package domain

// AuthConfig is the on-disk structure for auth.json.
type AuthConfig struct {
	MagicLinks map[string]AuthLink `json:"magic-links"`
}

// AuthLink maps a magic link UUID to an empire number.
type AuthLink struct {
	Empire int `json:"empire"`
}
