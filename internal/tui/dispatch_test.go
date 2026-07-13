package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Every built-in slash command dispatches and renders without panicking, even with no Actions wired
// (unbacked commands report "unavailable" rather than crashing). This exercises the command handlers
// and the overlays they open.
func TestSlashCommandsDispatch(t *testing.T) {
	cmds := []string{
		"/help", "/keys", "/compact", "/loop", "/reset", "/status", "/effort", "/theme",
		"/memory", "/skills", "/workflows", "/config", "/export", "/login", "/logout",
		"/init", "/doctor", "/update", "/mcp", "/agent", "/plan", "/shell",
	}
	for _, c := range cmds {
		m := New("ollama", "llama3", nil)
		m.width, m.height = 80, 24
		var tm tea.Model = m
		tm = typeString(tm, c)
		tm, _ = tm.Update(key(tea.KeyEnter))
		if got := tm.(Model).View().Content; got == "" {
			t.Errorf("%s produced an empty view", c)
		}
	}
}

// The /status panel opens, renders every tab as the arrow key walks across them, and closes on esc.
func TestStatusPanelTabsRender(t *testing.T) {
	m := New("groq", "llama-3.3-70b", nil)
	m.width, m.height = 90, 30
	var tm tea.Model = m
	tm = typeString(tm, "/status")
	tm, _ = tm.Update(key(tea.KeyEnter))
	if !tm.(Model).statusPanel {
		t.Fatal("/status should open the panel")
	}
	for i := 0; i < len(statusTabNames)+1; i++ {
		if tm.(Model).View().Content == "" {
			t.Errorf("tab %d rendered empty", i)
		}
		tm, _ = tm.Update(key(tea.KeyRight))
	}
	tm, _ = tm.Update(key(tea.KeyEscape))
	if tm.(Model).statusPanel {
		t.Error("esc should close the status panel")
	}
}

// Typing an "@" fragment renders without panic whether or not a file source is wired.
func TestAtMentionMenuRenders(t *testing.T) {
	m := New("groq", "x", nil)
	m.width, m.height = 80, 24
	var tm tea.Model = m
	tm = typeString(tm, "@re")
	if tm.(Model).View().Content == "" {
		t.Error("view empty while typing an @ mention")
	}
}
