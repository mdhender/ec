// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"fmt"

	"github.com/mdhender/ec"
	"github.com/mdhender/ec/internal/app"
	deliverycli "github.com/mdhender/ec/internal/delivery/cli"
	"github.com/mdhender/ec/internal/infra/filestore"
	"github.com/spf13/cobra"
)

// AddCommands adds command groups to the root command.
// It wires infra adapters and app services.
func AddCommands(root *cobra.Command) {
	store := filestore.NewStore("")

	clusterSvc := &app.ClusterService{
		Reader:     store,
		Writer:     store,
		GameWriter: store,
	}

	gameConfigSvc := &app.GameConfigService{
		Store: store,
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create game objects",
	}
	createCmd.AddCommand(deliverycli.CmdCreateCluster(clusterSvc))
	createCmd.AddCommand(deliverycli.CmdCreateGameState(clusterSvc))
	createCmd.AddCommand(deliverycli.CmdCreateGame(gameConfigSvc))
	createCmd.AddCommand(deliverycli.CmdAddEmpire(gameConfigSvc))

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "remove game objects",
	}
	removeCmd.AddCommand(deliverycli.CmdRemoveEmpire(gameConfigSvc))

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "show things",
	}
	showCmd.AddCommand(deliverycli.CmdShowMagicLink(gameConfigSvc))

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "test game distributions",
	}
	testCmd.AddCommand(deliverycli.CmdTestCluster(clusterSvc))

	showBuildInfo := false
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "display the application's version number",
		RunE: func(cmd *cobra.Command, args []string) error {
			if showBuildInfo {
				fmt.Println(ec.Version().String())
				return nil
			}
			fmt.Println(ec.Version().Core())
			return nil
		},
	}
	versionCmd.Flags().BoolVar(&showBuildInfo, "build-info", showBuildInfo, "show build information")

	root.AddCommand(createCmd)
	root.AddCommand(removeCmd)
	root.AddCommand(showCmd)
	root.AddCommand(testCmd)
	root.AddCommand(versionCmd)
}
