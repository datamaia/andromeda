package main

import (
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/auth"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/spf13/cobra"
)

func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage provider credentials (official mechanisms only)"}
	cmd.AddCommand(newAuthAddCommand(), newAuthListCommand(), newAuthRemoveCommand())
	return cmd
}

func newAuthManager() (*auth.Manager, error) {
	ss, err := app.SecretStore()
	if err != nil {
		return nil, err
	}
	return auth.New(ss), nil
}

func newAuthAddCommand() *cobra.Command {
	var profile, keyEnv string
	c := &cobra.Command{
		Use:   "add <provider>",
		Short: "Store an API key for a provider (key read from an environment variable)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if keyEnv == "" {
				return fmt.Errorf("--key-env is required (the env var holding the API key; the key is never passed on the command line)")
			}
			key := os.Getenv(keyEnv)
			if key == "" {
				return fmt.Errorf("environment variable %s is empty", keyEnv)
			}
			m, err := newAuthManager()
			if err != nil {
				return err
			}
			if err := m.StoreAPIKey(cmd.Context(), args[0], profile, key); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "stored credential for %s (profile %s)\n", args[0], orDefault(profile))
			return nil
		},
	}
	c.Flags().StringVar(&profile, "profile", "default", "credential profile")
	c.Flags().StringVar(&keyEnv, "key-env", "", "environment variable holding the API key")
	return c
}

func newAuthListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured credential profiles (no secrets shown)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := newAuthManager()
			if err != nil {
				return err
			}
			profiles, err := m.ListProfiles(cmd.Context())
			if err != nil {
				return err
			}
			if len(profiles) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no credentials stored")
				return nil
			}
			for _, p := range profiles {
				fmt.Fprintf(cmd.OutOrStdout(), "%s (profile %s)\n", p.Provider, p.Name)
			}
			return nil
		},
	}
}

func newAuthRemoveCommand() *cobra.Command {
	var profile string
	c := &cobra.Command{
		Use:   "remove <provider>",
		Short: "Remove a stored credential",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := newAuthManager()
			if err != nil {
				return err
			}
			if err := m.Revoke(cmd.Context(), ports.AuthenticationHandle{Provider: args[0], Profile: profile}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed credential for %s\n", args[0])
			return nil
		},
	}
	c.Flags().StringVar(&profile, "profile", "default", "credential profile")
	return c
}

func orDefault(s string) string {
	if s == "" {
		return "default"
	}
	return s
}
