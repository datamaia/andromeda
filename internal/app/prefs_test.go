package app

import "testing"

func TestPrefsRoundtrip(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	// A missing file reads as empty prefs, not an error (first-run).
	if p, err := LoadPrefs(); err != nil || p.Provider != "" || p.Model != "" {
		t.Fatalf("empty load = %+v, err=%v; want zero Prefs", p, err)
	}

	if err := SavePrefs(Prefs{Provider: "openai-chatgpt", Model: "gpt-5.5"}); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPrefs()
	if err != nil {
		t.Fatal(err)
	}
	if p.Provider != "openai-chatgpt" || p.Model != "gpt-5.5" {
		t.Fatalf("roundtrip = %+v, want {openai-chatgpt gpt-5.5}", p)
	}

	// Save overwrites in place.
	if err := SavePrefs(Prefs{Provider: "ollama", Model: "llama3"}); err != nil {
		t.Fatal(err)
	}
	if p, _ := LoadPrefs(); p.Provider != "ollama" {
		t.Fatalf("overwrite failed: %+v", p)
	}
}
