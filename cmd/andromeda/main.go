// Command andromeda is the composition root of the Andromeda AI engineering harness.
//
// This is the walking-skeleton entrypoint (EP-01). The full CLI grammar is specified in
// Volume 8 of docs/spec and is built out in later epics; for now the binary starts, reports
// its version, and exits with the codes defined in Volume 0 chapter 03 (ADR-016).
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/buildinfo"
	"github.com/spf13/cobra"
)

// Exit codes (subset of the Volume 0 / ADR-016 scheme used at this stage).
const (
	exitOK      = 0
	exitGeneral = 1
	exitUsage   = 2
)

// usageError is an invocation error that maps to the usage exit code (2). Its message,
// when non-empty, has already been written to stderr by the handler, so run() must not
// print it again (FR-CLI-003 branch 2 prints its own short usage).
type usageError struct{ msg string }

func (e *usageError) Error() string { return e.msg }

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	root := newRootCommand()
	root.SetArgs(args)
	err := root.Execute()
	if err == nil {
		return exitOK
	}
	var ue *usageError
	if errors.As(err, &ue) {
		if ue.msg != "" {
			fmt.Fprintln(os.Stderr, ue.msg)
		}
		return exitUsage
	}
	fmt.Fprintln(os.Stderr, "Error:", err.Error())
	return exitUsage
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "andromeda",
		Short:         "Andromeda — your terminal companion for shipping great software.",
		Long:          "Andromeda is an open-source, local-first, vendor-agnostic AI engineering harness (CLI + TUI).",
		SilenceUsage:  true,
		SilenceErrors: true, // run() owns error reporting so branch-2 usage isn't double-printed
		Args:          cobra.ArbitraryArgs,
		// Bare `andromeda` is the TUI entry point (FR-CLI-003): on an interactive terminal it
		// hands off to the TUI; piped/CI/dumb contexts get short usage on stderr and exit 2,
		// never a full-screen takeover of a non-TTY stream.
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return &usageError{fmt.Sprintf("andromeda: unknown command %q\nRun 'andromeda --help' for usage.", args[0])}
			}
			if !interactive() {
				fmt.Fprint(cmd.ErrOrStderr(), shortUsage(cmd))
				return &usageError{} // already printed; exit 2
			}
			// Bare `andromeda` is first-run onboarding: choose a provider (sign in / paste key) and
			// a model before the chat opens.
			return launchTUI(cmd.Context(), defaultTUIConfig(), true)
		},
	}
	// `andromeda --version` is the single sanctioned flag alias for `andromeda version`
	// (FR-CLI-003), byte-for-byte identical via the shared version line.
	root.Version = versionLine()
	root.SetVersionTemplate("{{.Version}}\n")
	root.CompletionOptions.HiddenDefaultCmd = true
	root.AddCommand(
		newVersionCommand(),
		newDoctorCommand(),
		newRunCommand(),
		newConfigCommand(),
		newGitCommand(),
		newMemoryCommand(),
		newToolCommand(),
		newIndexCommand(),
		newAuthCommand(),
		newWorkflowCommand(),
		newProviderCommand(),
		newModelCommand(),
		newUpdateCommand(),
		newTUICommand(),
		newLogsCommand(),
		newExportCommand(),
		newContextCommand(),
		newTraceCommand(),
	)
	root.AddCommand(newCompletionCommand(root))
	return root
}

func newRunCommand() *cobra.Command {
	var (
		providerName string
		baseURL      string
		apiKeyEnv    string
		model        string
		system       string
		allowWrite   bool
		allowExec    bool
		allowNetwork bool
		maxIter      int
	)
	cmd := &cobra.Command{
		Use:   "run <goal>",
		Short: "Run an agent to accomplish a goal in the current workspace",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			goal := strings.Join(args, " ")
			res, err := app.RunAgent(cmd.Context(), app.RunAgentOptions{
				WorkspaceRoot: wd, Goal: goal, System: system, Model: model,
				Provider: prov, AllowWrite: allowWrite, AllowExec: allowExec, AllowNetwork: allowNetwork, MaxIterations: maxIter,
			})
			out := cmd.OutOrStdout()
			if err != nil {
				_, _ = fmt.Fprintf(out, "run %s (%s)\n", res.State, res.RunID)
				return err
			}
			_, _ = fmt.Fprintln(out, res.FinalText)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
				"\n[run %s · %d iterations · %d tool calls · %d/%d tokens]\n",
				res.State, res.Iterations, res.ToolCalls, res.InputTokens, res.OutputTokens)
			return nil
		},
	}
	cmd.Flags().StringVar(&providerName, "provider", "ollama", "provider: ollama|openai-compatible|anthropic")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "provider base URL (required for openai-compatible)")
	cmd.Flags().StringVar(&apiKeyEnv, "api-key-env", "", "environment variable holding the API key")
	cmd.Flags().StringVar(&model, "model", "llama3", "model identifier")
	cmd.Flags().StringVar(&system, "system", "You are Andromeda, a helpful engineering agent.", "system prompt")
	cmd.Flags().BoolVar(&allowWrite, "allow-write", false, "grant the agent write access within the workspace")
	cmd.Flags().BoolVar(&allowExec, "allow-exec", false, "grant the agent command execution (terminal_run)")
	cmd.Flags().BoolVar(&allowNetwork, "allow-network", false, "grant the agent network access (http_request)")
	cmd.Flags().IntVar(&maxIter, "max-iterations", 0, "iteration budget (0 = default)")
	return cmd
}

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check the environment and foundation (config, databases, events)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			rep, err := app.Doctor(context.Background(), wd)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, c := range rep.Checks {
				mark := "ok  "
				if !c.OK {
					mark = "FAIL"
				}
				_, _ = fmt.Fprintf(out, "[%s] %-13s %s\n", mark, c.Name, c.Detail)
			}
			if !rep.OK() {
				return fmt.Errorf("one or more checks failed")
			}
			_, _ = fmt.Fprintln(out, "doctor: all checks passed")
			return nil
		},
	}
}

// versionLine is the canonical one-line version string, shared by the `version` command
// and the `--version` flag alias so they are byte-for-byte identical (FR-CLI-003).
func versionLine() string {
	i := buildinfo.Get()
	return fmt.Sprintf("andromeda %s (commit %s, built %s, %s/%s)",
		i.Version, i.Commit, i.Date, i.GoOS, i.GoArch)
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build metadata",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), versionLine())
			return err
		},
	}
}

// shortUsage is the compact command list printed to stderr when `andromeda` is invoked
// with no command in a non-interactive context (FR-CLI-003 branch 2). One line per
// command so a mistaken pipe gets an orienting map, not a hung full-screen program.
func shortUsage(root *cobra.Command) string {
	var b strings.Builder
	b.WriteString("andromeda: no interactive terminal detected; not launching the TUI.\n\n")
	b.WriteString("Run 'andromeda' in a terminal for the interactive UI, or use a command:\n")
	for _, c := range root.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		b.WriteString(fmt.Sprintf("  %-11s %s\n", c.Name(), c.Short))
	}
	b.WriteString("\nRun 'andromeda --help' for full help.\n")
	return b.String()
}
