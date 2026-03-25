// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffenv"
)

func TestEnvPrecedence(t *testing.T) {
	t.Run("string flag", func(t *testing.T) {
		tests := []struct {
			name     string
			args     []string
			envVal   string
			dotenv   string
			want     string
		}{
			{name: "default only", want: "fallback"},
			{name: "dotenv beats default", dotenv: "EC_MY_FLAG=from-dotenv", want: "from-dotenv"},
			{name: "env beats dotenv", envVal: "from-env", dotenv: "EC_MY_FLAG=from-dotenv", want: "from-env"},
			{name: "flag beats env", args: []string{"--my-flag", "from-flag"}, envVal: "from-env", want: "from-flag"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fs := ff.NewFlagSet("test")
				val := fs.StringLong("my-flag", "fallback", "test flag")

				if tt.envVal != "" {
					t.Setenv("EC_MY_FLAG", tt.envVal)
				}

				opts := []ff.Option{ff.WithEnvVarPrefix("EC")}
				if tt.dotenv != "" {
					dir := t.TempDir()
					envFile := filepath.Join(dir, ".env")
					if err := os.WriteFile(envFile, []byte(tt.dotenv+"\n"), 0o644); err != nil {
						t.Fatal(err)
					}
					opts = append(opts,
						ff.WithConfigFile(envFile),
						ff.WithConfigFileParser(ffenv.Parse),
						ff.WithConfigIgnoreFlagNames(),
					)
				}

				if err := ff.Parse(fs, tt.args, opts...); err != nil {
					t.Fatal(err)
				}
				if got := *val; got != tt.want {
					t.Errorf("got %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("duration flag", func(t *testing.T) {
		tests := []struct {
			name   string
			args   []string
			envVal string
			dotenv string
			want   time.Duration
		}{
			{name: "default only", want: 5 * time.Second},
			{name: "dotenv beats default", dotenv: "EC_MY_DUR=10s", want: 10 * time.Second},
			{name: "env beats dotenv", envVal: "30s", dotenv: "EC_MY_DUR=10s", want: 30 * time.Second},
			{name: "flag beats env", args: []string{"--my-dur", "1m"}, envVal: "30s", want: time.Minute},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fs := ff.NewFlagSet("test")
				val := fs.DurationLong("my-dur", 5*time.Second, "test duration")

				if tt.envVal != "" {
					t.Setenv("EC_MY_DUR", tt.envVal)
				}

				opts := []ff.Option{ff.WithEnvVarPrefix("EC")}
				if tt.dotenv != "" {
					dir := t.TempDir()
					envFile := filepath.Join(dir, ".env")
					if err := os.WriteFile(envFile, []byte(tt.dotenv+"\n"), 0o644); err != nil {
						t.Fatal(err)
					}
					opts = append(opts,
						ff.WithConfigFile(envFile),
						ff.WithConfigFileParser(ffenv.Parse),
						ff.WithConfigIgnoreFlagNames(),
					)
				}

				if err := ff.Parse(fs, tt.args, opts...); err != nil {
					t.Fatal(err)
				}
				if got := *val; got != tt.want {
					t.Errorf("got %v, want %v", got, tt.want)
				}
			})
		}
	})
}
