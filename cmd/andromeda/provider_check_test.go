package main

import (
	"testing"

	"github.com/datamaia/andromeda/internal/app"
)

func ids(infos []app.ProviderInfo) []string {
	out := make([]string, len(infos))
	for i, in := range infos {
		out[i] = in.ID
	}
	return out
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

// Named targets resolve by catalog ID; unknown names are dropped.
func TestProviderCheckTargetsNamed(t *testing.T) {
	got := ids(providerCheckTargets([]string{"groq", "nope", "gemini"}))
	if len(got) != 2 || got[0] != "groq" || got[1] != "gemini" {
		t.Fatalf("named targets = %v, want [groq gemini]", got)
	}
}

// With no names and no hosted keys set, only local providers are probed and the OAuth ChatGPT
// provider is never swept.
func TestProviderCheckTargetsDefault(t *testing.T) {
	for _, env := range []string{
		"GROQ_API_KEY", "CEREBRAS_API_KEY", "OPENROUTER_API_KEY", "GEMINI_API_KEY",
		"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "XAI_API_KEY", "HF_TOKEN", "VLLM_API_KEY",
	} {
		t.Setenv(env, "")
	}
	got := ids(providerCheckTargets(nil))
	if !contains(got, "ollama") {
		t.Errorf("default sweep should include local ollama, got %v", got)
	}
	if contains(got, "openai-chatgpt") {
		t.Error("default sweep must not include the OAuth ChatGPT provider")
	}
	if contains(got, "groq") {
		t.Error("hosted providers without a key must be excluded from the default sweep")
	}
}

// A hosted provider whose key is set is included in the default sweep.
func TestProviderCheckTargetsWithKey(t *testing.T) {
	t.Setenv("GROQ_API_KEY", "sk-test-not-real")
	got := ids(providerCheckTargets(nil))
	if !contains(got, "groq") {
		t.Errorf("groq should be swept when its key is set, got %v", got)
	}
}
