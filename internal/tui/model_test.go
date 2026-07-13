package tui

import (
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(view, "_____") {
		t.Error("start-screen view missing the ANDROMEDA wordmark art")
	}
	if !strings.Contains(view, "anthropic") || !strings.Contains(view, "ready") {
		t.Error("view missing status bar content")
	}
}

// The status bar reports, live, the provider/model/effort, elapsed session time, and state.
func TestStatusBarShowsSessionInfo(t *testing.T) {
	m := New("groq", "llama-3.3-70b", nil)
	m.effort = "medium"
	m.now = m.started.Add(75 * time.Second)
	bar := m.statusBar()
	for _, want := range []string{"groq", "llama-3.3-70b", "effort medium", "1:15", "ready"} {
		if !strings.Contains(bar, want) {
			t.Errorf("status bar missing %q: %q", want, bar)
		}
	}
}

// Effort is hidden until set; uptime rolls over to H:MM:SS past an hour.
func TestStatusBarEffortHiddenAndHourFormat(t *testing.T) {
	m := New("ollama", "llama3", nil)
	if strings.Contains(m.statusBar(), "effort") {
		t.Error("effort should be hidden when unset")
	}
	m.now = m.started.Add(3725 * time.Second) // 1h 02m 05s
	if got := m.uptime(); got != "1:02:05" {
		t.Errorf("uptime = %q, want 1:02:05", got)
	}
}

// The session clock advances on each tick message.
func TestTickAdvancesClock(t *testing.T) {
	var m tea.Model = New("p", "mdl", nil)
	next := m.(Model).started.Add(42 * time.Second)
	m, cmd := m.Update(tickMsg(next))
	if cmd == nil {
		t.Error("tick should reschedule itself")
	}
	if got := m.(Model).uptime(); got != "0:42" {
		t.Errorf("uptime after tick = %q, want 0:42", got)
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
