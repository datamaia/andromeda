package main

import (
	"context"
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/spf13/cobra"
)

// logs — show persisted events from the workspace event store.
func newLogsCommand() *cobra.Command {
	var limit int
	c := &cobra.Command{
		Use:   "logs",
		Short: "Show recent persisted events in this workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withWorkspaceDB(cmd.Context(), func(db *storage.DB) error {
				rows, err := db.SQL().QueryContext(cmd.Context(),
					`SELECT ts, name, producer FROM events ORDER BY id DESC LIMIT ?`, limit)
				if err != nil {
					return err
				}
				defer rows.Close()
				out := cmd.OutOrStdout()
				for rows.Next() {
					var ts, name, producer string
					if err := rows.Scan(&ts, &name, &producer); err != nil {
						return err
					}
					fmt.Fprintf(out, "%s  %-28s  %s\n", ts, name, producer)
				}
				return rows.Err()
			})
		},
	}
	c.Flags().IntVarP(&limit, "limit", "n", 20, "maximum events to show")
	return c
}

// export — export sessions and events as JSON for portability.
func newExportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export workspace sessions as JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withWorkspaceDB(cmd.Context(), func(db *storage.DB) error {
				store := storage.NewSessionStore(db)
				sessions, err := store.ListSessions(cmd.Context(), ports.SessionFilter{Limit: 1000})
				if err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				fmt.Fprintln(out, "[")
				for i, s := range sessions {
					comma := ","
					if i == len(sessions)-1 {
						comma = ""
					}
					fmt.Fprintf(out, `  {"id":%q,"state":%q,"created_at":%q}%s`+"\n", s.ID, s.State, s.CreatedAt, comma)
				}
				fmt.Fprintln(out, "]")
				return nil
			})
		},
	}
}

func withWorkspaceDB(ctx context.Context, fn func(*storage.DB) error) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	db, err := storage.OpenWorkspaceDB(ctx, wd)
	if err != nil {
		return err
	}
	defer db.Close()
	return fn(db)
}
