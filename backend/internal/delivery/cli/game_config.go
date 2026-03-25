// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"context"
	"fmt"

	"github.com/mdhender/ec/internal/app"
	"github.com/peterbourgon/ff/v4"
)

// CmdCreateGame returns an ff.Command that initializes game.json and auth.json in a directory.
func CmdCreateGame(svc *app.GameConfigService) *ff.Command {
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

// CmdAddEmpire returns an ff.Command that adds an empire to game.json and auth.json.
func CmdAddEmpire(svc *app.GameConfigService) *ff.Command {
	fs := ff.NewFlagSet("empire")
	dataPath := fs.StringLong("data-path", "", "directory containing game.json and auth.json")
	empireNo := fs.IntLong("empire", 0, "empire number (0 = auto-assign)")

	return &ff.Command{
		Name:      "empire",
		Usage:     "cli create empire [FLAGS]",
		ShortHelp: "add an empire to game.json and auth.json",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			if *dataPath == "" {
				return fmt.Errorf("--data-path is required (or set EC_DATA_PATH)")
			}
			n, uuid, err := svc.AddEmpire(*dataPath, *empireNo)
			if err != nil {
				return err
			}
			fmt.Printf("added empire %d, magic link: %s\n", n, uuid)
			return nil
		},
	}
}

// CmdRemoveEmpire returns an ff.Command that deactivates an empire.
func CmdRemoveEmpire(svc *app.GameConfigService) *ff.Command {
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
func CmdShowMagicLink(svc *app.GameConfigService) *ff.Command {
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
