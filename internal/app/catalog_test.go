package app

import (
	"strings"
	"testing"
)

func TestCatalogResolvesEndpointAndKeyEnv(t *testing.T) {
	info, ok := LookupProvider("groq")
	if !ok {
		t.Fatal("groq missing from catalog")
	}
	if info.BaseURL != "https://api.groq.com/openai/v1" || info.KeyEnv != "GROQ_API_KEY" {
		t.Errorf("groq config = %+v", info)
	}
}

func TestBuildProviderRequiresKeyForHosted(t *testing.T) {
	t.Setenv("GROQ_API_KEY", "")
	if _, err := BuildProvider(ProviderSpec{Name: "groq"}); err == nil {
		t.Error("expected an error when GROQ_API_KEY is unset")
	}
	t.Setenv("GROQ_API_KEY", "test-key")
	if _, err := BuildProvider(ProviderSpec{Name: "groq"}); err != nil {
		t.Errorf("groq with key: %v", err)
	}
}

func TestBuildProviderLocalNeedsNoKey(t *testing.T) {
	for _, id := range []string{"ollama", "vllm"} {
		if _, err := BuildProvider(ProviderSpec{Name: id}); err != nil {
			t.Errorf("%s should build without a key: %v", id, err)
		}
	}
}

func TestBuildProviderUnknownIsError(t *testing.T) {
	if _, err := BuildProvider(ProviderSpec{Name: "nope"}); err == nil || !strings.Contains(err.Error(), "unknown provider") {
		t.Errorf("unknown provider error = %v", err)
	}
}
