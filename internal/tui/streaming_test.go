package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// streamRunner emits the given events in order, then closes. Returns a cancel that records the call.
func streamRunner(events []AgentEvent, cancelled *bool) AgentRunner {
	return func(_, _ string) (<-chan AgentEvent, func()) {
		ch := make(chan AgentEvent, len(events))
		for _, e := range events {
			ch <- e
		}
		close(ch)
		return ch, func() {
			if cancelled != nil {
				*cancelled = true
			}
		}
	}
}

// drain repeatedly runs the pending command and feeds its message back until there is no command.
func drain(m tea.Model, cmd tea.Cmd) tea.Model {
	for cmd != nil {
		msg := cmd()
		m, cmd = m.Update(msg)
	}
	return m
}

// Streamed content deltas accumulate into a single agent line as they arrive.
func TestStreamingDeltasAccumulate(t *testing.T) {
	events := []AgentEvent{{Delta: "Hola"}, {Delta: ", "}, {Delta: "mundo"}, {Final: "Hola, mundo"}}
	var m tea.Model = New("groq", "x", nil).WithAgentRunner(streamRunner(events, nil))
	m = typeString(m, "saluda")
	m, cmd := m.Update(key(tea.KeyEnter))
	m = drain(m, cmd)
	got := m.(Model)
	if got.running {
		t.Fatal("run should have finished")
	}
	tr := got.Transcript()
	if !strings.Contains(tr[len(tr)-1], "Hola, mundo") {
		t.Fatalf("streamed line should read the final text, got %v", tr)
	}
	// The streamed deltas produced exactly one agent line (not four).
	agentLines := 0
	for _, l := range tr {
		if strings.HasPrefix(l, "agent:") {
			agentLines++
		}
	}
	if agentLines != 1 {
		t.Fatalf("expected 1 agent line from the stream, got %d: %v", agentLines, tr)
	}
}

// A tool step renders a subtle log: a human action line, then the result folded under a "⎿".
func TestStreamingToolCard(t *testing.T) {
	events := []AgentEvent{
		{Tool: &ToolStep{Phase: "call", Name: "fs_write", Input: `{"path":"a.md"}`}},
		{Tool: &ToolStep{Phase: "result", Name: "fs_write", Result: "wrote a.md"}},
		{Final: "done"},
	}
	var m tea.Model = New("groq", "x", nil).WithAgentRunner(streamRunner(events, nil))
	m = typeString(m, "crea a.md")
	m, cmd := m.Update(key(tea.KeyEnter))
	m = drain(m, cmd)
	view := m.(Model).View().Content
	// The raw tool name is hidden behind a human verb + subject, and the result folds under "⎿".
	if !strings.Contains(view, "Write a.md") || !strings.Contains(view, "⎿") || !strings.Contains(view, "wrote a.md") {
		t.Fatalf("tool log should show the action and its result:\n%s", view)
	}
}

// Esc during a run calls the run's cancel (interrupt) rather than clearing the input.
func TestEscInterruptsRun(t *testing.T) {
	var cancelled bool
	// A run that only pauses (approval) so it's still "running" when we press Esc.
	reply := make(chan ApprovalDecision, 1)
	runner := func(_, _ string) (<-chan AgentEvent, func()) {
		ch := make(chan AgentEvent, 1)
		ch <- AgentEvent{Approval: &ApprovalRequest{Action: "write", Subject: "/x", Reply: reply}}
		return ch, func() { cancelled = true }
	}
	var m tea.Model = New("groq", "x", nil).WithAgentRunner(runner)
	m = typeString(m, "do it")
	m, cmd := m.Update(key(tea.KeyEnter))
	m, _ = m.Update(cmd()) // read the approval pause
	// Answer to close the overlay but keep the run "running", then Esc to interrupt.
	m, _ = m.Update(key(tea.KeyEnter)) // approve once → resumes (running stays true)
	m2, _ := m.Update(key(tea.KeyEscape))
	if !cancelled {
		t.Fatal("esc during a run should call the run's cancel")
	}
	if !m2.(Model).interrupted {
		t.Error("esc should mark the run interrupted")
	}
}
