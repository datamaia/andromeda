package app

import (
	"fmt"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider/anthropic"
	"github.com/datamaia/andromeda/internal/provider/ollama"
	"github.com/datamaia/andromeda/internal/provider/openaicompat"
)

// ProviderSpec selects and configures a provider adapter for the CLI.
type ProviderSpec struct {
	Name    string // "ollama" | "openai-compatible" | "anthropic"
	BaseURL string
	APIKey  string
}

// BuildProvider constructs a ProviderPort adapter from a spec. Local Ollama needs no key; the
// cloud adapters require one. Unknown names are an error rather than a silent default.
func BuildProvider(spec ProviderSpec) (ports.ProviderPort, error) {
	switch spec.Name {
	case "", "ollama":
		return ollama.New(ollama.Config{BaseURL: spec.BaseURL}), nil
	case "openai-compatible", "openai":
		if spec.BaseURL == "" {
			return nil, fmt.Errorf("openai-compatible provider requires a base URL")
		}
		return openaicompat.New(openaicompat.Config{BaseURL: spec.BaseURL, APIKey: spec.APIKey}), nil
	case "anthropic":
		if spec.APIKey == "" {
			return nil, fmt.Errorf("anthropic provider requires an API key")
		}
		return anthropic.New(anthropic.Config{BaseURL: spec.BaseURL, APIKey: spec.APIKey}), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", spec.Name)
	}
}
