// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package dotfiles

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// dotfiles tries to emulate the priority list from the dotenv page at
// https://github.com/bkeepers/dotenv#what-other-env-files-can-i-use
// Pri  Filename______________  Env__  .gitignore?
// 1st  .env.development.local  dev    yes
// 1st  .env.test.local         test   yes
// 1st  .env.production.local   prod   yes
// 2nd  .env.local              all    yes
// 3rd  .env.development        dev    no, but be wary of secrets
// 3rd  .env.test               test   no, but be wary of secrets
// 3rd  .env.production         prod   no, but be wary of secrets
// 4th  .env                    all    no, but be wary of secrets
//
// Notes:
//   - The .env.*.local files are for local overrides of environment-specific settings.
//     We assume that they're created by the deployment process.
//     They can contain sensitive information like credentials or tokens.
//     They are loaded first, so they will override settings in the shared files.
//     They should never be checked into the repository.
//   - The .env.local file is loaded in development and production; never in test.
//     It should never be checked into the repository.
//   - The .env.* files are shared environment-specific settings.
//     They should not contain sensitive information like credentials or tokens.
//     They should always be checked into the repository.
//   - The .env file is loaded in all environments.
//     It should not contain sensitive information like credentials or tokens.
//     It is loaded last, so all other files will override any settings.
//     It should always be checked into the repository.

// Load the .env file.
func Load(prefix string, logger *slog.Logger) error {
	envvar := "ENV"
	if prefix != "" {
		envvar = prefix + "_ENV"
	}
	env := os.Getenv(envvar)
	logger.Debug("dotfiles", "env", env)

	// local environment files are the highest priority
	for _, local := range []string{"development", "test", "production"} {
		if local == env {
			if envfile := ".env." + local + ".local"; isFile(envfile) {
				if err := godotenv.Load(envfile); err != nil {
					return err
				}
				logger.Info("dotfiles", "loaded", envfile)
			}
		}
	}

	// .env.local is loaded for all environments except test.
	if env != "test" {
		if envfile := ".env.local"; isFile(envfile) {
			if err := godotenv.Load(envfile); err != nil {
				return err
			}
			logger.Info("dotfiles", "loaded", envfile)
		}
	}

	// shared environment specific settings
	for _, shared := range []string{"development", "test", "production"} {
		if shared == env {
			if envfile := ".env." + shared; isFile(envfile) {
				if err := godotenv.Load(envfile); err != nil {
					return err
				}
				logger.Info("dotfiles", "loaded", envfile)
			}
		}
	}

	// .env is the lowest priority
	if envfile := ".env"; isFile(envfile) {
		if err := godotenv.Load(envfile); err != nil {
			return err
		}
		logger.Info("dotfiles", "loaded", envfile)
	}

	return nil
}

func isFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir() && fi.Mode().IsRegular()
}
