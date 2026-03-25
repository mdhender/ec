// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/mdhender/ec/internal/app"
	"github.com/peterbourgon/ff/v4"
)

// CmdCreateCluster returns an ff.Command that generates a cluster and writes it to disk.
func CmdCreateCluster(svc *app.ClusterService) *ff.Command {
	fs := ff.NewFlagSet("cluster")
	dataPath := fs.StringLong("data-path", "testdata", "directory to save cluster.json")
	seed1Str := fs.StringLong("seed1", "10", "seed1")
	seed2Str := fs.StringLong("seed2", "10", "seed2")
	overwrite := fs.BoolLong("overwrite", "overwrite file if it exists")

	return &ff.Command{
		Name:      "cluster",
		Usage:     "cli create cluster [FLAGS]",
		ShortHelp: "create a new cluster",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			seed1, err := strconv.ParseUint(*seed1Str, 10, 64)
			if err != nil {
				return fmt.Errorf("--seed1: invalid value %q: %w", *seed1Str, err)
			}
			seed2, err := strconv.ParseUint(*seed2Str, 10, 64)
			if err != nil {
				return fmt.Errorf("--seed2: invalid value %q: %w", *seed2Str, err)
			}
			cluster, err := svc.CreateCluster(seed1, seed2, *dataPath, *overwrite)
			if err != nil {
				return err
			}
			WriteClusterReport(os.Stdout, cluster)
			slog.Info("cluster created", "data-path", *dataPath)
			return nil
		},
	}
}

// CmdTestCluster returns an ff.Command that runs N iterations and reports distribution stats.
func CmdTestCluster(svc *app.ClusterService) *ff.Command {
	fs := ff.NewFlagSet("cluster")
	iterations := fs.IntLong("iterations", 100, "number of iterations")
	seed1Str := fs.StringLong("seed1", "10", "seed1")
	seed2Str := fs.StringLong("seed2", "10", "seed2")

	return &ff.Command{
		Name:      "cluster",
		Usage:     "cli test cluster [FLAGS]",
		ShortHelp: "test cluster generation distributions",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			seed1, err := strconv.ParseUint(*seed1Str, 10, 64)
			if err != nil {
				return fmt.Errorf("--seed1: invalid value %q: %w", *seed1Str, err)
			}
			seed2, err := strconv.ParseUint(*seed2Str, 10, 64)
			if err != nil {
				return fmt.Errorf("--seed2: invalid value %q: %w", *seed2Str, err)
			}
			stats, err := svc.TestCluster(seed1, seed2, *iterations)
			if err != nil {
				return err
			}
			WriteStatsReport(os.Stdout, stats, *iterations)
			return nil
		},
	}
}
