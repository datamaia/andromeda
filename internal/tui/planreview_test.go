package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// planModeRun switches to plan mode, submits a goal, and drains the fake run to completion.
func planModeRun(t *testing.T) Model {
	t.Helper()
	events := []AgentEvent{{Final: "1. investigate\n2. change X\n3. verify"}}
	var m tea.Model = New("ollama", "llama3", nil).WithAgentRunner(streamRunner(events, nil))
	m = typeString(m, "/plan")
	m, _ = m.Update(key(tea.KeyEnter))
	m = typeString(m, "add a feature")
	m, cmd := m.Update(key(tea.KeyEnter))
	m = drain(m, cmd)
	return m.(Model)
}

// A completed plan-mode turn opens the approve/refine/reject overlay.
func TestPlanReviewOpens(t *testing.T) {
	got := planModeRun(t)
	if !got.planReview {
		t.Fatal("plan review overlay should open after a plan-mode turn completes")
	}
	if got.mode != "plan" {
		t.Errorf("mode should still be plan while reviewing, got %q", got.mode)
	}
}

// Approving switches to agent mode and submits the build prompt (a new run starts).
func TestPlanReviewApproveSwitchesToAgent(t *testing.T) {
	m := tea.Model(planModeRun(t))
	m, cmd := m.Update(key(tea.KeyEnter)) // cursor 0 = Approve & execute
	got := m.(Model)
	if got.planReview {
		t.Error("overlay should close after a choice")
	}
	if got.mode != "agent" {
		t.Errorf("approve should switch to agent mode, got %q", got.mode)
	}
	if !got.running && cmd == nil {
		t.Error("approve should kick off a new agent run")
	}
	_ = drain(m, cmd)
}

// Rejecting keeps plan mode and closes the overlay.
func TestPlanReviewReject(t *testing.T) {
	m := tea.Model(planModeRun(t))
	m, _ = m.Update(key(tea.KeyDown)) // Refine
	m, _ = m.Update(key(tea.KeyDown)) // Reject
	m, _ = m.Update(key(tea.KeyEnter))
	got := m.(Model)
	if got.planReview {
		t.Error("overlay should close after reject")
	}
	if got.mode != "plan" {
		t.Errorf("reject should stay in plan mode, got %q", got.mode)
	}
}

// Esc keeps planning without executing.
func TestPlanReviewEscKeepsPlanning(t *testing.T) {
	m := tea.Model(planModeRun(t))
	m, _ = m.Update(key(tea.KeyEscape))
	got := m.(Model)
	if got.planReview {
		t.Error("esc should close the overlay")
	}
	if got.mode != "plan" {
		t.Errorf("esc should stay in plan mode, got %q", got.mode)
	}
}
