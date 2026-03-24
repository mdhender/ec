// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/mdhender/ec"
	"github.com/mdhender/ec/internal/dotfiles"
	"github.com/spf13/cobra"
)

var (
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
)

func main() {
	if err := dotfiles.Load("EC", logger); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	addFlags := func(cmd *cobra.Command) error {
		cmd.PersistentFlags().Bool("debug", false, "enable debug logging (same as --log-level=debug)")
		cmd.PersistentFlags().Bool("info", false, "enable info logging (same as --log-level=info)")
		cmd.PersistentFlags().Bool("quiet", false, "only log errors (same as --log-level=error)")
		cmd.PersistentFlags().String("log-level", "info", "logging level (debug|info|warn|error))")
		cmd.PersistentFlags().Bool("log-source", false, "log source file and line")
		return nil
	}

	var cmdRoot = &cobra.Command{
		Short:         "cli - yet another game engine command line interface",
		Long:          `cli is the CLI for EC.`,
		Version:       ec.Version().Core(),
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Root().PersistentFlags()
			addSource, err := flags.GetBool("log-source")
			if err != nil {
				return err
			}
			logLevel, err := flags.GetString("log-level")
			if err != nil {
				return err
			}
			debug, err := flags.GetBool("debug")
			if err != nil {
				return err
			}
			quiet, err := flags.GetBool("quiet")
			if err != nil {
				return err
			}
			if debug && quiet {
				return fmt.Errorf("--debug and --quiet are mutually exclusive")
			}
			var lvl slog.Level
			switch {
			case debug:
				lvl = slog.LevelDebug
			case quiet:
				lvl = slog.LevelError
			default:
				switch strings.ToLower(logLevel) {
				case "debug":
					lvl = slog.LevelDebug
				case "info":
					lvl = slog.LevelInfo
				case "warn", "warning":
					lvl = slog.LevelWarn
				case "error":
					lvl = slog.LevelError
				default:
					return fmt.Errorf("log-level: unknown value %q", logLevel)
				}
			}
			handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level:     lvl,
				AddSource: addSource || lvl == slog.LevelDebug,
			})
			logger = slog.New(handler)
			slog.SetDefault(logger) // optional, but convenient
			return nil
		},
	}

	cmdRoot.AddCommand(cmdShow())
	err := addFlags(cmdRoot)
	if err != nil {
		logger.Error("root: addFlags",
			"err", err,
		)
		os.Exit(1)
	}

	err = cmdRoot.Execute()
	if err != nil {
		if quiet, _ := cmdRoot.Flags().GetBool("quiet"); !quiet {
			logger.Error("command failed", "err", err)
		}
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

// resolveDuration returns the flag value if explicitly set, otherwise
// parses the env var, otherwise returns the fallback.
func resolveDuration(cmd *cobra.Command, flagName, envVar string, fallback time.Duration) (time.Duration, error) {
	if cmd.Flags().Changed(flagName) {
		return cmd.Flags().GetDuration(flagName)
	}
	if v := os.Getenv(envVar); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return 0, fmt.Errorf("--%s: invalid duration %q from %s: %w", flagName, v, envVar, err)
		}
		return d, nil
	}
	return fallback, nil
}

// resolveString returns the flag value if the flag was explicitly set,
// otherwise the env var value, otherwise the fallback.
// Priority: flag → env → fallback.
func resolveString(cmd *cobra.Command, flagName, envVar, fallback string) string {
	if cmd.Flags().Changed(flagName) {
		v, _ := cmd.Flags().GetString(flagName)
		return v
	}
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	return fallback
}

func cmdShow() *cobra.Command {
	addFlags := func(cmd *cobra.Command) error {
		return nil
	}
	cmd := &cobra.Command{
		Use:   "show",
		Short: "show things",
	}
	if err := addFlags(cmd); err != nil {
		logger.Error("show: addFlags",
			"err", err,
		)
		os.Exit(1)
	}
	cmd.AddCommand(cmdShowVersion())
	return cmd
}

func cmdShowVersion() *cobra.Command {
	showBuildInfo := false
	addFlags := func(cmd *cobra.Command) error {
		cmd.Flags().BoolVar(&showBuildInfo, "build-info", showBuildInfo, "show build information")
		return nil
	}
	var cmd = &cobra.Command{
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
	if err := addFlags(cmd); err != nil {
		logger.Error("show: version",
			"err", err,
		)
		os.Exit(1)
	}
	return cmd
}
