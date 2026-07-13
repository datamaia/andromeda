package main

import (
	"context"
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/ports"
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

// tuiSession owns the live provider for a TUI session so the in-app provider menu can rebuild it
// mid-session. The agent responder and the menu's select callback both read/mutate it.
type tuiSession struct {
	wd   string
	cfg  tuiConfig
	prov ports.ProviderPort
}

func (s *tuiSession) build() error {
	apiKey := ""
	if s.cfg.apiKeyEnv != "" {
		apiKey = os.Getenv(s.cfg.apiKeyEnv)
	}
	prov, err := app.BuildProvider(app.ProviderSpec{Name: s.cfg.provider, BaseURL: s.cfg.baseURL, APIKey: apiKey})
	if err != nil {
		return err
	}
	s.prov = prov
	return nil
}

// respond drives the real agent for each submitted goal, using whatever provider is current.
func (s *tuiSession) respond(goal string) string {
	res, err := app.RunAgent(context.Background(), app.RunAgentOptions{
		WorkspaceRoot: s.wd, Goal: goal, Model: s.cfg.model, Provider: s.prov,
		AllowWrite: s.cfg.allowWrite, AllowExec: s.cfg.allowExec,
	})
	if err != nil {
		return "error: " + err.Error()
	}
	return res.FinalText
}

// selectProvider rebuilds the live provider from a catalog ID (menu selection), adopting the
// catalog's default model, and returns the model now in use. It reads the key from the catalog's
// environment variable; an unset required key surfaces as an error the menu shows.
func (s *tuiSession) selectProvider(id string) (string, error) {
	info, ok := app.LookupProvider(id)
	if !ok {
		return "", fmt.Errorf("unknown provider %q", id)
	}
	prov, err := app.BuildProvider(app.ProviderSpec{Name: id})
	if err != nil {
		return "", err
	}
	s.prov = prov
	s.cfg.provider = id
	if info.DefaultModel != "" {
		s.cfg.model = info.DefaultModel
	}
	return s.cfg.model, nil
}

// providerChoices adapts the app catalog into the TUI's menu entries.
func providerChoices() []tui.ProviderChoice {
	infos := app.Providers()
	choices := make([]tui.ProviderChoice, 0, len(infos))
	for _, p := range infos {
		auth := "no key"
		if p.KeyEnv != "" {
			auth = p.KeyEnv
		}
		choices = append(choices, tui.ProviderChoice{ID: p.ID, Display: p.Display, Auth: auth, Note: p.Note})
	}
	return choices
}

// launchTUI builds the session and hands control to the TUI shell with the provider menu wired.
func launchTUI(ctx context.Context, cfg tuiConfig) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	sess := &tuiSession{wd: wd, cfg: cfg}
	if err := sess.build(); err != nil {
		return err
	}
	m := tui.New(sess.cfg.provider, sess.cfg.model, sess.respond).
		WithProviderMenu(providerChoices(), sess.selectProvider).
		WithActions(sess.sessionActions())
	return tui.RunModel(ctx, m)
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
