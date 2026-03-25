// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"fmt"
	"os"

	"github.com/mdhender/ec/internal/app"
	"github.com/spf13/cobra"
)

// CmdCreateGame returns a cobra command that initializes game.json and auth.json in a directory.
func CmdCreateGame(svc *app.GameConfigService) *cobra.Command {
	var path string

	cmd := &cobra.Command{
		Use:   "game",
		Short: "initialize game.json and auth.json in a directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.CreateGame(path); err != nil {
				return err
			}
			fmt.Printf("game created: %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "directory to write game.json and auth.json")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

// CmdAddEmpire returns a cobra command that adds an empire to game.json and auth.json.
func CmdAddEmpire(svc *app.GameConfigService) *cobra.Command {
	var path string
	var empireNo int

	cmd := &cobra.Command{
		Use:   "empire",
		Short: "add an empire to game.json and auth.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			n, uuid, err := svc.AddEmpire(path, empireNo)
			if err != nil {
				return err
			}
			fmt.Printf("added empire %d, magic link: %s\n", n, uuid)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "directory containing game.json and auth.json")
	cmd.Flags().IntVar(&empireNo, "empire", 0, "empire number (0 = auto-assign)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

// CmdRemoveEmpire returns a cobra command that deactivates an empire.
func CmdRemoveEmpire(svc *app.GameConfigService) *cobra.Command {
	var path string
	var empireNo int

	cmd := &cobra.Command{
		Use:   "empire",
		Short: "deactivate an empire in game.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.RemoveEmpire(path, empireNo); err != nil {
				return err
			}
			fmt.Printf("removed empire %d\n", empireNo)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "directory containing game.json and auth.json")
	cmd.Flags().IntVar(&empireNo, "empire", 0, "empire number")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagRequired("empire")
	return cmd
}

// CmdShowMagicLink returns a cobra command that prints the magic link URL for an empire.
func CmdShowMagicLink(svc *app.GameConfigService) *cobra.Command {
	var path string
	var empireNo int
	var baseURL string

	cmd := &cobra.Command{
		Use:   "magic-link",
		Short: "show the magic link URL for an empire",
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid, err := svc.ShowMagicLink(path, empireNo)
			if err != nil {
				return err
			}
			fmt.Printf("%s?magic=%s\n", baseURL, uuid)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "directory containing auth.json")
	cmd.Flags().IntVar(&empireNo, "empire", 0, "empire number")
	cmd.Flags().StringVar(&baseURL, "base-url", os.Getenv("EC_BASE_URL"), "application base URL (env: EC_BASE_URL)")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagRequired("empire")
	_ = cmd.MarkFlagRequired("base-url")
	return cmd
}
