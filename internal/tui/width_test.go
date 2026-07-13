package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// The transcript and status bar must fit the terminal width — no rendered line may exceed it. The
// user reported long lines being cut off / not adapting to the terminal size.
func TestOutputFitsWidth(t *testing.T) {
	m := New("gemini", "models/gemini-3.1-flash-lite", nil)
	m.width, m.height = 40, 20
	m.transcript = append(m.transcript, entry{"agent",
		"This is a deliberately long assistant reply that would overflow a narrow terminal if the renderer did not wrap it to the available width."})
	// Simulate a window-size message and render.
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	view := tm.(Model).View().Content
	for i, line := range strings.Split(view, "\n") {
		if w := ansi.StringWidth(line); w > 40 {
			t.Errorf("line %d width %d exceeds terminal width 40: %q", i, w, ansi.Strip(line))
		}
	}
}

// The status bar drops its hint (and truncates if needed) on a narrow terminal instead of bleeding.
func TestStatusBarFitsNarrow(t *testing.T) {
	m := New("openrouter", "openai/gpt-oss-120b", nil)
	m.width = 30
	if w := ansi.StringWidth(m.statusBar()); w > 30 {
		t.Errorf("status bar width %d exceeds 30", w)
	}
}
