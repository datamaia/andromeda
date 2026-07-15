package tui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// stripANSI removes escape sequences so tests can assert on visible text.
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			inEsc = true
		case inEsc && ((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')):
			inEsc = false
		case !inEsc:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Regression: a slash command run on a fresh session (splash still up) must show its output. Before
// the fix, the splash suppressed every system line, so /graph, /ontology, etc. produced no visible
// feedback until after the first conversation turn. Here /graph → build emits output that must be on
// screen, and the splash tagline must be gone.
func TestCommandOutputVisibleOnStartScreen(t *testing.T) {
	acts := Actions{Graph: func(_ context.Context, _ string) string { return "graph · 5 nodes written to .andromeda/graph" }}
	var m tea.Model = New("ollama", "llama3", nil).WithActions(acts)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Fresh launch shows the splash (its tagline is literal text; the wordmark is block glyphs).
	if got := stripANSI(m.(Model).View().Content); !strings.Contains(got, "Your terminal companion") {
		t.Fatal("fresh session should show the brand splash")
	}

	m = typeString(m, "/graph")
	m, _ = m.Update(key(tea.KeyEnter)) // open the graph menu
	m, _ = m.Update(key(tea.KeyEnter)) // select "Build"
	view := stripANSI(m.(Model).View().Content)
	if !strings.Contains(view, "written to .andromeda/graph") {
		t.Errorf("graph build feedback not visible on the start screen:\n%s", view)
	}
	if strings.Contains(view, "Your terminal companion") {
		t.Error("splash tagline should be gone once a command has produced output")
	}
}
