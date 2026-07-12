package main

import (
	"context"
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/app"
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
				_, _ = fmt.Fprintf(out, "%2d. %s\n", i+1, name)
			}
			return nil
		},
	}
}

func newWorkflowRunCommand() *cobra.Command {
	var autoApprove bool
	var goal, providerName, baseURL, apiKeyEnv, model string
	c := &cobra.Command{
		Use:   "run sdd",
		Short: "Run the SDD workflow; with --goal, each stage is driven by the agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "sdd" {
				return fmt.Errorf("unknown workflow %q (only 'sdd' is built in)", args[0])
			}
			out := cmd.OutOrStdout()

			// With a goal + provider, drive every stage through the agent.
			if goal != "" {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				apiKey := ""
				if apiKeyEnv != "" {
					apiKey = os.Getenv(apiKeyEnv)
				}
				prov, err := app.BuildProvider(app.ProviderSpec{Name: providerName, BaseURL: baseURL, APIKey: apiKey})
				if err != nil {
					return err
				}
				rs, err := app.RunSDD(cmd.Context(), app.SDDOptions{
					WorkspaceRoot: wd, Objective: goal, Provider: prov, Model: model, AutoApprove: autoApprove,
					OnStage: func(stage, _ string) { _, _ = fmt.Fprintf(out, "  ▸ %s\n", stage) },
				})
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\n[workflow sdd · %d stages · %s]\n", len(rs.History), rs.State)
				return nil
			}

			// Without a goal, run the pipeline shell (stage names only).
			def := workflow.SDDDefinition(func(_ context.Context, stage string, _ *workflow.RunState) (workflow.StageResult, error) {
				_, _ = fmt.Fprintf(out, "  ▸ %s\n", stage)
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
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\n[workflow %s · %d stages · %s]\n", rs.Workflow, len(rs.History), rs.State)
			if rs.State == workflow.StateAwaitingApproval {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "halted at gate stage %q — re-run with --auto-approve to proceed non-interactively\n",
					workflow.SDDStageNames()[rs.StageIdx])
			}
			return nil
		},
	}
	c.Flags().BoolVar(&autoApprove, "auto-approve", false, "auto-approve gate stages (non-interactive)")
	c.Flags().StringVar(&goal, "goal", "", "objective to drive each stage through the agent")
	c.Flags().StringVar(&providerName, "provider", "ollama", "provider name")
	c.Flags().StringVar(&baseURL, "base-url", "", "provider base URL")
	c.Flags().StringVar(&apiKeyEnv, "api-key-env", "", "environment variable holding the API key")
	c.Flags().StringVar(&model, "model", "llama3", "model identifier")
	return c
}
