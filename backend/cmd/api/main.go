// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/mdhender/ec"
	"github.com/mdhender/ec/internal/runtime/server"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffenv"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

func main() {
	rootFlags := ff.NewFlagSet("api")
	logLevel := rootFlags.StringLong("log-level", "info", "log level (debug|info|warn|error)")
	logSource := rootFlags.BoolLong("log-source", "log source file and line")
	debug := rootFlags.BoolLong("debug", "enable debug logging (same as --log-level=debug)")
	quiet := rootFlags.BoolLong("quiet", "only log errors (same as --log-level=error)")

	serveFlags := ff.NewFlagSet("serve").SetParent(rootFlags)
	host := serveFlags.StringLong("host", "localhost", "listen host")
	port := serveFlags.StringLong("port", "8080", "listen port")
	dataPath := serveFlags.StringLong("data-path", "", "path to data directory")
	jwtSecret := serveFlags.StringLong("jwt-secret", "", "HMAC secret for JWT signing")
	shutdownKey := serveFlags.StringLong("shutdown-key", "", "secret key to enable /api/shutdown endpoint")
	timeout := serveFlags.DurationLong("timeout", 0, "auto-shutdown after duration, 0 = disabled")

	serveCmd := &ff.Command{
		Name:      "serve",
		Usage:     "api serve [FLAGS]",
		ShortHelp: "start the API server",
		Flags:     serveFlags,
		Exec: func(ctx context.Context, args []string) error {
			logger, err := buildLogger(*logLevel, *logSource, *debug, *quiet)
			if err != nil {
				return err
			}
			slog.SetDefault(logger)

			if *dataPath == "" {
				return fmt.Errorf("serve: --data-path is required (or set EC_DATA_PATH)")
			}
			if *jwtSecret == "" {
				return fmt.Errorf("serve: --jwt-secret is required (or set EC_JWT_SECRET)")
			}
			if fi, err := os.Stat(*dataPath); err != nil || !fi.IsDir() {
				return fmt.Errorf("serve: data-path %q is not a directory", *dataPath)
			}

			opts := []server.Option{
				server.WithHost(*host),
				server.WithPort(*port),
				server.WithShutdownKey(*shutdownKey),
				server.WithDataPath(*dataPath),
				server.WithJWTSecret(*jwtSecret),
			}
			if *timeout > 0 {
				opts = append(opts, server.WithShutdownAfter(*timeout))
			}

			srv, err := server.New(opts...)
			if err != nil {
				return fmt.Errorf("serve: create server: %w", err)
			}
			return srv.Start()
		},
	}

	showFlags := ff.NewFlagSet("show").SetParent(rootFlags)
	showCmd := &ff.Command{
		Name:      "show",
		Usage:     "api show SUBCOMMAND ...",
		ShortHelp: "show things",
		Flags:     showFlags,
	}

	versionFlags := ff.NewFlagSet("version").SetParent(showFlags)
	buildInfo := versionFlags.BoolLong("build-info", "show build information")
	versionCmd := &ff.Command{
		Name:      "version",
		Usage:     "api show version [FLAGS]",
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
	showCmd.Subcommands = []*ff.Command{versionCmd}

	rootCmd := &ff.Command{
		Name:      "api",
		Usage:     "api [FLAGS] SUBCOMMAND ...",
		ShortHelp: "yet another game engine API server",
		Flags:     rootFlags,
		Subcommands: []*ff.Command{serveCmd, showCmd},
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

func buildLogger(level string, addSource, debug, quiet bool) (*slog.Logger, error) {
	if debug && quiet {
		return nil, fmt.Errorf("--debug and --quiet are mutually exclusive")
	}
	var lvl slog.Level
	switch {
	case debug:
		lvl = slog.LevelDebug
	case quiet:
		lvl = slog.LevelError
	default:
		switch strings.ToLower(level) {
		case "debug":
			lvl = slog.LevelDebug
		case "info":
			lvl = slog.LevelInfo
		case "warn", "warning":
			lvl = slog.LevelWarn
		case "error":
			lvl = slog.LevelError
		default:
			return nil, fmt.Errorf("log-level: unknown value %q", level)
		}
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     lvl,
		AddSource: addSource || lvl == slog.LevelDebug,
	})
	return slog.New(handler), nil
}
