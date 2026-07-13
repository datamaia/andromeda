package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func menuModel(onSelect ProviderSelectFunc) tea.Model {
	choices := []ProviderChoice{
		{"ollama", "Ollama", "no key", ""},
		{"groq", "Groq", "GROQ_API_KEY", ""},
		{"openai", "OpenAI", "OPENAI_API_KEY", ""},
	}
	return tea.Model(New("ollama", "llama3", nil).WithProviderMenu(choices, onSelect))
}

func ctrlP() tea.KeyPressMsg { return tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl} }

func TestMenuOpenNavigateSelect(t *testing.T) {
	var picked string
	m := menuModel(func(id string) (string, error) { picked = id; return "llama-3.3-70b-versatile", nil })
	m, _ = m.Update(ctrlP())
	if !m.(Model).pickerOpen {
		t.Fatal("ctrl+p should open the menu")
	}
	// cursor starts on the current provider (ollama, index 0); move down to groq (index 1)
	m, _ = m.Update(key(tea.KeyDown))
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if got.pickerOpen {
		t.Error("menu should close after selection")
	}
	if picked != "groq" {
		t.Errorf("selected %q, want groq", picked)
	}
	if got.provider != "groq" || got.model != "llama-3.3-70b-versatile" {
		t.Errorf("after select provider=%q model=%q", got.provider, got.model)
	}
}

func TestMenuEscGoesBackUnchanged(t *testing.T) {
	m := menuModel(func(id string) (string, error) { return "x", nil })
	m, _ = m.Update(ctrlP())
	m, _ = m.Update(key(tea.KeyDown))
	m, _ = m.Update(key(tea.KeyEscape))
	got := m.(Model)
	if got.pickerOpen {
		t.Error("esc should close the menu")
	}
	if got.provider != "ollama" {
		t.Errorf("esc changed provider to %q", got.provider)
	}
}

func TestMenuSelectErrorStaysOpen(t *testing.T) {
	m := menuModel(func(id string) (string, error) { return "", errors.New("needs GROQ_API_KEY") })
	m, _ = m.Update(ctrlP())
	m, _ = m.Update(key(tea.KeyDown))
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if !got.pickerOpen {
		t.Error("menu should stay open on a select error")
	}
	if got.pickerErr == "" {
		t.Error("expected an error message on the menu")
	}
	if got.provider != "ollama" {
		t.Error("provider should be unchanged on error")
	}
}

func TestCtrlPInertWithoutMenu(t *testing.T) {
	var m tea.Model = New("ollama", "llama3", nil) // no WithProviderMenu
	m, _ = m.Update(ctrlP())
	if m.(Model).pickerOpen {
		t.Error("ctrl+p should be inert without a configured menu")
	}
}
