package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/auth"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/spf13/cobra"
)

func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage provider credentials (official mechanisms only)"}
	cmd.AddCommand(newAuthAddCommand(), newAuthLoginCommand(), newAuthListCommand(), newAuthRemoveCommand())
	return cmd
}

// newAuthLoginCommand runs a browser OAuth login (currently: openai-chatgpt) and stores the
// resulting token bundle. The subscription grant never becomes an API key — it is an OAuth
// session used only against the provider's own backend.
func newAuthLoginCommand() *cobra.Command {
	var profile string
	c := &cobra.Command{
		Use:   "login <provider>",
		Short: "Sign in through the browser (OAuth). Supported: openai-chatgpt (your ChatGPT account)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			var flow auth.BrowserFlowConfig
			switch provider {
			case "openai-chatgpt", "chatgpt":
				provider = auth.OpenAIChatGPTProvider
				flow = auth.OpenAIChatGPTFlow()
			default:
				return fmt.Errorf("browser login is not available for %q (supported: openai-chatgpt)", provider)
			}
			m, err := newAuthManager()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(out, "Opening your browser to sign in. If it doesn't open, visit this URL:")
			openAndPrint := func(u string) error {
				_, _ = fmt.Fprintln(out, "  "+u)
				return openBrowser(u)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()
			tok, err := auth.RunBrowserFlow(ctx, flow, openAndPrint)
			if err != nil {
				return err
			}
			if err := m.StoreOAuthToken(ctx, provider, profile, tok); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(out, "Signed in to %s (profile %s).\n", provider, orDefault(profile))
			return nil
		},
	}
	c.Flags().StringVar(&profile, "profile", "default", "credential profile")
	return c
}

// openBrowser launches the system browser at u (best-effort; the caller also prints the URL).
func openBrowser(u string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name, args = "open", []string{u}
	case "windows":
		name, args = "rundll32", []string{"url.dll,FileProtocolHandler", u}
	default:
		name, args = "xdg-open", []string{u}
	}
	return exec.Command(name, args...).Start()
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
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "stored credential for %s (profile %s)\n", args[0], orDefault(profile))
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
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no credentials stored")
				return nil
			}
			for _, p := range profiles {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s (profile %s)\n", p.Provider, p.Name)
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
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "removed credential for %s\n", args[0])
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
