// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/mdhender/ec/internal/app"
	"github.com/peterbourgon/ff/v4"
)

// CmdParseOrders returns an ff.Command that parses one or more order files
// and prints diagnostics to stdout.
func CmdParseOrders(svc *app.ParseOrdersService) *ff.Command {
	fs := ff.NewFlagSet("orders")

	return &ff.Command{
		Name:      "orders",
		Usage:     "cli parse orders FILE [FILE ...]",
		ShortHelp: "parse order files and report diagnostics",
		Flags:     fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("at least one order file is required")
			}

			hasErrors := false
			for _, path := range args {
				data, err := os.ReadFile(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
					hasErrors = true
					continue
				}

				result, err := svc.Parse(string(data))
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s: parser error: %v\n", path, err)
					hasErrors = true
					continue
				}

				fmt.Printf("%s: %d orders, %d diagnostics\n", path, len(result.Orders), len(result.Diagnostics))
				for _, d := range result.Diagnostics {
					fmt.Printf("  line %d [%s]: %s\n", d.Line, d.Code, d.Message)
				}
				if len(result.Diagnostics) > 0 {
					hasErrors = true
				}
			}

			if hasErrors {
				return fmt.Errorf("one or more files had errors")
			}
			return nil
		},
	}
}
