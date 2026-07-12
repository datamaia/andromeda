package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/datamaia/andromeda/internal/buildinfo"
	"github.com/datamaia/andromeda/internal/indexer"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/updater"
	"github.com/spf13/cobra"
)

func joinArgs(args []string) string { return strings.Join(args, " ") }

// tool list — enumerate the built-in tools available to the agent.
func newToolCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "tool", Short: "Inspect available tools"}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List built-in tools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			for _, t := range builtinToolSummaries() {
				fmt.Fprintf(out, "%-14s %-8s %s\n", t.name, t.perms, t.desc)
			}
			return nil
		},
	})
	return cmd
}

type toolSummary struct{ name, perms, desc string }

func builtinToolSummaries() []toolSummary {
	return []toolSummary{
		{"fs_read", "read", "Read a text file"},
		{"fs_write", "write", "Write a text file (requires --allow-write)"},
		{"fs_search", "read", "Search files for a substring"},
		{"terminal_run", "execute", "Run a shell command (requires --allow-exec)"},
	}
}

// index build/query — lexical index over the current workspace.
func newIndexCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "index", Short: "Build and query the workspace index"}
	cmd.AddCommand(&cobra.Command{
		Use:   "query <text>",
		Short: "Build a fresh lexical index and query it",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			e := indexer.New()
			id, err := e.Build(cmd.Context(), ports.IndexSpec{Include: []ports.Path{wd}})
			if err != nil {
				return err
			}
			hits, err := e.Query(cmd.Context(), id, ports.IndexQuery{Text: joinArgs(args), MaxResults: 20})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			st, _ := e.Status(context.Background(), id)
			fmt.Fprintf(cmd.ErrOrStderr(), "[indexed %d files, generation %d]\n", st.Coverage, st.Generation)
			for _, h := range hits {
				fmt.Fprintf(out, "%.2f  %s\n", h.Score, h.Path)
			}
			return nil
		},
	})
	return cmd
}

// update — check for a newer release on the configured channel.
func newUpdateCommand() *cobra.Command {
	var channel string
	c := &cobra.Command{
		Use:   "update",
		Short: "Check for and apply updates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			self, _ := os.Executable()
			// No release source is wired for a from-source dev build; Check reports cleanly.
			u := updater.New(buildinfo.Get().Version, channel, self, nil)
			res, err := u.Check(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			switch res.Status {
			case "update_available":
				fmt.Fprintf(out, "update available: %s → %s (channel %s)\n", res.Current, res.Latest, res.Channel)
			default:
				fmt.Fprintf(out, "up to date: %s (channel %s)\n", res.Current, res.Channel)
			}
			return nil
		},
	}
	c.Flags().StringVar(&channel, "channel", "stable", "release channel: stable|beta|nightly|rc")
	return c
}

// completion — generate shell completion scripts.
func newCompletionCommand(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:       "completion [bash|zsh|fish]",
		Short:     "Generate a shell completion script",
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(out, true)
			case "zsh":
				return root.GenZshCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, true)
			}
			return fmt.Errorf("unsupported shell %q", args[0])
		},
	}
}
