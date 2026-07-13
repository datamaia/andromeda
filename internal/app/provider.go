package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/datamaia/andromeda/internal/auth"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider/anthropic"
	"github.com/datamaia/andromeda/internal/provider/ollama"
	"github.com/datamaia/andromeda/internal/provider/openaichatgpt"
	"github.com/datamaia/andromeda/internal/provider/openaicompat"
)

// ProviderSpec selects and configures a provider adapter for the CLI. Name may be a catalog ID
// (see catalog.go — ollama, openai, anthropic, gemini, groq, cerebras, openrouter, xai, …) or one
// of the generic aliases "openai-compatible"/"openai". BaseURL and APIKey override the catalog
// defaults when set; otherwise the catalog's endpoint and key environment variable are used.
type ProviderSpec struct {
	Name    string
	BaseURL string
	APIKey  string
}

// BuildProvider constructs a ProviderPort adapter from a spec. A catalog ID resolves its endpoint
// and reads its key from the catalog's environment variable unless one is supplied explicitly;
// local providers (Ollama, vLLM) need no key. Unknown names are an error, not a silent default.
func BuildProvider(spec ProviderSpec) (ports.ProviderPort, error) {
	if info, ok := LookupProvider(spec.Name); ok {
		return buildFromCatalog(info, spec)
	}
	switch spec.Name {
	case "":
		return ollama.New(ollama.Config{BaseURL: spec.BaseURL}), nil
	case "openai-compatible":
		if spec.BaseURL == "" {
			return nil, fmt.Errorf("openai-compatible provider requires a base URL")
		}
		return openaicompat.New(openaicompat.Config{BaseURL: spec.BaseURL, APIKey: spec.APIKey}), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", spec.Name)
	}
}

// buildFromCatalog resolves a catalog entry into a concrete adapter, filling in the endpoint and
// reading the API key from the entry's environment variable when not overridden.
func buildFromCatalog(info ProviderInfo, spec ProviderSpec) (ports.ProviderPort, error) {
	if info.Kind == KindOpenAIChatGPT {
		return openaichatgpt.New(openaichatgpt.Config{Token: chatGPTTokenSource()}), nil
	}
	baseURL := spec.BaseURL
	if baseURL == "" {
		baseURL = info.BaseURL
	}
	apiKey := spec.APIKey
	if apiKey == "" && info.KeyEnv != "" {
		apiKey = os.Getenv(info.KeyEnv)
	}
	if info.KeyRequired && apiKey == "" {
		return nil, fmt.Errorf("provider %q requires an API key in %s", info.ID, info.KeyEnv)
	}
	switch info.Kind {
	case KindOllama:
		return ollama.New(ollama.Config{BaseURL: baseURL}), nil
	case KindAnthropic:
		return anthropic.New(anthropic.Config{BaseURL: baseURL, APIKey: apiKey}), nil
	case KindOpenAICompat:
		if baseURL == "" {
			return nil, fmt.Errorf("provider %q requires a base URL", info.ID)
		}
		return openaicompat.New(openaicompat.Config{BaseURL: baseURL, APIKey: apiKey}), nil
	default:
		return nil, fmt.Errorf("unknown provider kind for %q", info.ID)
	}
}

// chatGPTTokenSource returns a token accessor for the ChatGPT backend that loads the stored OAuth
// session and refreshes it when it is within five minutes of expiry, restoring the account id the
// refresh response may omit. A missing session yields an actionable "sign in" error.
func chatGPTTokenSource() openaichatgpt.TokenSource {
	return func(ctx context.Context) (string, string, error) {
		ss, err := SecretStore()
		if err != nil {
			return "", "", err
		}
		mgr := auth.New(ss)
		tok, err := mgr.LoadOAuthToken(ctx, auth.OpenAIChatGPTProvider, "default")
		if err != nil {
			return "", "", fmt.Errorf("not signed in to ChatGPT — run: andromeda auth login openai-chatgpt")
		}
		if tok.Expired(5*time.Minute) && tok.RefreshToken != "" {
			if refreshed, rerr := auth.RefreshBrowserToken(ctx, auth.OpenAIChatGPTFlow(), tok.RefreshToken); rerr == nil {
				if refreshed.AccountID == "" {
					refreshed.AccountID = tok.AccountID
				}
				_ = mgr.StoreOAuthToken(ctx, auth.OpenAIChatGPTProvider, "default", refreshed)
				tok = refreshed
			}
		}
		return tok.AccessToken, tok.AccountID, nil
	}
}
