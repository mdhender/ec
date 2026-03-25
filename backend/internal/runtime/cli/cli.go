// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"github.com/mdhender/ec/internal/app"
	deliverycli "github.com/mdhender/ec/internal/delivery/cli"
	"github.com/mdhender/ec/internal/infra/filestore"
	"github.com/spf13/cobra"
)

// AddCommands adds the create and test command groups to the root command.
// It wires infra adapters and app services.
func AddCommands(root *cobra.Command) {
	store := filestore.NewStore("")

	clusterSvc := &app.ClusterService{
		Reader:     store,
		Writer:     store,
		GameWriter: store,
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create game objects",
	}
	createCmd.AddCommand(deliverycli.CmdCreateCluster(clusterSvc))
	createCmd.AddCommand(deliverycli.CmdCreateGame(clusterSvc))

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "test game distributions",
	}
	testCmd.AddCommand(deliverycli.CmdTestCluster(clusterSvc))

	root.AddCommand(createCmd)
	root.AddCommand(testCmd)
}
