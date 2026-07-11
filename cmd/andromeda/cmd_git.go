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
			fmt.Fprintf(out, "On branch %s\n", st.Branch)
			if st.Clean {
				fmt.Fprintln(out, "nothing to commit, working tree clean")
				return nil
			}
			for _, p := range st.Staged {
				fmt.Fprintf(out, "  staged:    %s\n", p)
			}
			for _, p := range st.Unstaged {
				fmt.Fprintf(out, "  modified:  %s\n", p)
			}
			for _, p := range st.Untracked {
				fmt.Fprintf(out, "  untracked: %s\n", p)
			}
			return nil
		},
	}
}

func newGitLogCommand() *cobra.Command {
	var max int
	c := &cobra.Command{
		Use:   "log",
		Short: "Show recent commits",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo, err := repoRef()
			if err != nil {
				return err
			}
			st, err := git.New("").Log(cmd.Context(), repo, ports.LogSpec{Max: max})
			if err != nil {
				return err
			}
			defer st.Close()
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
				fmt.Fprintf(out, "%s  %s  %s\n", short, ci.Date, ci.Subject)
			}
			return nil
		},
	}
	c.Flags().IntVarP(&max, "max", "n", 10, "maximum commits to show")
	return c
}
