// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import "github.com/mdhender/ec/internal/domain"

// GameConfigStore reads and writes game.json and auth.json files.
// The path parameter is the directory containing the files.
type GameConfigStore interface {
	ValidateDir(path string) error
	GameConfigExists(dirPath string) (bool, error)
	AuthConfigExists(dirPath string) (bool, error)
	ReadGameConfig(path string) (domain.GameConfig, error)
	WriteGameConfig(path string, cfg domain.GameConfig) error
	ReadAuthConfig(path string) (domain.AuthConfig, error)
	WriteAuthConfig(path string, cfg domain.AuthConfig) error
	CreateEmpireDir(dirPath string, empireNo int) error
}
