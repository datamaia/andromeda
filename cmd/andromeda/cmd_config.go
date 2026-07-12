package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect resolved configuration",
	}
	cmd.AddCommand(newConfigShowCommand())
	return cmd
}

func newConfigShowCommand() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "show",
		Short: "Show the effective configuration with source attribution",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			res, err := app.LoadedConfig(cmd.Context(), wd)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{"values": res.Values, "sources": res.Sources})
			}
			keys := make([]string, 0, len(res.Values))
			for k := range res.Values {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				_, _ = fmt.Fprintf(out, "%-32s = %-20v [%s]\n", k, res.Values[k], res.Sources[k])
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return c
}
