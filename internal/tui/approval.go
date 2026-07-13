package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// AgentRunner starts an agent run for a submitted line and returns the stream of events the Model
// consumes. It is injected by the driver so the TUI keeps no agent/permission imports. A run emits
// zero or more ApprovalRequest pauses, then exactly one terminal event (Final or Err); the channel
// is closed afterwards. When no runner is wired the Model falls back to the synchronous Responder.
type AgentRunner func(goal, mode string) <-chan AgentEvent

// AgentEvent is one step of a running agent: a pause awaiting approval, or the terminal result.
type AgentEvent struct {
	Approval *ApprovalRequest // non-nil: the run is paused until the user answers
	Final    string           // the run's final text (terminal, when Approval == nil && Err == nil)
	Err      error            // the run failed (terminal)
}

// ApprovalRequest describes a state-changing action awaiting the user's decision. The Model sends
// the choice on Reply exactly once; the driver's approver blocks until it does. Reply is buffered
// so the send from the UI never stalls.
type ApprovalRequest struct {
	Action  string // "write" | "execute" | "git mutation" | "network" | …
	Subject string // the concrete resource: path, command, host, repository
	Detail  string // one-line human explanation
	Reply   chan ApprovalDecision
}

// ApprovalDecision is the user's answer to an ApprovalRequest.
type ApprovalDecision struct{ Choice ApprovalChoice }

// ApprovalChoice is one of the standard permission answers (Volume 9 decision kinds). The driver
// maps each to a persisted grant (session/workspace) or a one-off allow/deny.
type ApprovalChoice int

const (
	// ApproveOnce allows this single action.
	ApproveOnce ApprovalChoice = iota
	// ApproveSession allows this action for the rest of the session (session allowlist).
	ApproveSession
	// ApproveWorkspace allows this action in this workspace, persisted (workspace allowlist).
	ApproveWorkspace
	// RejectOnce denies this single action.
	RejectOnce
	// AlwaysDeny denies and persists a standing deny (denylist).
	AlwaysDeny
	// Discuss denies this action so the user can talk it over before retrying.
	Discuss
)

// approvalOption pairs a choice with its menu label.
type approvalOption struct {
	choice ApprovalChoice
	label  string
}

// approvalChoices is the ordered set of answers shown in the approval overlay.
func approvalChoices() []approvalOption {
	return []approvalOption{
		{ApproveOnce, "Approve once"},
		{ApproveSession, "Approve for this session"},
		{ApproveWorkspace, "Approve for this workspace (persist)"},
		{RejectOnce, "Reject"},
		{AlwaysDeny, "Always deny (add to denylist)"},
		{Discuss, "Reject & discuss"},
	}
}

// WithAgentRunner wires the async, approval-capable agent runner. Without it, agent/plan replies go
// through the synchronous Responder and no approval prompts are raised.
func (m Model) WithAgentRunner(r AgentRunner) Model {
	m.runner = r
	return m
}

// agentEventMsg carries one AgentEvent (or a closed-channel signal) back into Update.
type agentEventMsg struct {
	ev     AgentEvent
	closed bool
}

// waitAgent returns a command that reads the next event from a run's stream.
func waitAgent(ch <-chan AgentEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return agentEventMsg{closed: true}
		}
		return agentEventMsg{ev: ev}
	}
}

// handleAgentEvent advances a running agent: it opens the approval overlay on a pause, or finishes
// the run on a terminal event.
func (m Model) handleAgentEvent(msg agentEventMsg) (tea.Model, tea.Cmd) {
	if msg.closed {
		// Stream ended without a separate terminal event; settle to ready.
		m.running = false
		m.agentEvents = nil
		m.state = "ready"
		return m, nil
	}
	if msg.ev.Approval != nil {
		m.approval = msg.ev.Approval
		m.approvalCursor = 0
		m.state = "awaiting approval"
		return m, nil // wait for the user; do not re-subscribe yet
	}
	// terminal event: append the result and settle
	m.running = false
	m.agentEvents = nil
	m.state = "ready"
	if msg.ev.Err != nil {
		m.transcript = append(m.transcript, entry{"agent", "error: " + msg.ev.Err.Error()})
	} else {
		m.transcript = append(m.transcript, entry{"agent", msg.ev.Final})
	}
	return m, nil
}

// handleApprovalKey drives the approval overlay: arrows move, enter chooses, esc rejects.
func (m Model) handleApprovalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	opts := approvalChoices()
	switch {
	case msg.Code == tea.KeyEscape:
		return m.answerApproval(RejectOnce)
	case msg.Code == tea.KeyUp || msg.Text == "k":
		if m.approvalCursor > 0 {
			m.approvalCursor--
		}
	case msg.Code == tea.KeyDown || msg.Text == "j":
		if m.approvalCursor < len(opts)-1 {
			m.approvalCursor++
		}
	case msg.Code == tea.KeyEnter:
		return m.answerApproval(opts[clamp(m.approvalCursor, len(opts))].choice)
	}
	return m, nil
}

// answerApproval delivers the user's choice to the blocked approver and resumes the run.
func (m Model) answerApproval(choice ApprovalChoice) (tea.Model, tea.Cmd) {
	ap := m.approval
	m.approval = nil
	m.state = "running"
	if ap != nil && ap.Reply != nil {
		ap.Reply <- ApprovalDecision{Choice: choice} // buffered: never blocks the UI
	}
	return m, waitAgent(m.agentEvents)
}

// renderApproval draws the permission overlay with the action, resource, and answer choices.
func (m Model) renderApproval() string {
	ap := m.approval
	var b strings.Builder
	b.WriteString("\n  " + m.styles.Title.Render("Permission required") + "\n\n")
	b.WriteString("  " + m.styles.User.Render(ap.Action) + "  " + ap.Subject + "\n")
	if ap.Detail != "" {
		b.WriteString("  " + m.styles.Muted.Render(ap.Detail) + "\n")
	}
	b.WriteString("\n")
	for i, opt := range approvalChoices() {
		if i == m.approvalCursor {
			b.WriteString("  " + m.styles.Agent.Render("▸ "+opt.label) + "\n")
		} else {
			b.WriteString("    " + m.styles.Muted.Render(opt.label) + "\n")
		}
	}
	b.WriteString("\n  " + m.styles.Muted.Render("↑/↓ move · enter choose · esc reject") + "\n")
	return b.String()
}
