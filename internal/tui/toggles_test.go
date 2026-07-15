package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCmdDetailsToggle(t *testing.T) {
	m := New("ollama", "llama3", nil)
	nm, _ := cmdDetails(m, "on")
	if !nm.(Model).showDetails || !strings.Contains(lastText(nm.(Model)), "ON") {
		t.Fatalf("on: showDetails=%v out=%q", nm.(Model).showDetails, lastText(nm.(Model)))
	}
	nm, _ = cmdDetails(nm.(Model), "")
	if nm.(Model).showDetails {
		t.Fatal("bare /details should toggle back off")
	}
	nm, _ = cmdDetails(m, "frob")
	if !strings.Contains(lastText(nm.(Model)), "usage") {
		t.Fatalf("bad arg: %q", lastText(nm.(Model)))
	}
}

func TestToolDetailsVerbosity(t *testing.T) {
	call := &ToolStep{Phase: "call", Name: "fs_read", Input: `{"path":"/x/y.go","limit":40}`}
	result := &ToolStep{Phase: "result", Result: `{"content":"package main\nfunc main(){}"}`}

	// Compact (default): the argument JSON is not shown verbatim.
	compact := New("ollama", "llama3", nil)
	compact = compact.appendToolStep(call).appendToolStep(result)
	ctext := compact.transcript[compact.toolIdx].text
	if strings.Contains(ctext, `"limit":40`) {
		t.Fatalf("compact mode should not dump raw args: %q", ctext)
	}

	// Details on: the full arguments appear.
	verbose := New("ollama", "llama3", nil)
	verbose.showDetails = true
	verbose = verbose.appendToolStep(call).appendToolStep(result)
	vtext := verbose.transcript[verbose.toolIdx].text
	if !strings.Contains(vtext, `"limit":40`) {
		t.Fatalf("details mode should show full args: %q", vtext)
	}
}

func TestToolResultDetail(t *testing.T) {
	got := toolResultDetail(`{"output":"hello world"}`)
	if !strings.Contains(got, "hello world") {
		t.Fatalf("detail unwrap: %q", got)
	}
	if toolResultDetail("  ") != "" {
		t.Fatal("blank result should be empty")
	}
}

func TestCmdEditorUnavailableAndWired(t *testing.T) {
	// No Editor action.
	m := New("ollama", "llama3", nil)
	nm, cmd := cmdEditor(m, "")
	if cmd != nil || !strings.Contains(lastText(nm.(Model)), "not available") {
		t.Fatalf("editor unavailable: cmd=%v out=%q", cmd, lastText(nm.(Model)))
	}
	// Wired: returns the driver's command untouched.
	var seededWith string
	m2 := New("ollama", "llama3", nil).WithActions(Actions{
		Editor: func(seed string) tea.Cmd { seededWith = seed; return func() tea.Msg { return nil } },
	})
	m2.input = "draft prompt"
	_, cmd = cmdEditor(m2, "")
	if cmd == nil || seededWith != "draft prompt" {
		t.Fatalf("editor should run seeded with the composer text, got seed=%q cmd=%v", seededWith, cmd)
	}
}

func TestApplyEditor(t *testing.T) {
	m := New("ollama", "llama3", nil)
	// Error surfaces.
	nm, _ := m.applyEditor(EditorMsg{Err: errors.New("boom")})
	if !strings.Contains(lastText(nm.(Model)), "boom") {
		t.Fatalf("editor error: %q", lastText(nm.(Model)))
	}
	// Empty buffer is a no-send note.
	nm, _ = m.applyEditor(EditorMsg{Text: "   "})
	if !strings.Contains(lastText(nm.(Model)), "nothing to send") {
		t.Fatalf("empty editor: %q", lastText(nm.(Model)))
	}
	// A composed prompt is submitted (appears as a user line).
	nm, _ = m.applyEditor(EditorMsg{Text: "refactor the parser"})
	var sawUser bool
	for _, e := range nm.(Model).transcript {
		if e.role == "user" && strings.Contains(e.text, "refactor the parser") {
			sawUser = true
		}
	}
	if !sawUser {
		t.Fatal("composed prompt should be submitted as a user goal")
	}
	// Refused mid-run.
	running := New("ollama", "llama3", nil)
	running.running = true
	nm, _ = running.applyEditor(EditorMsg{Text: "x"})
	if !strings.Contains(lastText(nm.(Model)), "finish the current run") {
		t.Fatalf("editor mid-run: %q", lastText(nm.(Model)))
	}
}

func TestCmdUndoRedo(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Undo: func(_ context.Context) string { return "undo · reverted the workspace" },
		Redo: func(_ context.Context) string { return "redo · re-applied the change" },
	})
	if nm, _ := cmdUndo(m, ""); !strings.Contains(lastText(nm.(Model)), "reverted") {
		t.Fatalf("undo: %q", lastText(nm.(Model)))
	}
	if nm, _ := cmdRedo(m, ""); !strings.Contains(lastText(nm.(Model)), "re-applied") {
		t.Fatalf("redo: %q", lastText(nm.(Model)))
	}
	// Refused mid-run so a restore never races the agent's writes.
	m.running = true
	if nm, _ := cmdUndo(m, ""); !strings.Contains(lastText(nm.(Model)), "interrupt the current run") {
		t.Fatalf("undo mid-run: %q", lastText(nm.(Model)))
	}
	if nm, _ := cmdRedo(m, ""); !strings.Contains(lastText(nm.(Model)), "interrupt the current run") {
		t.Fatalf("redo mid-run: %q", lastText(nm.(Model)))
	}
}

func TestCmdUndoRedoUnavailable(t *testing.T) {
	m := New("ollama", "llama3", nil)
	if nm, _ := cmdUndo(m, ""); !strings.Contains(lastText(nm.(Model)), "not available") {
		t.Fatalf("undo unavailable: %q", lastText(nm.(Model)))
	}
	if nm, _ := cmdRedo(m, ""); !strings.Contains(lastText(nm.(Model)), "not available") {
		t.Fatalf("redo unavailable: %q", lastText(nm.(Model)))
	}
}
