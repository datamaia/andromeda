package tui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func onboardChoices() []ProviderChoice {
	return []ProviderChoice{
		{"ollama", "Ollama", "no key", ""},
		{"groq", "Groq", "GROQ_API_KEY", ""},
		{"openai-chatgpt", "ChatGPT", "no key", ""},
	}
}

// onboardModel wires a first-run model with provider selection, model discovery, OAuth for ChatGPT,
// and key entry for groq.
func onboardModel(t *testing.T, onSelect ProviderSelectFunc, setKey ProviderSetKeyFunc) tea.Model {
	t.Helper()
	authFn := func(id string) <-chan AuthEvent {
		if id != "openai-chatgpt" {
			return nil
		}
		ch := make(chan AuthEvent, 2)
		go func() {
			ch <- AuthEvent{URL: "https://auth.example/login"}
			ch <- AuthEvent{Done: true}
			close(ch)
		}()
		return ch
	}
	keyEnvFn := func(id string) string {
		if id == "groq" {
			return "GROQ_API_KEY"
		}
		return ""
	}
	return tea.Model(New("ollama", "llama3", nil).
		WithProviderMenu(onboardChoices(), onSelect).
		WithActions(Actions{Models: func(context.Context) []string { return []string{"m1", "m2"} }}).
		WithProviderAuth(authFn).
		WithProviderKeyEntry(keyEnvFn, setKey).
		WithOnboarding())
}

// Onboarding opens on the provider picker.
func TestOnboardingOpensProviderPicker(t *testing.T) {
	m := onboardModel(t, func(id string) (string, error) { return "x", nil }, nil)
	got := m.(Model)
	if !got.onboarding || !got.pickerOpen || got.pickerKind != "provider" {
		t.Fatalf("onboarding should open the provider picker: onboarding=%v open=%v kind=%q",
			got.onboarding, got.pickerOpen, got.pickerKind)
	}
}

// Picking a local provider advances to the model picker, and picking a model finishes onboarding.
func TestOnboardingProviderThenModel(t *testing.T) {
	m := onboardModel(t, func(id string) (string, error) { return "model-for-" + id, nil }, nil)
	// cursor starts on ollama (index 0); enter selects it
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if !got.pickerOpen || got.pickerKind != "model" {
		t.Fatalf("after choosing a provider onboarding should show the model picker, kind=%q", got.pickerKind)
	}
	if got.provider != "ollama" {
		t.Errorf("provider = %q, want ollama", got.provider)
	}
	// pick the first discovered model
	m, _ = m.Update(key(tea.KeyEnter))
	got = m.(Model)
	if got.onboarding {
		t.Error("choosing a model should finish onboarding")
	}
	if got.pickerOpen {
		t.Error("picker should be closed after onboarding")
	}
	if got.model != "m1" {
		t.Errorf("model = %q, want m1", got.model)
	}
}

// Esc on the provider picker during onboarding quits (a provider is required).
func TestOnboardingEscOnProviderQuits(t *testing.T) {
	m := onboardModel(t, func(id string) (string, error) { return "x", nil }, nil)
	_, cmd := m.Update(key(tea.KeyEscape))
	if cmd == nil {
		t.Fatal("esc on the onboarding provider picker should quit")
	}
}

// Esc on the model picker during onboarding steps back to the provider picker.
func TestOnboardingEscOnModelGoesBack(t *testing.T) {
	m := onboardModel(t, func(id string) (string, error) { return "model-for-" + id, nil }, nil)
	m, _ = m.Update(key(tea.KeyEnter)) // provider → model picker
	m, _ = m.Update(key(tea.KeyEscape))
	got := m.(Model)
	if got.pickerKind != "provider" {
		t.Errorf("esc on the model picker should return to the provider picker, kind=%q", got.pickerKind)
	}
	if got.quitting {
		t.Error("esc on the model picker must not quit")
	}
}

// Selecting a provider whose key is missing opens the key-entry prompt; pasting activates it.
func TestOnboardingKeyEntry(t *testing.T) {
	var storedID, storedKey string
	m := onboardModel(t,
		func(id string) (string, error) { return "model-for-" + id, nil },
		func(id, key string) { storedID, storedKey = id, key })
	// move to groq (index 1) and select
	m, _ = m.Update(key(tea.KeyDown))
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if !got.keyEntry || got.keyEnvName != "GROQ_API_KEY" {
		t.Fatalf("selecting groq should open the key prompt for GROQ_API_KEY, keyEntry=%v env=%q",
			got.keyEntry, got.keyEnvName)
	}
	// type a key and submit
	m = typeString(m, "sk-abc123")
	m, _ = m.Update(key(tea.KeyEnter))
	got = m.(Model)
	if storedID != "groq" || storedKey != "sk-abc123" {
		t.Errorf("setProviderKey got (%q,%q), want (groq, sk-abc123)", storedID, storedKey)
	}
	if got.provider != "groq" {
		t.Errorf("provider = %q, want groq after key entry", got.provider)
	}
	if !got.pickerOpen || got.pickerKind != "model" {
		t.Error("after key entry onboarding should advance to the model picker")
	}
}

// Selecting the ChatGPT provider runs the browser sign-in: the URL is surfaced, then on completion
// the provider activates and onboarding advances to the model picker.
func TestOnboardingOAuthFlow(t *testing.T) {
	var activated string
	m := onboardModel(t, func(id string) (string, error) { activated = id; return "gpt-5.1-codex", nil }, nil)
	// move to openai-chatgpt (index 2) and select
	m, _ = m.Update(key(tea.KeyDown))
	m, _ = m.Update(key(tea.KeyDown))
	m, cmd := m.Update(key(tea.KeyEnter))
	if !m.(Model).authing {
		t.Fatal("selecting ChatGPT should start the sign-in flow")
	}
	// first event: the browser URL
	m, cmd = m.Update(cmd())
	if !strings.Contains(m.(Model).View().Content, "auth.example/login") {
		t.Error("the sign-in URL should be shown to the user")
	}
	// second event: completion → activate + advance
	m, _ = m.Update(cmd())
	got := m.(Model)
	if activated != "openai-chatgpt" || got.provider != "openai-chatgpt" {
		t.Errorf("after sign-in the provider should be openai-chatgpt, got %q", got.provider)
	}
	if got.authing {
		t.Error("authing should be cleared after completion")
	}
	if !got.pickerOpen || got.pickerKind != "model" {
		t.Error("after sign-in onboarding should advance to the model picker")
	}
}
