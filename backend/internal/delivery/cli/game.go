// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"context"
	"fmt"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/domain"
	"github.com/peterbourgon/ff/v4"
)

// CmdCreateGame returns an ff.Command that initializes game.json and auth.json in a directory.
func CmdCreateGame(svc *app.GameService) *ff.Command {
	fs := ff.NewFlagSet("game")
	dataPath := fs.StringLong("data-path", "", "directory to write game.json and auth.json")

	return &ff.Command{
		Name:      "game",
		Usage:     "cli create game [FLAGS]",
		ShortHelp: "initialize game.json and auth.json in a directory",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			if *dataPath == "" {
				return fmt.Errorf("--data-path is required (or set EC_DATA_PATH)")
			}
			if err := svc.CreateGame(*dataPath); err != nil {
				return err
			}
			fmt.Printf("game created: %s\n", *dataPath)
			return nil
		},
	}
}

// CmdCreateHomeWorld returns an ff.Command that selects or validates a homeworld planet.
func CmdCreateHomeWorld(svc *app.GameService) *ff.Command {
	fs := ff.NewFlagSet("homeworld")
	dataPath := fs.StringLong("data-path", "", "directory containing game.json and cluster.json")
	planet := fs.IntLong("planet", 0, "planet ID to use as homeworld (0 = auto-select)")
	minDistance := fs.IntLong("min-distance", 3, "minimum distance from existing homeworlds")

	return &ff.Command{
		Name:      "homeworld",
		Usage:     "cli create homeworld [FLAGS]",
		ShortHelp: "select or validate a homeworld planet and record a new race",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			if *dataPath == "" {
				return fmt.Errorf("--data-path is required")
			}
			planetID, err := svc.CreateHomeWorld(*dataPath, domain.PlanetID(*planet), *minDistance)
			if err != nil {
				return err
			}
			fmt.Printf("homeworld created: planet %d\n", planetID)
			return nil
		},
	}
}

// CmdAddEmpire returns an ff.Command that adds an empire to game.json and auth.json.
func CmdAddEmpire(svc *app.GameService) *ff.Command {
	fs := ff.NewFlagSet("empire")
	dataPath := fs.StringLong("data-path", "", "directory containing game.json and auth.json")
	empireNo := fs.IntLong("empire", 0, "empire number (0 = auto-assign)")
	name := fs.StringLong("name", "", "empire name")
	homeworld := fs.IntLong("homeworld", 0, "homeworld planet ID (0 = use active homeworld)")

	return &ff.Command{
		Name:      "empire",
		Usage:     "cli create empire [FLAGS]",
		ShortHelp: "add an empire to game.json and auth.json",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			if *dataPath == "" {
				return fmt.Errorf("--data-path is required (or set EC_DATA_PATH)")
			}
			n, scrubbedName, uuid, err := svc.AddEmpire(*dataPath, *empireNo, *name, domain.PlanetID(*homeworld))
			if err != nil {
				return err
			}
			fmt.Printf("added empire %d (%s), magic link: %s\n", n, scrubbedName, uuid)
			return nil
		},
	}
}

// CmdRemoveEmpire returns an ff.Command that deactivates an empire.
func CmdRemoveEmpire(svc *app.GameService) *ff.Command {
	fs := ff.NewFlagSet("empire")
	dataPath := fs.StringLong("data-path", "", "directory containing game.json and auth.json")
	empireNo := fs.IntLong("empire", 0, "empire number")

	return &ff.Command{
		Name:      "empire",
		Usage:     "cli remove empire [FLAGS]",
		ShortHelp: "deactivate an empire in game.json",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			if *dataPath == "" {
				return fmt.Errorf("--data-path is required (or set EC_DATA_PATH)")
			}
			if *empireNo == 0 {
				return fmt.Errorf("--empire is required")
			}
			if err := svc.RemoveEmpire(*dataPath, *empireNo); err != nil {
				return err
			}
			fmt.Printf("removed empire %d\n", *empireNo)
			return nil
		},
	}
}

// CmdShowMagicLink returns an ff.Command that prints the magic link URL for an empire.
func CmdShowMagicLink(svc *app.GameService) *ff.Command {
	fs := ff.NewFlagSet("magic-link")
	dataPath := fs.StringLong("data-path", "", "directory containing auth.json")
	baseURL := fs.StringLong("base-url", "", "application base URL")
	empireNo := fs.IntLong("empire", 0, "empire number")

	return &ff.Command{
		Name:      "magic-link",
		Usage:     "cli show magic-link [FLAGS]",
		ShortHelp: "show the magic link URL for an empire",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			if *dataPath == "" {
				return fmt.Errorf("--data-path is required (or set EC_DATA_PATH)")
			}
			if *empireNo == 0 {
				return fmt.Errorf("--empire is required")
			}
			uuid, err := svc.ShowMagicLink(*dataPath, *empireNo)
			if err != nil {
				return err
			}
			fmt.Printf("%s?magic=%s\n", *baseURL, uuid)
			return nil
		},
	}
}
