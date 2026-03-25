// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import "github.com/mdhender/ec/internal/domain"

// GameConfigStore reads and writes game.json and auth.json files.
// The path parameter is the directory containing the files.
type GameConfigStore interface {
	ValidateDir(path string) error
	GameExists(dirPath string) (bool, error)
	AuthConfigExists(dirPath string) (bool, error)
	ReadGame(path string) (domain.Game, error)
	WriteGame(path string, game domain.Game) error
	ReadAuthConfig(path string) (domain.AuthConfig, error)
	WriteAuthConfig(path string, cfg domain.AuthConfig) error
	CreateEmpireDir(dirPath string, empireNo int) error
}
