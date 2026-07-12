package app

import (
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider/anthropic"
	"github.com/datamaia/andromeda/internal/provider/ollama"
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
