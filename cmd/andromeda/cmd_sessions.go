package main

import (
	"fmt"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/spf13/cobra"
)

// newSessionsCommand manages saved TUI conversations: `andromeda sessions [list]` prints them and
// `andromeda sessions rm <id>` deletes one. Resume a session with `andromeda tui --resume <id>`
// (or `--continue` for the most recent).
func newSessionsCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "sessions",
		Short: "List and manage saved TUI sessions",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return listSessions(cmd) },
	}
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List saved sessions (newest first)",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return listSessions(cmd) },
	})
	c.AddCommand(&cobra.Command{
		Use:   "rm <id>",
		Short: "Delete a saved session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.RemoveSession(args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "removed session "+args[0])
			return nil
		},
	})
	return c
}

func listSessions(cmd *cobra.Command) error {
	sessions, err := app.ListSessions()
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if len(sessions) == 0 {
		_, _ = fmt.Fprintln(w, "no saved sessions yet — start one with `andromeda` and it saves as you go")
		return nil
	}
	_, _ = fmt.Fprintf(w, "%-22s %6s  %-20s  %s\n", "ID", "TURNS", "UPDATED", "TITLE")
	for _, s := range sessions {
		_, _ = fmt.Fprintf(w, "%-22s %6d  %-20s  %s\n", s.ID, app.CountTurns(s.Messages), s.UpdatedAt, s.Title)
	}
	return nil
}
