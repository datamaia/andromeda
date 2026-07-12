package main

import (
	"context"
	"os"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/tui"
	"github.com/spf13/cobra"
)

func newTUICommand() *cobra.Command {
	var providerName, baseURL, apiKeyEnv, model string
	var allowWrite, allowExec bool
	c := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive terminal UI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
			// The TUI responder drives the real agent for each submitted goal.
			respond := func(goal string) string {
				res, err := app.RunAgent(context.Background(), app.RunAgentOptions{
					WorkspaceRoot: wd, Goal: goal, Model: model, Provider: prov,
					AllowWrite: allowWrite, AllowExec: allowExec,
				})
				if err != nil {
					return "error: " + err.Error()
				}
				return res.FinalText
			}
			return tui.Run(cmd.Context(), providerName, model, respond)
		},
	}
	c.Flags().StringVar(&providerName, "provider", "ollama", "provider name")
	c.Flags().StringVar(&baseURL, "base-url", "", "provider base URL")
	c.Flags().StringVar(&apiKeyEnv, "api-key-env", "", "environment variable holding the API key")
	c.Flags().StringVar(&model, "model", "llama3", "model identifier")
	c.Flags().BoolVar(&allowWrite, "allow-write", false, "grant the agent write access")
	c.Flags().BoolVar(&allowExec, "allow-exec", false, "grant the agent command execution")
	return c
}
