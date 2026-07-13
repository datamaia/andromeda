package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// A large model list must render within a viewport, not dump every row (which hid the cursor and
// overflowed the screen). Regression for "menu enorme sin cursor" on OpenRouter's 400+ models.
func TestModelPickerViewportBounded(t *testing.T) {
	many := make([]string, 300)
	for i := range many {
		many[i] = fmt.Sprintf("model-%03d", i)
	}
	m := New("openrouter", "model-000", nil).
		WithActions(Actions{Models: func(context.Context) []string { return many }})
	m.height = 24
	var tm tea.Model = m
	tm = typeString(tm, "/model")
	tm, cmd := tm.Update(key(tea.KeyEnter)) // async discovery
	tm = stepCmd(tm, cmd)                   // model picker opens
	got := tm.(Model)
	if !got.pickerOpen {
		t.Fatal("model picker should be open")
	}
	view := got.View().Content
	// Not every one of the 300 models can be on screen at once.
	rendered := strings.Count(view, "model-")
	if rendered >= 300 {
		t.Fatalf("viewport rendered %d rows; expected a bounded window", rendered)
	}
	if !strings.Contains(view, "more") {
		t.Error("a scrolled list should show a '↓ N more' indicator")
	}
}

// Typing filters the picker; Enter selects the (single) match.
func TestModelPickerTypeToFilter(t *testing.T) {
	models := []string{"gpt-oss-120b", "llama-3.3-70b", "qwen2.5-coder", "deepseek-r1"}
	var chosen string
	m := New("groq", "gpt-oss-120b", nil).
		WithActions(Actions{Models: func(context.Context) []string { return models }}).
		WithModelSelect(func(id string) { chosen = id })
	var tm tea.Model = m
	tm = typeString(tm, "/model")
	tm, cmd := tm.Update(key(tea.KeyEnter))
	tm = stepCmd(tm, cmd) // picker open
	// Type "qwen" to narrow to one match.
	tm = typeString(tm, "qwen")
	got := tm.(Model)
	if len(got.filteredItems()) != 1 {
		t.Fatalf("filter 'qwen' matched %d, want 1", len(got.filteredItems()))
	}
	_, _ = tm.Update(key(tea.KeyEnter))
	if chosen != "qwen2.5-coder" {
		t.Fatalf("selected %q, want qwen2.5-coder", chosen)
	}
}

// Esc first clears a non-empty filter (revealing the full list) before it closes the picker.
func TestPickerEscClearsFilterFirst(t *testing.T) {
	models := []string{"a-model", "b-model", "c-model"}
	m := New("groq", "a-model", nil).
		WithActions(Actions{Models: func(context.Context) []string { return models }})
	var tm tea.Model = m
	tm = typeString(tm, "/model")
	tm, cmd := tm.Update(key(tea.KeyEnter))
	tm = stepCmd(tm, cmd)
	tm = typeString(tm, "b")
	tm, _ = tm.Update(key(tea.KeyEscape)) // clears filter
	got := tm.(Model)
	if !got.pickerOpen || got.pickerFilter != "" {
		t.Fatalf("first esc should clear the filter and keep the picker open (open=%v filter=%q)", got.pickerOpen, got.pickerFilter)
	}
	tm, _ = tm.Update(key(tea.KeyEscape)) // now closes
	if tm.(Model).pickerOpen {
		t.Error("second esc should close the picker")
	}
}

// Choosing a provider mid-session (not onboarding) opens the model picker so the user picks a live
// model instead of silently inheriting a possibly-retired catalog default.
func TestMidSessionProviderSwitchOpensModelPicker(t *testing.T) {
	choices := []ProviderChoice{{"groq", "Groq", "GROQ_API_KEY", ""}, {"cerebras", "Cerebras", "CEREBRAS_API_KEY", ""}}
	m := New("groq", "llama-3.3-70b", nil).
		WithProviderMenu(choices, func(_ string) (string, error) { return "default-model", nil }).
		WithActions(Actions{Models: func(context.Context) []string { return []string{"gpt-oss-120b", "qwen-3"} }})
	var tm tea.Model = m
	// Open the provider picker mid-session and choose cerebras.
	tm = typeString(tm, "/provider")
	tm, _ = tm.Update(key(tea.KeyEnter)) // run /provider → provider picker
	tm, _ = tm.Update(key(tea.KeyDown))  // to cerebras
	tm, cmd := tm.Update(key(tea.KeyEnter))
	tm = stepCmd(tm, cmd) // provider applied → model discovery → model picker
	got := tm.(Model)
	if !got.pickerOpen || got.pickerKind != "model" {
		t.Fatalf("provider switch should open the model picker (open=%v kind=%q)", got.pickerOpen, got.pickerKind)
	}
}
