package main

import (
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/git"
	"github.com/datamaia/andromeda/internal/indexer"
	"github.com/datamaia/andromeda/internal/memory"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/spf13/cobra"
)

// context — summarize what the agent would see: workspace, VCS, memory, and index coverage.
func newContextCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "context",
		Short: "Summarize the assembled workspace context",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "workspace: %s\n", wd)

			if st, err := git.New("").Status(ctx, ports.RepoRef{Root: wd}); err == nil {
				clean := "dirty"
				if st.Clean {
					clean = "clean"
				}
				_, _ = fmt.Fprintf(out, "vcs:       branch %s (%s)\n", st.Branch, clean)
			}

			db, err := storage.OpenWorkspaceDB(ctx, wd)
			if err == nil {
				defer func() { _ = db.Close() }()
				recs, _ := memory.New(db).Retrieve(ctx, ports.MemoryQuery{Limit: 1000})
				_, _ = fmt.Fprintf(out, "memory:    %d records\n", len(recs))
			}

			e := indexer.New()
			if id, err := e.Build(ctx, ports.IndexSpec{Include: []ports.Path{wd}}); err == nil {
				st, _ := e.Status(ctx, id)
				_, _ = fmt.Fprintf(out, "index:     %d files (generation %d)\n", st.Coverage, st.Generation)
			}
			return nil
		},
	}
}

// trace — show the persisted event trail for a run (by run ID).
func newTraceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "trace <run-id>",
		Short: "Show the persisted event trail for a run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withWorkspaceDB(cmd.Context(), func(db *storage.DB) error {
				rows, err := db.SQL().QueryContext(cmd.Context(),
					`SELECT ts, name, producer FROM events WHERE run_id = ? ORDER BY id`, args[0])
				if err != nil {
					return err
				}
				defer rows.Close()
				out := cmd.OutOrStdout()
				var n int
				for rows.Next() {
					var ts, name, producer string
					if err := rows.Scan(&ts, &name, &producer); err != nil {
						return err
					}
					_, _ = fmt.Fprintf(out, "%s  %-28s  %s\n", ts, name, producer)
					n++
				}
				if n == 0 {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "no events for run %s\n", args[0])
				}
				return rows.Err()
			})
		},
	}
}
