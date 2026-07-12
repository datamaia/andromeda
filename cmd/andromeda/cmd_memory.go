package main

import (
	"context"
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/memory"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/spf13/cobra"
)

func newMemoryCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "memory", Short: "Manage workspace memory"}
	cmd.AddCommand(newMemoryAddCommand(), newMemoryListCommand(), newMemorySearchCommand())
	return cmd
}

func withMemory(ctx context.Context, fn func(*memory.Store) error) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	db, err := storage.OpenWorkspaceDB(ctx, wd)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	return fn(memory.New(db))
}

func newMemoryAddCommand() *cobra.Command {
	var layer string
	c := &cobra.Command{
		Use:   "add <content>",
		Short: "Add a memory record",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content := joinArgs(args)
			return withMemory(cmd.Context(), func(s *memory.Store) error {
				ids, err := s.Ingest(cmd.Context(), []ports.MemoryRecordDraft{{Layer: layer, Content: content, Source: "cli"}})
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "added %s\n", ids[0])
				return nil
			})
		},
	}
	c.Flags().StringVar(&layer, "layer", "workspace", "memory layer")
	return c
}

func newMemoryListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List memory records",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withMemory(cmd.Context(), func(s *memory.Store) error {
				recs, err := s.Retrieve(cmd.Context(), ports.MemoryQuery{Limit: 50})
				if err != nil {
					return err
				}
				for _, r := range recs {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", r.Layer, r.Content)
				}
				return nil
			})
		},
	}
}

func newMemorySearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search memory records",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			q := joinArgs(args)
			return withMemory(cmd.Context(), func(s *memory.Store) error {
				recs, err := s.Retrieve(cmd.Context(), ports.MemoryQuery{Text: q, Limit: 50})
				if err != nil {
					return err
				}
				for _, r := range recs {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", r.Layer, r.Content)
				}
				return nil
			})
		},
	}
}
