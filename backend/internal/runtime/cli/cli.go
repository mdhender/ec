// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"github.com/mdhender/ec/internal/app"
	deliverycli "github.com/mdhender/ec/internal/delivery/cli"
	"github.com/mdhender/ec/internal/infra/filestore"
	"github.com/peterbourgon/ff/v4"
)

// BuildCommands creates the CLI command tree and returns the top-level
// subcommands. The caller (cmd/cli/main.go) attaches them to the root.
func BuildCommands() []*ff.Command {
	store := filestore.NewStore("")

	clusterSvc := &app.ClusterService{
		Writer: store,
	}

	gameSvc := &app.GameService{
		Store:   store,
		Cluster: store,
	}

	createCmd := &ff.Command{
		Name:      "create",
		Usage:     "cli create SUBCOMMAND ...",
		ShortHelp: "create game objects",
		Subcommands: []*ff.Command{
			deliverycli.CmdCreateCluster(clusterSvc),
			deliverycli.CmdCreateGame(gameSvc),
			deliverycli.CmdCreateHomeWorld(gameSvc),
			deliverycli.CmdAddEmpire(gameSvc),
		},
	}

	removeCmd := &ff.Command{
		Name:      "remove",
		Usage:     "cli remove SUBCOMMAND ...",
		ShortHelp: "remove game objects",
		Subcommands: []*ff.Command{
			deliverycli.CmdRemoveEmpire(gameSvc),
		},
	}

	showCmd := &ff.Command{
		Name:      "show",
		Usage:     "cli show SUBCOMMAND ...",
		ShortHelp: "show things",
		Subcommands: []*ff.Command{
			deliverycli.CmdShowMagicLink(gameSvc),
		},
	}

	testCmd := &ff.Command{
		Name:      "test",
		Usage:     "cli test SUBCOMMAND ...",
		ShortHelp: "test game distributions",
		Subcommands: []*ff.Command{
			deliverycli.CmdTestCluster(clusterSvc),
		},
	}

	return []*ff.Command{createCmd, removeCmd, showCmd, testCmd}
}
