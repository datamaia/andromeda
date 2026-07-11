package main

import (
	"context"
	"fmt"

	"github.com/datamaia/andromeda/internal/workflow"
	"github.com/spf13/cobra"
)

func newWorkflowCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "workflow", Short: "Specification-driven development workflows"}
	cmd.AddCommand(newWorkflowListCommand(), newWorkflowRunCommand())
	return cmd
}

func newWorkflowListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the stages of the built-in SDD workflow",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			for i, name := range workflow.SDDStageNames() {
				fmt.Fprintf(out, "%2d. %s\n", i+1, name)
			}
			return nil
		},
	}
}

func newWorkflowRunCommand() *cobra.Command {
	var autoApprove bool
	c := &cobra.Command{
		Use:   "run sdd",
		Short: "Run the SDD workflow shell (stages are stubbed pending agent wiring)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "sdd" {
				return fmt.Errorf("unknown workflow %q (only 'sdd' is built in)", args[0])
			}
			out := cmd.OutOrStdout()
			def := workflow.SDDDefinition(func(_ context.Context, stage string, _ *workflow.RunState) (workflow.StageResult, error) {
				fmt.Fprintf(out, "  ▸ %s\n", stage)
				return workflow.StageResult{Summary: stage}, nil
			})
			opts := []workflow.Option{}
			if autoApprove {
				opts = append(opts, workflow.WithAutoApproveGates())
			}
			rs, err := workflow.New(opts...).Execute(cmd.Context(), def, nil, 0)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "\n[workflow %s · %d stages · %s]\n", rs.Workflow, len(rs.History), rs.State)
			if rs.State == workflow.StateAwaitingApproval {
				fmt.Fprintf(cmd.ErrOrStderr(), "halted at gate stage %q — re-run with --auto-approve to proceed non-interactively\n",
					workflow.SDDStageNames()[rs.StageIdx])
			}
			return nil
		},
	}
	c.Flags().BoolVar(&autoApprove, "auto-approve", false, "auto-approve gate stages (non-interactive)")
	return c
}
