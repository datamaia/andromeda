package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	ctx  context.Context // program lifetime; cancelling it tears a run and its approver down
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

// planModeSystem constrains the agent to proposing a plan without touching anything.
const planModeSystem = "You are in PLAN MODE. Analyze the request and propose a concise, numbered " +
	"plan of the steps you would take. You have read-only access: do NOT create, edit, delete, or " +
	"run anything. End by asking the user to switch to agent mode (/agent) to execute."

// respond handles a submitted line according to the active interaction mode: shell runs it as a
// command, plan drives the agent read-only with a planning prompt, agent runs the full loop with
// whatever capabilities the session was granted.
func (s *tuiSession) respond(goal, mode string) string {
	switch mode {
	case "shell":
		return s.runShell(goal)
	case "plan":
		return s.runAgent(goal, planModeSystem, false, false)
	default:
		return s.runAgent(goal, "", s.cfg.allowWrite, s.cfg.allowExec)
	}
}

// runAgent drives the real agent for a goal with explicit capability grants.
func (s *tuiSession) runAgent(goal, system string, allowWrite, allowExec bool) string {
	res, err := app.RunAgent(context.Background(), app.RunAgentOptions{
		WorkspaceRoot: s.wd, Goal: goal, System: system, Model: s.cfg.model, Provider: s.prov,
		AllowWrite: allowWrite, AllowExec: allowExec,
	})
	if err != nil {
		return "error: " + err.Error()
	}
	return res.FinalText
}

// runShell runs the line as a command in the workspace directory (shell mode is the user's own
// command, not the agent's, so it is not gated by the agent permission model).
func (s *tuiSession) runShell(line string) string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.CommandContext(context.Background(), shell, "-c", line)
	cmd.Dir = s.wd
	out, err := cmd.CombinedOutput()
	text := strings.TrimRight(string(out), "\n")
	if err != nil {
		if text != "" {
			return text + "\n[" + err.Error() + "]"
		}
		return "[" + err.Error() + "]"
	}
	if text == "" {
		return "[no output]"
	}
	return text
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

// selectModel records the model the user chose so the agent runs on it (not the provider default).
func (s *tuiSession) selectModel(id string) { s.cfg.model = id }

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

// launchTUI builds the session and hands control to the TUI shell with the provider menu, agent
// runner, interactive sign-in, and API-key entry wired. When onboard is true the session opens in
// first-run mode: a provider (with sign-in/key) and a model must be chosen before chatting.
func launchTUI(ctx context.Context, cfg tuiConfig, onboard bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	sess := &tuiSession{ctx: ctx, wd: wd, cfg: cfg}
	if err := sess.build(); err != nil {
		return err
	}
	m := tui.New(sess.cfg.provider, sess.cfg.model, sess.respond).
		WithProviderMenu(providerChoices(), sess.selectProvider).
		WithModelSelect(sess.selectModel).
		WithActions(sess.sessionActions()).
		WithAgentRunner(sess.startAgentRun).
		WithProviderAuth(sess.startProviderAuth).
		WithProviderKeyEntry(providerKeyEnvFor, setProviderKey)
	if onboard {
		m = m.WithOnboarding()
	}
	return tui.RunModel(ctx, m)
}

func newTUICommand() *cobra.Command {
	cfg := defaultTUIConfig()
	c := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive terminal UI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Onboard (pick provider + model) unless the caller pinned a provider with --provider.
			return launchTUI(cmd.Context(), cfg, !cmd.Flags().Changed("provider"))
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
