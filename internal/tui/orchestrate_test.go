package tui

import (
	"context"
	"strings"
	"testing"
)

func TestCmdBackgroundAsync(t *testing.T) {
	var got string
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Background: func(_ context.Context, a string) string { got = a; return "started background task 20260715" },
	})
	nm, cmd := cmdBackground(m, "--exec fix the flake")
	if cmd == nil || !strings.Contains(lastText(nm.(Model)), "launching") {
		t.Fatalf("background should be async: %q", lastText(nm.(Model)))
	}
	if nt, ok := cmd().(noticeMsg); !ok || !strings.Contains(nt.text, "started background") || got != "--exec fix the flake" {
		t.Fatalf("background notice: got=%q", got)
	}
}

func TestCmdAutofixPRAsyncAndGuards(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		AutofixPR: func(_ context.Context, _ string) (string, string) {
			return "fix these checks", "found 2 failing checks on PR #42 — starting a fix"
		},
	})
	nm, cmd := cmdAutofixPR(m, "42")
	if cmd == nil || !strings.Contains(lastText(nm.(Model)), "checking the pull request") {
		t.Fatalf("autofix should be async: %q", lastText(nm.(Model)))
	}
	msg := cmd()
	af, ok := msg.(autofixMsg)
	if !ok || af.goal != "fix these checks" {
		t.Fatalf("autofix msg: %+v", msg)
	}
	// Mid-run guard.
	m.running = true
	nm, _ = cmdAutofixPR(m, "42")
	if !strings.Contains(lastText(nm.(Model)), "interrupt the current run") {
		t.Fatalf("autofix mid-run: %q", lastText(nm.(Model)))
	}
}

func TestApplyAutofixDispatchesAndSkips(t *testing.T) {
	m := New("ollama", "llama3", nil)
	// With a goal: switches to agent mode and submits the fix as a user turn.
	nm, _ := m.applyAutofix(autofixMsg{goal: "fix the lint failure", status: "found 1 failing check"})
	got := nm.(Model)
	if got.mode != "agent" {
		t.Fatalf("autofix should switch to agent mode, got %q", got.mode)
	}
	var sawGoal, sawStatus bool
	for _, e := range got.transcript {
		if strings.Contains(e.text, "fix the lint failure") {
			sawGoal = true
		}
		if strings.Contains(e.text, "found 1 failing check") {
			sawStatus = true
		}
	}
	if !sawGoal || !sawStatus {
		t.Fatalf("autofix should show status and dispatch the goal: goal=%v status=%v", sawGoal, sawStatus)
	}
	// No goal: just the status, no dispatch.
	nm, _ = m.applyAutofix(autofixMsg{goal: "", status: "CI is green on PR #42 — nothing to fix"})
	if !strings.Contains(lastText(nm.(Model)), "green") {
		t.Fatalf("green autofix: %q", lastText(nm.(Model)))
	}
}

func TestCmdBackgroundAutofixUnavailable(t *testing.T) {
	m := New("ollama", "llama3", nil)
	if nm, _ := cmdBackground(m, "x"); !strings.Contains(lastText(nm.(Model)), "not available") {
		t.Fatalf("background unavailable: %q", lastText(nm.(Model)))
	}
	if nm, _ := cmdAutofixPR(m, "1"); !strings.Contains(lastText(nm.(Model)), "not available") {
		t.Fatalf("autofix unavailable: %q", lastText(nm.(Model)))
	}
}
