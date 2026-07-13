package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// A long conversation is scrollable: by default the latest lines show; PageUp reveals older lines
// and a "more" indicator; End returns to the bottom. Regression: history could not be navigated.
func TestTranscriptScrolls(t *testing.T) {
	m := New("groq", "x", nil)
	m.width, m.height = 60, 12
	for i := 0; i < 40; i++ {
		role := "user"
		if i%2 == 1 {
			role = "agent"
		}
		m.transcript = append(m.transcript, entry{role, fmt.Sprintf("message number %02d", i)})
	}
	var tm tea.Model = m

	// At the bottom, the latest message is visible and there is no room for the earliest.
	bottom := tm.(Model).View().Content
	if !strings.Contains(bottom, "message number 39") {
		t.Fatalf("bottom view should show the latest message:\n%s", bottom)
	}
	if strings.Contains(bottom, "message number 00") {
		t.Fatalf("bottom view should not show the earliest message yet")
	}

	// PageUp scrolls back and shows the "more below/above" indicator.
	tm, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	up := tm.(Model).View().Content
	if tm.(Model).scrollOffset == 0 {
		t.Fatal("PageUp should increase the scroll offset")
	}
	if !strings.Contains(up, "more") {
		t.Errorf("scrolled view should show a scroll indicator:\n%s", up)
	}

	// End jumps back to the latest.
	tm, _ = tm.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	if tm.(Model).scrollOffset != 0 {
		t.Fatal("End should return to the bottom (offset 0)")
	}
	if !strings.Contains(tm.(Model).View().Content, "message number 39") {
		t.Error("End should show the latest message again")
	}
}

// The mouse wheel scrolls the transcript: up reveals older output (back toward the oldest of the
// session), down returns toward the latest. This is distinct from ↑/↓, which recall input history.
func TestMouseWheelScrollsTranscript(t *testing.T) {
	m := New("groq", "x", nil)
	m.width, m.height = 60, 12
	for i := 0; i < 40; i++ {
		m.transcript = append(m.transcript, entry{"agent", fmt.Sprintf("message number %02d", i)})
	}
	var tm tea.Model = m

	// One wheel-up notch scrolls back a few lines.
	tm, _ = tm.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if got := tm.(Model).scrollOffset; got != wheelScrollLines {
		t.Fatalf("wheel up should scroll back %d lines, got offset %d", wheelScrollLines, got)
	}
	// Further notches keep moving toward the oldest output.
	tm, _ = tm.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	deep := tm.(Model).scrollOffset
	if deep <= wheelScrollLines {
		t.Fatalf("further wheel-up should keep scrolling back, got %d", deep)
	}
	// Wheel down returns toward the latest.
	tm, _ = tm.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if got := tm.(Model).scrollOffset; got != deep-wheelScrollLines {
		t.Fatalf("wheel down should scroll toward latest, got %d (from %d)", got, deep)
	}
}

// Regression: with mouse tracking off the terminal turned the wheel into ↑/↓ keys, so scrolling
// recalled previously submitted prompts. The wheel must move the transcript, never the input line.
func TestMouseWheelDoesNotRecallHistory(t *testing.T) {
	m := New("groq", "x", func(string, string) string { return "ok" })
	m.width, m.height = 60, 12
	var tm tea.Model = m
	tm = typeString(tm, "remember me")
	tm, _ = tm.Update(key(tea.KeyEnter)) // submit → clears input, records history
	tm, _ = tm.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if got := tm.(Model).input; got != "" {
		t.Fatalf("mouse wheel must not recall input history, input=%q", got)
	}
}

// Submitting a new goal snaps the view back to the bottom.
func TestSubmitResetsScroll(t *testing.T) {
	m := New("groq", "x", func(string, string) string { return "ok" })
	m.width, m.height = 60, 12
	for i := 0; i < 30; i++ {
		m.transcript = append(m.transcript, entry{"agent", fmt.Sprintf("line %02d", i)})
	}
	m.scrollOffset = 10 // pretend the user scrolled up
	var tm tea.Model = m
	tm = typeString(tm, "hola")
	tm, _ = tm.Update(key(tea.KeyEnter))
	if tm.(Model).scrollOffset != 0 {
		t.Fatalf("submitting should reset scroll to bottom, got offset %d", tm.(Model).scrollOffset)
	}
}
