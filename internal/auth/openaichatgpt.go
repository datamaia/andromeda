package auth

// OpenAIChatGPTClientID is OpenAI's public Codex OAuth client, used by the Codex CLI and
// compatible "log in with ChatGPT" logins. It is a public client and carries no secret.
const OpenAIChatGPTClientID = "app_EMoamEEZ73f0CkXaXp7hrann"

// OpenAIChatGPTProvider is the catalog/credential ID for a ChatGPT-subscription login.
const OpenAIChatGPTProvider = "openai-chatgpt"

// OpenAIChatGPTFlow returns the browser OAuth flow for signing in with a ChatGPT account (the
// Codex subscription flow). Endpoints, client, scopes, and the extra authorization parameters
// mirror OpenAI's Codex login so the resulting tokens are accepted by the ChatGPT backend.
func OpenAIChatGPTFlow() BrowserFlowConfig {
	return BrowserFlowConfig{
		AuthorizeURL: "https://auth.openai.com/oauth/authorize",
		TokenURL:     "https://auth.openai.com/oauth/token",
		ClientID:     OpenAIChatGPTClientID,
		RedirectPath: "/auth/callback",
		RedirectPort: 1455,
		Scopes:       []string{"openid", "profile", "email", "offline_access"},
		ExtraAuth: map[string]string{
			"id_token_add_organizations": "true",
			"codex_cli_simplified_flow":  "true",
			"originator":                 "codex_cli_rs",
		},
	}
}
