package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// submitLine types a line and presses enter (synchronous responder path).
func submitLine(m tea.Model, s string) tea.Model {
	m = typeString(m, s)
	m, _ = m.Update(key(tea.KeyEnter))
	return m
}

func TestHistoryRecallUpDown(t *testing.T) {
	m := tea.Model(New("p", "mdl", func(string, string) string { return "ok" }))
	m = submitLine(m, "one")
	m = submitLine(m, "two")

	m, _ = m.Update(key(tea.KeyUp))
	if got := m.(Model).input; got != "two" {
		t.Fatalf("first ↑ = %q, want two", got)
	}
	m, _ = m.Update(key(tea.KeyUp))
	if got := m.(Model).input; got != "one" {
		t.Fatalf("second ↑ = %q, want one", got)
	}
	m, _ = m.Update(key(tea.KeyDown))
	if got := m.(Model).input; got != "two" {
		t.Fatalf("↓ = %q, want two", got)
	}
	m, _ = m.Update(key(tea.KeyDown))
	if got := m.(Model).input; got != "" {
		t.Fatalf("↓ past newest = %q, want empty draft", got)
	}
}

func TestHistoryDedupesConsecutive(t *testing.T) {
	m := tea.Model(New("p", "mdl", func(string, string) string { return "ok" }))
	m = submitLine(m, "same")
	m = submitLine(m, "same")
	if n := len(m.(Model).promptHistory); n != 1 {
		t.Fatalf("promptHistory has %d entries, want 1 (deduped)", n)
	}
}

func TestHistoryPreservesLiveDraft(t *testing.T) {
	m := tea.Model(New("p", "mdl", func(string, string) string { return "ok" }))
	m = submitLine(m, "past")
	m = typeString(m, "draft in progress")

	m, _ = m.Update(key(tea.KeyUp)) // recall "past", saving the draft
	if got := m.(Model).input; got != "past" {
		t.Fatalf("↑ = %q, want past", got)
	}
	m, _ = m.Update(key(tea.KeyDown)) // back to the live draft
	if got := m.(Model).input; got != "draft in progress" {
		t.Fatalf("↓ restored %q, want the live draft", got)
	}
}

func TestHistoryIncludesSlashLines(t *testing.T) {
	m := tea.Model(New("p", "mdl", func(string, string) string { return "ok" }))
	m = submitLine(m, "/keys")
	m, _ = m.Update(key(tea.KeyUp))
	if got := m.(Model).input; got != "/keys" {
		t.Fatalf("↑ after slash submit = %q, want /keys", got)
	}
}
