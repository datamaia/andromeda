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
	// The list is bounded to a scrolling window; the first commands and a "more" marker are shown.
	for _, want := range []string{"/help", "/provider", "↓"} {
		if !strings.Contains(view, want) {
			t.Errorf("palette view missing %q", want)
		}
	}
	// Every built-in is still reachable via the (unbounded) filter, e.g. quit.
	if resolveCommand("quit", got.mergedCommands()) == nil {
		t.Error("quit command missing from the registry")
	}
}

// Typing narrows and ranks the palette: prefix matches lead, then substring matches.
func TestPaletteFilters(t *testing.T) {
	var m tea.Model = New("ollama", "llama3", nil)
	m = typeString(m, "/mo")
	cmds := m.(Model).filteredCommands()
	if len(cmds) == 0 || cmds[0].name != "model" {
		t.Fatalf("filter /mo = %v, want model ranked first", names(cmds))
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
	if !strings.Contains(got.headerString(), "plan") {
		t.Error("header should show the active mode")
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
	_, _ = m.Update(key(tea.KeyEnter))
	if gotMode != "shell" {
		t.Fatalf("responder received mode %q, want shell", gotMode)
	}
}

// Shift+Tab cycles the interaction mode agent → plan → shell → agent.
func TestShiftTabCyclesMode(t *testing.T) {
	shiftTab := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	var m tea.Model = New("ollama", "llama3", nil)
	if got := m.(Model).modeOrDefault(); got != "agent" {
		t.Fatalf("initial mode = %q, want agent", got)
	}
	for _, want := range []string{"plan", "shell", "agent"} {
		m, _ = m.Update(shiftTab)
		if got := m.(Model).mode; got != want {
			t.Fatalf("after shift+tab mode = %q, want %q", got, want)
		}
	}
}

// /goal <text> runs the responder just like a typed goal.
func TestGoalCommandRunsResponder(t *testing.T) {
	var seen string
	m := tea.Model(New("ollama", "llama3", func(g, _ string) string { seen = g; return "done" }))
	m = typeString(m, "/goal ship it")
	_, _ = m.Update(key(tea.KeyEnter))
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
	m = typeString(m, "/config")
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
	tm, cmd := tm.Update(key(tea.KeyEnter)) // run /model → async discovery
	tm = stepCmd(tm, cmd)                   // discovery → model picker
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

// Choosing a model propagates to the driver (so the agent runs on it), not just the view.
func TestModelSelectPropagatesToDriver(t *testing.T) {
	var driverModel string
	m := New("gemini", "gemini-2.5-flash", nil).
		WithActions(Actions{Models: func(context.Context) []string { return []string{"gemini-2.5-flash", "gemini-3.1-flash-lite"} }}).
		WithModelSelect(func(id string) { driverModel = id })
	var tm tea.Model = m
	tm = typeString(tm, "/model")
	tm, cmd := tm.Update(key(tea.KeyEnter)) // discovery
	tm = stepCmd(tm, cmd)
	tm, _ = tm.Update(key(tea.KeyDown))  // to gemini-3.1-flash-lite
	tm, _ = tm.Update(key(tea.KeyEnter)) // select it
	if driverModel != "gemini-3.1-flash-lite" {
		t.Errorf("driver model = %q, want gemini-3.1-flash-lite", driverModel)
	}
	if tm.(Model).model != "gemini-3.1-flash-lite" {
		t.Errorf("view model = %q", tm.(Model).model)
	}
}

// /model <name> also propagates to the driver.
func TestModelArgPropagatesToDriver(t *testing.T) {
	var driverModel string
	var m tea.Model = New("groq", "x", nil).WithModelSelect(func(id string) { driverModel = id })
	m = typeString(m, "/model llama-3.3-70b-versatile")
	_, _ = m.Update(key(tea.KeyEnter))
	if driverModel != "llama-3.3-70b-versatile" {
		t.Errorf("driver model = %q", driverModel)
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

// A line that is really a filesystem path must reach the responder as a goal, not be rejected as an
// unknown slash command (regression: "/Users/me/project" printed "unknown command").
func TestPathIsAGoalNotACommand(t *testing.T) {
	var seen string
	m := tea.Model(New("ollama", "llama3", func(g, _ string) string { seen = g; return "ok" }))
	m = typeString(m, "/Users/maia/Documents/lyra/andromeda")
	m, _ = m.Update(key(tea.KeyEnter))
	if seen != "/Users/maia/Documents/lyra/andromeda" {
		t.Fatalf("path should reach the responder as a goal, got %q", seen)
	}
	tr := m.(Model).Transcript()
	if strings.Contains(tr[len(tr)-1], "unknown command") {
		t.Error("a path must not be reported as an unknown command")
	}
}

func TestLooksLikeSlashCommand(t *testing.T) {
	cases := map[string]bool{
		"/help":               true,
		"/model gpt-4":        true,
		"/goal do the thing":  true,
		"/Users/maia/project": false,
		"/tmp/out.md":         false,
		"/home/x/.config":     false,
		"~/notes":             false,
		"hello world":         false,
	}
	for line, want := range cases {
		if got := looksLikeSlashCommand(line); got != want {
			t.Errorf("looksLikeSlashCommand(%q) = %v, want %v", line, got, want)
		}
	}
}

func names(cmds []slashCommand) []string {
	out := make([]string, len(cmds))
	for i, c := range cmds {
		out[i] = c.name
	}
	return out
}

// /effort <level> sets the reasoning effort and propagates it to the driver.
func TestEffortCommandPropagates(t *testing.T) {
	var got string
	m := tea.Model(New("p", "m", nil).WithEffortSelect(func(e string) { got = e }))
	m = typeString(m, "/effort high")
	m, _ = m.Update(key(tea.KeyEnter))
	if got != "high" {
		t.Fatalf("driver effort = %q, want high", got)
	}
	if m.(Model).effort != "high" {
		t.Fatalf("view effort = %q, want high", m.(Model).effort)
	}
}

// /theme light switches the live style set.
func TestThemeCommandSwitches(t *testing.T) {
	var m tea.Model = New("p", "m", nil)
	m = typeString(m, "/theme light")
	m, _ = m.Update(key(tea.KeyEnter))
	if got := m.(Model).theme; got != "light" {
		t.Fatalf("theme = %q, want light", got)
	}
}

// Typing an argument after a command enters argument-completion mode and ranks candidates.
func TestArgCompletionRanks(t *testing.T) {
	var m tea.Model = New("p", "m", nil)
	m = typeString(m, "/effort me")
	if got := m.(Model).menuKind(); got != "arg" {
		t.Fatalf("menuKind = %q, want arg", got)
	}
	if args := filteredArgs("effort", "me"); len(args) != 1 || args[0] != "medium" {
		t.Fatalf("filteredArgs = %v, want [medium]", args)
	}
}

// A custom command expands its template ($ARGUMENTS/$1) and reaches the responder as a goal.
func TestCustomCommandExpands(t *testing.T) {
	var seen string
	m := New("p", "m", func(g, _ string) string { seen = g; return "ok" }).
		WithCustomCommands([]CustomCommand{{Name: "greet", Desc: "greet", Template: "Say hi to $1 in $ARGUMENTS"}})
	var tm tea.Model = m
	tm = typeString(tm, "/greet world over")
	_, _ = tm.Update(key(tea.KeyEnter))
	if seen != "Say hi to world in world over" {
		t.Fatalf("expanded goal = %q", seen)
	}
}
