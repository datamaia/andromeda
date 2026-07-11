package main

import (
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/spf13/cobra"
)

func newProviderCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "provider", Short: "Inspect model providers"}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List the provider adapters Andromeda supports",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			rows := []struct{ name, auth, note string }{
				{"ollama", "none", "local (default), http://localhost:11434"},
				{"openai-compatible", "api key", "generic OpenAI Chat Completions surface"},
				{"anthropic", "api key", "Anthropic Messages API"},
			}
			for _, r := range rows {
				fmt.Fprintf(out, "%-18s %-8s %s\n", r.name, r.auth, r.note)
			}
			return nil
		},
	})
	return cmd
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
				fmt.Fprintln(cmd.ErrOrStderr(), "no models reported (the provider may not expose a discovery endpoint)")
				return nil
			}
			for _, m := range models {
				fmt.Fprintln(out, m.ID)
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
