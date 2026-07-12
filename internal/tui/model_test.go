package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// press builds a Bubble Tea v2 key-press message for a printable rune (Text carries the chars).
func press(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }

// key builds a v2 key-press message for a special key (no printable text).
func key(code rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: code} }

func typeString(m tea.Model, s string) tea.Model {
	for _, r := range s {
		m, _ = m.Update(press(r))
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
	m, _ = m.Update(key(tea.KeyEnter))

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
	m, _ = m.Update(key(tea.KeyBackspace))
	view := m.(Model).View().Content
	if !strings.Contains(view, "abc▏") {
		t.Errorf("view input = %q", view)
	}
}

func TestEmptySubmitIsNoop(t *testing.T) {
	m := tea.Model(New("p", "m", func(string) string { t.Fatal("responder should not run"); return "" }))
	m, _ = m.Update(key(tea.KeyEnter))
	if len(m.(Model).Transcript()) != 1 { // only the banner
		t.Errorf("empty submit changed transcript")
	}
}

func TestEscQuits(t *testing.T) {
	m := tea.Model(New("p", "m", nil))
	_, cmd := m.Update(key(tea.KeyEscape))
	if cmd == nil {
		t.Fatal("esc should return a quit command")
	}
}

func TestCtrlCQuits(t *testing.T) {
	m := tea.Model(New("p", "m", nil))
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("ctrl+c should return a quit command")
	}
}

func TestViewRendersSplashAndStatus(t *testing.T) {
	m := New("anthropic", "claude", nil)
	m.width = 100
	view := m.View().Content
	if !strings.Contains(view, Tagline) {
		t.Error("start-screen view missing the tagline")
	}
	if !strings.Contains(view, "andromeda") {
		t.Error("start-screen view missing the wordmark")
	}
	if !strings.Contains(view, "anthropic") || !strings.Contains(view, "ready") {
		t.Error("view missing status bar content")
	}
}

func TestSplashHiddenAfterExchange(t *testing.T) {
	m := tea.Model(New("p", "m", func(string) string { return "ok" }))
	m = typeString(m, "hi")
	m, _ = m.Update(key(tea.KeyEnter))
	// After an exchange the mascot splash is no longer shown.
	if strings.Contains(m.(Model).View().Content, Tagline) {
		t.Error("splash should be hidden once the conversation starts")
	}
}
