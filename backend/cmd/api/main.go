// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdhender/ec"
	"github.com/mdhender/ec/internal/dotfiles"
	"github.com/mdhender/ec/internal/fsck"
	"github.com/mdhender/ec/internal/infra/auth"
	"github.com/mdhender/ec/internal/infra/filestore"
	"github.com/mdhender/ec/internal/runtime/server"
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
		Short:         "api - yet another game engine API server",
		Long:          `api is the API for EC.`,
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

	cmdRoot.AddCommand(cmdServe())
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

func cmdServe() *cobra.Command {
	var (
		host        string
		port        string
		dataPath    string
		jwtSecret   string
		shutdownKey string
		timeout     time.Duration
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "start the API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if host == "" {
				host = os.Getenv("EC_HOST")
			}
			if host == "" {
				host = "localhost"
			}
			if port == "" {
				port = os.Getenv("EC_PORT")
			}
			if port == "" {
				port = "8080"
			}
			if dataPath == "" {
				dataPath = os.Getenv("EC_DATA_PATH")
			}
			if dataPath == "" {
				return fmt.Errorf("serve: --data-path is required (or set EC_DATA_PATH)")
			}
			if jwtSecret == "" {
				jwtSecret = os.Getenv("EC_JWT_SECRET")
			}
			if jwtSecret == "" {
				return fmt.Errorf("serve: --jwt-secret is required (or set EC_JWT_SECRET)")
			}
			if shutdownKey == "" {
				shutdownKey = os.Getenv("EC_SHUTDOWN_KEY")
			}
			if !fsck.IsDir(dataPath) {
				return fmt.Errorf("serve: data-path %q is not a directory", dataPath)
			}

			authStore, err := auth.NewMagicLinkStore(filepath.Join(dataPath, "auth.json"))
			if err != nil {
				return fmt.Errorf("serve: load auth: %w", err)
			}

			jwtMgr := auth.NewJWTManager(jwtSecret, 24*time.Hour)
			fileStore := filestore.NewStore(dataPath)

			opts := []server.Option{
				server.WithHost(host),
				server.WithPort(port),
				server.WithShutdownKey(shutdownKey),
				server.WithJWTMiddleware(jwtMgr.Middleware()),
				server.WithAuthStore(authStore),
				server.WithTokenSigner(jwtMgr),
				server.WithOrderStore(fileStore),
				server.WithReportStore(fileStore),
			}
			if timeout > 0 {
				opts = append(opts, server.WithShutdownAfter(timeout))
			}

			srv, err := server.New(opts...)
			if err != nil {
				return fmt.Errorf("serve: create server: %w", err)
			}
			return srv.Start()
		},
	}
	cmd.Flags().StringVar(&host, "host", "", "listen host (default: localhost, env: EC_HOST)")
	cmd.Flags().StringVar(&port, "port", "", "listen port (default: 8080, env: EC_PORT)")
	cmd.Flags().StringVar(&dataPath, "data-path", "", "path to data directory (env: EC_DATA_PATH)")
	cmd.Flags().StringVar(&jwtSecret, "jwt-secret", "", "HMAC secret for JWT signing (env: EC_JWT_SECRET)")
	cmd.Flags().StringVar(&shutdownKey, "shutdown-key", "", "secret key to enable /api/shutdown endpoint (env: EC_SHUTDOWN_KEY)")
	cmd.Flags().DurationVar(&timeout, "timeout", 0, "auto-shutdown after duration (0 = disabled)")
	return cmd
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
