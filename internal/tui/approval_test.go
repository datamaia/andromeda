package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// fakeRunner emits one approval request, then a final message naming the choice the user made.
func fakeRunner(captured *ApprovalChoice) AgentRunner {
	return func(_, _ string) (<-chan AgentEvent, func()) {
		ch := make(chan AgentEvent, 1)
		reply := make(chan ApprovalDecision, 1)
		go func() {
			ch <- AgentEvent{Approval: &ApprovalRequest{Action: "write", Subject: "/tmp/x", Reply: reply}}
			dec := <-reply
			if captured != nil {
				*captured = dec.Choice
			}
			ch <- AgentEvent{Final: "run complete"}
			close(ch)
		}()
		return ch, func() {}
	}
}

// A submitted goal that triggers a state-changing action pauses on an approval overlay; choosing an
// answer delivers it to the runner and the run resumes to completion.
func TestApprovalOverlayApprove(t *testing.T) {
	var chosen ApprovalChoice = -1
	var m tea.Model = New("ollama", "llama3", nil).WithAgentRunner(fakeRunner(&chosen))

	m = typeString(m, "edit the file")
	m, cmd := m.Update(key(tea.KeyEnter))
	if !m.(Model).running {
		t.Fatal("submitting a goal should start a run")
	}
	// drain the first event (the approval pause)
	m, _ = m.Update(cmd())
	got := m.(Model)
	if got.approval == nil {
		t.Fatal("expected the approval overlay to open")
	}
	if !strings.Contains(got.View().Content, "Permission required") {
		t.Error("approval overlay should render the permission prompt")
	}

	// choose the highlighted answer (Approve once) and resume
	m, cmd = m.Update(key(tea.KeyEnter))
	if m.(Model).approval != nil {
		t.Error("overlay should close after answering")
	}
	m, _ = m.Update(cmd()) // deliver the final event
	got = m.(Model)
	if chosen != ApproveOnce {
		t.Errorf("runner received choice %d, want ApproveOnce", chosen)
	}
	if got.running {
		t.Error("run should have finished")
	}
	tr := got.Transcript()
	if !strings.Contains(tr[len(tr)-1], "run complete") {
		t.Errorf("transcript missing the run result: %v", tr)
	}
}

// Navigating to "Always deny" delivers that choice (the denylist path).
func TestApprovalOverlayAlwaysDeny(t *testing.T) {
	var chosen ApprovalChoice = -1
	var m tea.Model = New("ollama", "llama3", nil).WithAgentRunner(fakeRunner(&chosen))
	m = typeString(m, "rm everything")
	m, cmd := m.Update(key(tea.KeyEnter))
	m, _ = m.Update(cmd()) // approval pause
	// Always deny is index 4 (Approve once/session/workspace, Reject, Always deny)
	for i := 0; i < 4; i++ {
		m, _ = m.Update(key(tea.KeyDown))
	}
	m, cmd = m.Update(key(tea.KeyEnter))
	_, _ = m.Update(cmd())
	if chosen != AlwaysDeny {
		t.Errorf("runner received choice %d, want AlwaysDeny", chosen)
	}
}

// Esc on the approval overlay rejects the action (safe default) rather than quitting.
func TestApprovalEscRejects(t *testing.T) {
	var chosen ApprovalChoice = -1
	var m tea.Model = New("ollama", "llama3", nil).WithAgentRunner(fakeRunner(&chosen))
	m = typeString(m, "touch a file")
	m, cmd := m.Update(key(tea.KeyEnter))
	m, _ = m.Update(cmd()) // approval pause
	m, cmd = m.Update(key(tea.KeyEscape))
	if m.(Model).quitting {
		t.Error("esc on the approval overlay must not quit")
	}
	_, _ = m.Update(cmd())
	if chosen != RejectOnce {
		t.Errorf("esc should reject once, got choice %d", chosen)
	}
}

// While a run is in flight, submitting another goal is ignored.
func TestSubmitIgnoredWhileRunning(t *testing.T) {
	var m tea.Model = New("ollama", "llama3", nil).WithAgentRunner(fakeRunner(nil))
	m = typeString(m, "first")
	m, _ = m.Update(key(tea.KeyEnter)) // starts a run (now awaiting the first event)
	before := len(m.(Model).Transcript())
	m = typeString(m, "second")
	m, _ = m.Update(key(tea.KeyEnter)) // should be ignored while running
	if got := len(m.(Model).Transcript()); got != before {
		t.Errorf("submitting while running added a transcript entry (%d → %d)", before, got)
	}
}
