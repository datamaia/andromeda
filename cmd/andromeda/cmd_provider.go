package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/spf13/cobra"
)

func newProviderCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "provider", Short: "Inspect model providers"}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List the model providers Andromeda supports",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			for _, p := range app.Providers() {
				auth := "none"
				if p.KeyEnv != "" {
					auth = p.KeyEnv
					if !p.KeyRequired {
						auth += " (optional)"
					}
				}
				_, _ = fmt.Fprintf(out, "%-12s %-24s %-24s %s\n", p.ID, p.Display, auth, p.Note)
			}
			return nil
		},
	})
	cmd.AddCommand(newProviderCheckCommand())
	return cmd
}

// newProviderCheckCommand probes provider connectivity using each provider's catalog key
// environment variable — a one-shot validation of the keys the user has exported (e.g. from .env).
// With no arguments it checks every local provider and every hosted provider whose key is set,
// skipping those without a key and the OAuth ChatGPT provider (use `auth login` for that).
func newProviderCheckCommand() *cobra.Command {
	var timeout time.Duration
	c := &cobra.Command{
		Use:   "check [provider...]",
		Short: "Validate provider connectivity using each provider's API key from the environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			targets := providerCheckTargets(args)
			if len(targets) == 0 {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "no providers to check (set a key like GROQ_API_KEY, or name a provider)")
				return nil
			}
			anyFail := false
			for _, info := range targets {
				status, detail := checkProvider(cmd.Context(), info, timeout)
				if status == "FAILED" {
					anyFail = true
				}
				_, _ = fmt.Fprintf(out, "%-12s %-7s %s\n", info.ID, status, detail)
			}
			if anyFail {
				return &usageError{} // non-zero exit; results already written to stdout
			}
			return nil
		},
	}
	c.Flags().DurationVar(&timeout, "timeout", 15*time.Second, "per-provider probe timeout")
	return c
}

// providerCheckTargets resolves the providers to probe. Named providers are looked up as given;
// with no names, every local provider and every hosted provider whose key is present is included.
func providerCheckTargets(names []string) []app.ProviderInfo {
	if len(names) > 0 {
		var out []app.ProviderInfo
		for _, n := range names {
			if info, ok := app.LookupProvider(n); ok {
				out = append(out, info)
			}
		}
		return out
	}
	var out []app.ProviderInfo
	for _, info := range app.Providers() {
		switch {
		case info.Kind == app.KindOpenAIChatGPT:
			continue // OAuth, not a key — validated via `auth login openai-chatgpt`
		case info.Local:
			out = append(out, info)
		case info.KeyEnv != "" && os.Getenv(info.KeyEnv) != "":
			out = append(out, info)
		}
	}
	return out
}

// checkProvider builds a provider from its catalog key and probes it with a model discovery call,
// returning a status word (OK / SKIP / FAILED) and a human-readable detail.
func checkProvider(ctx context.Context, info app.ProviderInfo, timeout time.Duration) (string, string) {
	if !info.Local && info.KeyEnv != "" && os.Getenv(info.KeyEnv) == "" {
		return "SKIP", "no key in " + info.KeyEnv
	}
	prov, err := app.BuildProvider(app.ProviderSpec{Name: info.ID})
	if err != nil {
		return "FAILED", err.Error()
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	models, err := prov.DiscoverModels(ctx)
	if err != nil {
		return "FAILED", err.Error()
	}
	if len(models) == 0 {
		return "OK", "reachable (no model list exposed)"
	}
	return "OK", fmt.Sprintf("%d models (e.g. %s)", len(models), models[0].ID)
}

func newModelCommand() *cobra.Command {
	var providerName, baseURL, apiKeyEnv string
	cmd := &cobra.Command{Use: "model", Short: "Discover models from a provider"}
	list := &cobra.Command{
		Use:   "list",
		Short: "List models the configured provider exposes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			apiKey := ""
			if apiKeyEnv != "" {
				apiKey = os.Getenv(apiKeyEnv)
			}
			prov, err := app.BuildProvider(app.ProviderSpec{Name: providerName, BaseURL: baseURL, APIKey: apiKey})
			if err != nil {
				return err
			}
			models, err := prov.DiscoverModels(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(models) == 0 {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "no models reported (the provider may not expose a discovery endpoint)")
				return nil
			}
			for _, m := range models {
				_, _ = fmt.Fprintln(out, m.ID)
			}
			return nil
		},
	}
	list.Flags().StringVar(&providerName, "provider", "ollama", "provider name")
	list.Flags().StringVar(&baseURL, "base-url", "", "provider base URL")
	list.Flags().StringVar(&apiKeyEnv, "api-key-env", "", "environment variable holding the API key")
	cmd.AddCommand(list)
	return cmd
}
