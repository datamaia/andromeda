package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/datamaia/andromeda/internal/indexer"
	"github.com/datamaia/andromeda/internal/ports"
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
				_, _ = fmt.Fprintf(out, "%-14s %-8s %s\n", t.name, t.perms, t.desc)
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
		{"fs_diff", "read", "Compute a unified diff"},
		{"fs_replace", "read,write", "Replace text inside a file (requires --allow-write)"},
		{"fs_patch", "write", "Apply a unified diff atomically (requires --allow-write)"},
		{"git_exec", "read,git_mutation", "Run a structured Git operation (mutations require --allow-write)"},
		{"terminal_run", "execute", "Run a shell command (requires --allow-exec)"},
		{"process_control", "process_spawn", "List/inspect/signal/terminate supervised processes (requires --allow-exec)"},
		{"sqlite_query", "read,write", "Run SQL against a workspace SQLite database"},
		{"http_request", "network,credential_access", "Perform one HTTP request (requires --allow-network)"},
		{"docker_control", "container_access,network", "Operate the local Docker Engine (Beta; needs docker)"},
		{"kubernetes_control", "container_access,network,execute", "Operate Kubernetes via kubectl (v1; needs kubectl)"},
		{"browser_control", "network,process_spawn", "Drive a browser via W3C WebDriver (v1; needs a driver endpoint)"},
		{"github_request", "external_service_access,network", "GitHub official API transport (Beta; needs [services.github])"},
		{"gitlab_request", "external_service_access,network", "GitLab official API transport (v1; needs [services.gitlab])"},
		{"jira_request", "external_service_access,network", "Jira official API transport (v1; needs [services.jira])"},
		{"slack_request", "external_service_access,network,notifications", "Slack official API transport (v1; needs [services.slack])"},
		{"notion_request", "external_service_access,network", "Notion official API transport (v2; needs [services.notion])"},
		{"linear_request", "external_service_access,network", "Linear official API transport (v2; needs [services.linear])"},
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
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[indexed %d files, generation %d]\n", st.Coverage, st.Generation)
			for _, h := range hits {
				_, _ = fmt.Fprintf(out, "%.2f  %s\n", h.Score, h.Path)
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
		Short: "Check for updates and how to apply them",
		Long: "Report the running version and check the release feed for a newer one. Upgrades are " +
			"applied through Homebrew or the install script, not in place, so this command tells you " +
			"what to run rather than replacing the binary.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			self, _ := os.Executable()
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), checkForUpdate(cmd.Context(), channel, self))
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
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
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
