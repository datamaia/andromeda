package main

import (
	"context"
	"os"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/tui"
	"github.com/spf13/cobra"
)

// tuiConfig holds the provider wiring for a TUI session. It is shared by the explicit
// `tui` subcommand and the bare `andromeda` invocation (FR-CLI-003) so both launch the
// same shell; the defaults match the `run` command (local Ollama, no granted capabilities).
type tuiConfig struct {
	provider, baseURL, apiKeyEnv, model string
	allowWrite, allowExec               bool
}

func defaultTUIConfig() tuiConfig {
	return tuiConfig{provider: "ollama", model: "llama3"}
}

// launchTUI builds the provider and hands control to the TUI shell. The responder drives
// the real agent for each submitted goal.
func launchTUI(ctx context.Context, cfg tuiConfig) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	apiKey := ""
	if cfg.apiKeyEnv != "" {
		apiKey = os.Getenv(cfg.apiKeyEnv)
	}
	prov, err := app.BuildProvider(app.ProviderSpec{Name: cfg.provider, BaseURL: cfg.baseURL, APIKey: apiKey})
	if err != nil {
		return err
	}
	respond := func(goal string) string {
		res, err := app.RunAgent(ctx, app.RunAgentOptions{
			WorkspaceRoot: wd, Goal: goal, Model: cfg.model, Provider: prov,
			AllowWrite: cfg.allowWrite, AllowExec: cfg.allowExec,
		})
		if err != nil {
			return "error: " + err.Error()
		}
		return res.FinalText
	}
	return tui.Run(ctx, cfg.provider, cfg.model, respond)
}

func newTUICommand() *cobra.Command {
	cfg := defaultTUIConfig()
	c := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive terminal UI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return launchTUI(cmd.Context(), cfg)
		},
	}
	c.Flags().StringVar(&cfg.provider, "provider", cfg.provider, "provider name")
	c.Flags().StringVar(&cfg.baseURL, "base-url", "", "provider base URL")
	c.Flags().StringVar(&cfg.apiKeyEnv, "api-key-env", "", "environment variable holding the API key")
	c.Flags().StringVar(&cfg.model, "model", cfg.model, "model identifier")
	c.Flags().BoolVar(&cfg.allowWrite, "allow-write", false, "grant the agent write access")
	c.Flags().BoolVar(&cfg.allowExec, "allow-exec", false, "grant the agent command execution")
	return c
}
