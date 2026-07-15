package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// /status opens the tabbed panel; Tab cycles tabs; a number jumps directly; Esc closes.
func TestStatusPanelNavigation(t *testing.T) {
	var m tea.Model = New("groq", "llama-3.3-70b", nil)
	m = typeString(m, "/status")
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if !got.statusPanel {
		t.Fatal("/status should open the panel")
	}
	view := got.View().Content
	if !strings.Contains(view, "Status") || !strings.Contains(view, "Overview") {
		t.Errorf("panel view missing tab strip: %q", view)
	}
	// Tab → Usage (index 1)
	m, _ = m.Update(key(tea.KeyTab))
	if m.(Model).statusTab != 1 {
		t.Errorf("after Tab statusTab = %d, want 1", m.(Model).statusTab)
	}
	// "3" jumps to Tools (index 2)
	m, _ = m.Update(press('3'))
	if m.(Model).statusTab != 2 {
		t.Errorf("after '3' statusTab = %d, want 2", m.(Model).statusTab)
	}
	// Esc closes
	m, _ = m.Update(key(tea.KeyEscape))
	if m.(Model).statusPanel {
		t.Error("esc should close the panel")
	}
}

// /status <tab> opens directly on the named tab.
func TestStatusPanelJumpArg(t *testing.T) {
	var m tea.Model = New("p", "m", nil)
	m = typeString(m, "/status tools")
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if !got.statusPanel || statusTabNames[got.statusTab] != "tools" {
		t.Fatalf("/status tools should open the Tools tab, got tab=%d open=%v", got.statusTab, got.statusPanel)
	}
}

// Reported token usage from a run accumulates and shows on the Usage tab and the status bar.
func TestTokenUsageAccumulates(t *testing.T) {
	events := []AgentEvent{{Final: "done", InTokens: 1200, OutTokens: 340}}
	var m tea.Model = New("groq", "x", nil).WithAgentRunner(streamRunner(events, nil))
	m = typeString(m, "do it")
	m, cmd := m.Update(key(tea.KeyEnter))
	m = drain(m, cmd)
	got := m.(Model)
	if got.inTokens != 1200 || got.outTokens != 340 {
		t.Fatalf("tokens = %d/%d, want 1200/340", got.inTokens, got.outTokens)
	}
	if !strings.Contains(got.headerString(), "1.2k") {
		t.Errorf("header should show abbreviated token count: %q", got.headerString())
	}
}

func TestHumanCount(t *testing.T) {
	cases := map[int]string{0: "0", 999: "999", 1200: "1.2k", 1_500_000: "1.5M"}
	for n, want := range cases {
		if got := humanCount(n); got != want {
			t.Errorf("humanCount(%d) = %q, want %q", n, got, want)
		}
	}
}

// A resumed session re-seeds the transcript from restored history (no splash, prior turns visible).
func TestWithHistoryReseedsTranscript(t *testing.T) {
	m := New("groq", "x", nil).WithHistory([]HistoryEntry{
		{Role: "user", Text: "add a feature"},
		{Role: "agent", Text: "here is what I did"},
	})
	if !m.hasContent() {
		t.Error("a resumed session should not show the start splash")
	}
	tr := m.Transcript()
	joined := strings.Join(tr, "\n")
	if !strings.Contains(joined, "add a feature") || !strings.Contains(joined, "here is what I did") {
		t.Errorf("restored transcript missing turns: %v", tr)
	}
	if !strings.Contains(joined, "resumed session") {
		t.Errorf("expected a resumed-session note: %v", tr)
	}
}
