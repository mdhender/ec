// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mdhender/ec"
	runtimecli "github.com/mdhender/ec/internal/runtime/cli"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffenv"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

func main() {
	rootFlags := ff.NewFlagSet("cli")
	_ = rootFlags.StringLong("log-level", "info", "log level (debug|info|warn|error)")
	_ = rootFlags.BoolLong("log-source", "log source file and line")
	_ = rootFlags.BoolLong("debug", "enable debug logging (same as --log-level=debug)")
	_ = rootFlags.BoolLong("quiet", "only log errors (same as --log-level=error)")

	versionFlags := ff.NewFlagSet("version").SetParent(rootFlags)
	buildInfo := versionFlags.BoolLong("build-info", "show build information")
	versionCmd := &ff.Command{
		Name:      "version",
		Usage:     "cli version [FLAGS]",
		ShortHelp: "display the application's version number",
		Flags:     versionFlags,
		Exec: func(ctx context.Context, args []string) error {
			if *buildInfo {
				fmt.Println(ec.Version().String())
				return nil
			}
			fmt.Println(ec.Version().Core())
			return nil
		},
	}

	subcommands := runtimecli.BuildCommands()
	subcommands = append(subcommands, versionCmd)

	rootCmd := &ff.Command{
		Name:        "cli",
		Usage:       "cli [FLAGS] SUBCOMMAND ...",
		ShortHelp:   "yet another game engine command line interface",
		Flags:       rootFlags,
		Subcommands: subcommands,
	}

	err := rootCmd.ParseAndRun(context.Background(), os.Args[1:],
		ff.WithEnvVarPrefix("EC"),
		ff.WithConfigFile(".env"),
		ff.WithConfigFileParser(ffenv.Parse),
		ff.WithConfigAllowMissingFile(),
		ff.WithConfigIgnoreFlagNames(),
		ff.WithConfigIgnoreUndefinedFlags(),
	)
	if err != nil {
		if errors.Is(err, ff.ErrHelp) {
			fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Command(rootCmd))
			os.Exit(0)
		}
		if errors.Is(err, ff.ErrNoExec) {
			fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Command(rootCmd))
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
