package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func typeString(m tea.Model, s string) tea.Model {
	for _, r := range s {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func TestSubmitCallsResponderAndAppends(t *testing.T) {
	var got string
	m := tea.Model(New("ollama", "llama3", func(goal string) string {
		got = goal
		return "reply to: " + goal
	}))
	m = typeString(m, "hello agent")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if got != "hello agent" {
		t.Fatalf("responder received %q", got)
	}
	tr := m.(Model).Transcript()
	// system banner + user + agent
	if len(tr) != 3 {
		t.Fatalf("transcript = %v", tr)
	}
	if !strings.Contains(tr[1], "hello agent") || !strings.Contains(tr[2], "reply to: hello agent") {
		t.Fatalf("transcript content = %v", tr)
	}
}

func TestBackspaceEditsInput(t *testing.T) {
	m := tea.Model(New("p", "m", nil))
	m = typeString(m, "abcd")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	view := m.(Model).View()
	if !strings.Contains(view, "abc▏") {
		t.Errorf("view input = %q", view)
	}
}

func TestEmptySubmitIsNoop(t *testing.T) {
	m := tea.Model(New("p", "m", func(string) string { t.Fatal("responder should not run"); return "" }))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.(Model).Transcript()) != 1 { // only the banner
		t.Errorf("empty submit changed transcript")
	}
}

func TestEscQuits(t *testing.T) {
	m := tea.Model(New("p", "m", nil))
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc should return a quit command")
	}
}

func TestViewRendersBannerAndStatus(t *testing.T) {
	m := New("anthropic", "claude", nil)
	m.width = 100
	view := m.View()
	if !strings.Contains(view, Tagline) {
		t.Error("view missing tagline banner")
	}
	if !strings.Contains(view, "anthropic") || !strings.Contains(view, "ready") {
		t.Error("view missing status bar content")
	}
}
