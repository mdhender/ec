// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"log/slog"
	"os"

	"github.com/mdhender/ec/internal/app"
	"github.com/spf13/cobra"
)

// CmdCreateCluster returns a cobra command that generates a cluster and writes it to disk.
func CmdCreateCluster(svc *app.ClusterService) *cobra.Command {
	var path string
	var seed1, seed2 uint64 = 10, 10
	overwrite := false

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "create a new cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			cluster, err := svc.CreateCluster(seed1, seed2, path, overwrite)
			if err != nil {
				return err
			}
			WriteClusterReport(os.Stdout, cluster)
			slog.Info("cluster created", "path", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "testdata/cluster.json", "path to save cluster JSON")
	cmd.Flags().Uint64Var(&seed1, "seed1", seed1, "seed1")
	cmd.Flags().Uint64Var(&seed2, "seed2", seed2, "seed2")
	cmd.Flags().BoolVar(&overwrite, "overwrite", overwrite, "overwrite file if it exists")
	return cmd
}

// CmdTestCluster returns a cobra command that runs N iterations and reports distribution stats.
func CmdTestCluster(svc *app.ClusterService) *cobra.Command {
	var iterations int = 100
	var seed1, seed2 uint64 = 10, 10

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "test cluster generation distributions",
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := svc.TestCluster(seed1, seed2, iterations)
			if err != nil {
				return err
			}
			WriteStatsReport(os.Stdout, stats, iterations)
			return nil
		},
	}
	cmd.Flags().IntVar(&iterations, "iterations", iterations, "number of iterations")
	cmd.Flags().Uint64Var(&seed1, "seed1", seed1, "seed1")
	cmd.Flags().Uint64Var(&seed2, "seed2", seed2, "seed2")
	return cmd
}

// CmdCreateGameState returns a cobra command that creates a game from a cluster file.
func CmdCreateGameState(svc *app.ClusterService) *cobra.Command {
	var clusterPath string
	var savePath string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "game-state",
		Short: "create a new game state from a cluster file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.CreateGame(clusterPath, savePath, overwrite); err != nil {
				return err
			}
			slog.Info("game created", "cluster", clusterPath, "save", savePath)
			return nil
		},
	}
	cmd.Flags().StringVar(&clusterPath, "cluster", "", "path to cluster JSON file")
	cmd.Flags().StringVar(&savePath, "save", "", "path to save game JSON file")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing save file")
	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.MarkFlagRequired("save")
	return cmd
}
