package main

import (
	"context"
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/git"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/spf13/cobra"
)

func newGitCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "git", Short: "Read-only Git operations on the current repository"}
	cmd.AddCommand(newGitStatusCommand(), newGitLogCommand())
	return cmd
}

func repoRef() (ports.RepoRef, error) {
	wd, err := os.Getwd()
	if err != nil {
		return ports.RepoRef{}, err
	}
	return ports.RepoRef{Root: wd}, nil
}

func newGitStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show working-tree status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo, err := repoRef()
			if err != nil {
				return err
			}
			st, err := git.New("").Status(cmd.Context(), repo)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "On branch %s\n", st.Branch)
			if st.Clean {
				_, _ = fmt.Fprintln(out, "nothing to commit, working tree clean")
				return nil
			}
			for _, p := range st.Staged {
				_, _ = fmt.Fprintf(out, "  staged:    %s\n", p)
			}
			for _, p := range st.Unstaged {
				_, _ = fmt.Fprintf(out, "  modified:  %s\n", p)
			}
			for _, p := range st.Untracked {
				_, _ = fmt.Fprintf(out, "  untracked: %s\n", p)
			}
			return nil
		},
	}
}

func newGitLogCommand() *cobra.Command {
	var maxN int
	c := &cobra.Command{
		Use:   "log",
		Short: "Show recent commits",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo, err := repoRef()
			if err != nil {
				return err
			}
			st, err := git.New("").Log(cmd.Context(), repo, ports.LogSpec{Max: maxN})
			if err != nil {
				return err
			}
			defer func() { _ = st.Close() }()
			out := cmd.OutOrStdout()
			for {
				ci, err := st.Next(context.Background())
				if err == ports.ErrEndOfStream {
					break
				}
				if err != nil {
					return err
				}
				short := ci.ID
				if len(short) > 8 {
					short = short[:8]
				}
				_, _ = fmt.Fprintf(out, "%s  %s  %s\n", short, ci.Date, ci.Subject)
			}
			return nil
		},
	}
	c.Flags().IntVarP(&maxN, "max", "n", 10, "maximum commits to show")
	return c
}
