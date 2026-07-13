package tui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Typing "/" opens the palette and the command list renders above the prompt.
func TestSlashOpensPalette(t *testing.T) {
	var m tea.Model = New("ollama", "llama3", nil)
	m = typeString(m, "/")
	got := m.(Model)
	if !got.paletteActive() {
		t.Fatal(`typing "/" should activate the palette`)
	}
	view := got.View().Content
	for _, want := range []string{"/help", "/provider", "/quit"} {
		if !strings.Contains(view, want) {
			t.Errorf("palette view missing %q", want)
		}
	}
}

// Typing narrows the palette to matching commands.
func TestPaletteFilters(t *testing.T) {
	var m tea.Model = New("ollama", "llama3", nil)
	m = typeString(m, "/mo")
	cmds := m.(Model).filteredCommands()
	if len(cmds) != 1 || cmds[0].name != "model" {
		t.Fatalf("filter /mo = %v, want [model]", names(cmds))
	}
}

// Arrow navigation + Enter runs the highlighted command (here /clear resets the transcript).
func TestPaletteNavigateAndRun(t *testing.T) {
	m := tea.Model(New("ollama", "llama3", func(string, string) string { return "ok" }))
	// seed an exchange so we can observe /clear
	m = typeString(m, "hi")
	m, _ = m.Update(key(tea.KeyEnter))
	if len(m.(Model).Transcript()) == 1 {
		t.Fatal("setup: expected an exchange before clear")
	}
	// "/clear" is uniquely matched by "/cl"; Enter runs it.
	m = typeString(m, "/cl")
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if got.paletteActive() {
		t.Error("palette should close after running a command")
	}
	tr := got.Transcript()
	if len(tr) != 1 || !strings.Contains(tr[0], "cleared") {
		t.Errorf("after /clear transcript = %v", tr)
	}
}

// Esc closes the palette (clears the typed command) without quitting the program.
func TestPaletteEscClosesWithoutQuitting(t *testing.T) {
	var m tea.Model = New("ollama", "llama3", nil)
	m = typeString(m, "/prov")
	m, cmd := m.Update(key(tea.KeyEscape))
	got := m.(Model)
	if cmd != nil {
		t.Error("esc in the palette should not quit")
	}
	if got.paletteActive() || got.input != "" {
		t.Errorf("esc should clear the palette input, got %q", got.input)
	}
	if got.quitting {
		t.Error("esc in the palette should not set quitting")
	}
}

// Submitting a full "/command args" line dispatches to the handler (mode switch here).
func TestSlashCommandSwitchesMode(t *testing.T) {
	var m tea.Model = New("ollama", "llama3", nil)
	m = typeString(m, "/plan")
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if got.mode != "plan" {
		t.Fatalf("mode = %q, want plan", got.mode)
	}
	if !strings.Contains(got.statusBar(), "plan") {
		t.Error("status bar should show the active mode")
	}
}

// The active interaction mode is passed to the responder so the driver can enforce it.
func TestResponderReceivesActiveMode(t *testing.T) {
	var gotMode string
	m := tea.Model(New("ollama", "llama3", func(_, mode string) string { gotMode = mode; return "ok" }))
	// switch to shell mode, then submit a line
	m = typeString(m, "/shell")
	m, _ = m.Update(key(tea.KeyEnter))
	m = typeString(m, "ls -la")
	m, _ = m.Update(key(tea.KeyEnter))
	if gotMode != "shell" {
		t.Fatalf("responder received mode %q, want shell", gotMode)
	}
}

// /goal <text> runs the responder just like a typed goal.
func TestGoalCommandRunsResponder(t *testing.T) {
	var seen string
	m := tea.Model(New("ollama", "llama3", func(g, _ string) string { seen = g; return "done" }))
	m = typeString(m, "/goal ship it")
	m, _ = m.Update(key(tea.KeyEnter))
	if seen != "ship it" {
		t.Fatalf("responder got %q, want 'ship it'", seen)
	}
}

// An unknown slash command reports itself instead of being sent to the model.
func TestUnknownSlashCommand(t *testing.T) {
	m := tea.Model(New("ollama", "llama3", func(string, string) string {
		t.Fatal("unknown command must not reach the responder")
		return ""
	}))
	m = typeString(m, "/nope now")
	m, _ = m.Update(key(tea.KeyEnter))
	tr := m.(Model).Transcript()
	if !strings.Contains(tr[len(tr)-1], "unknown command") {
		t.Errorf("expected an unknown-command notice, got %v", tr)
	}
}

// App-backed actions surface their text; unwired ones degrade gracefully.
func TestDoctorActionWired(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Doctor: func(context.Context) string { return "all systems go" },
	})
	var tm tea.Model = m
	tm = typeString(tm, "/doctor")
	tm, _ = tm.Update(key(tea.KeyEnter))
	tr := tm.(Model).Transcript()
	if !strings.Contains(tr[len(tr)-1], "all systems go") {
		t.Errorf("doctor action not surfaced: %v", tr)
	}
}

func TestUnwiredActionDegrades(t *testing.T) {
	var m tea.Model = New("ollama", "llama3", nil) // no actions
	m = typeString(m, "/mcp")
	m, _ = m.Update(key(tea.KeyEnter))
	tr := m.(Model).Transcript()
	if !strings.Contains(tr[len(tr)-1], "not available") {
		t.Errorf("unwired action should degrade with a hint: %v", tr)
	}
}

// /model with no args opens a navigable picker of discovered models; Enter sets the model.
func TestModelCommandOpensPicker(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Models: func(context.Context) []string { return []string{"llama3", "qwen2.5-coder", "deepseek-r1"} },
	})
	var tm tea.Model = m
	tm = typeString(tm, "/model")
	tm, _ = tm.Update(key(tea.KeyEnter)) // run the highlighted /model command
	got := tm.(Model)
	if !got.pickerOpen || got.pickerTitle != "Select a model" {
		t.Fatalf("/model should open the model picker, got open=%v title=%q", got.pickerOpen, got.pickerTitle)
	}
	// cursor starts on the current model (llama3, index 0); move to deepseek-r1 (index 2) and select
	tm, _ = tm.Update(key(tea.KeyDown))
	tm, _ = tm.Update(key(tea.KeyDown))
	tm, _ = tm.Update(key(tea.KeyEnter))
	got = tm.(Model)
	if got.pickerOpen {
		t.Error("picker should close after selecting a model")
	}
	if got.model != "deepseek-r1" {
		t.Errorf("model = %q, want deepseek-r1", got.model)
	}
}

// /model <name> sets the model directly without opening the picker.
func TestModelCommandWithArgSetsDirectly(t *testing.T) {
	var m tea.Model = New("ollama", "llama3", nil)
	m = typeString(m, "/model qwen2.5")
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if got.pickerOpen {
		t.Error("/model <name> should not open the picker")
	}
	if got.model != "qwen2.5" {
		t.Errorf("model = %q, want qwen2.5", got.model)
	}
}

func names(cmds []slashCommand) []string {
	out := make([]string, len(cmds))
	for i, c := range cmds {
		out[i] = c.name
	}
	return out
}
