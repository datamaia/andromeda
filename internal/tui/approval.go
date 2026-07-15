package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// AgentRunner starts an agent run for a submitted line and returns the stream of events the Model
// consumes plus a cancel func the Model calls to interrupt the run (Esc). It is injected by the
// driver so the TUI keeps no agent/permission imports. A run emits streamed Delta/Tool events and
// zero or more ApprovalRequest pauses, then exactly one terminal event (Final or Err); the channel
// is closed afterwards. When no runner is wired the Model falls back to the synchronous Responder.
type AgentRunner func(goal, mode string) (<-chan AgentEvent, func())

// AgentEvent is one step of a running agent: a streamed content delta, a tool step, a pause awaiting
// approval, or the terminal result.
type AgentEvent struct {
	Delta     string           // a streamed chunk of assistant text (non-terminal)
	Tool      *ToolStep        // a tool call starting or its result (non-terminal)
	Approval  *ApprovalRequest // non-nil: the run is paused until the user answers
	Notice    string           // an out-of-band system note during a run (e.g. auto-compaction); non-terminal
	Final     string           // the run's final text (terminal, when the others are unset)
	Err       error            // the run failed (terminal)
	InTokens  int              // input tokens consumed by the run (reported on the terminal event)
	OutTokens int              // output tokens produced by the run
}

// ToolStep is a tool call starting ("call") or completing ("result"), rendered inline as a card.
type ToolStep struct {
	Phase  string // "call" | "result"
	Name   string
	Input  string
	Result string
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

// handleAgentEvent advances a running agent: it streams content into the live agent line, renders
// tool steps as inline cards, opens the approval overlay on a pause, or finishes on a terminal event.
func (m Model) handleAgentEvent(msg agentEventMsg) (tea.Model, tea.Cmd) {
	if msg.closed {
		// Stream ended without a separate terminal event; settle to ready.
		return m.settleRun(), nil
	}
	switch {
	case msg.ev.Delta != "":
		m = m.appendDelta(msg.ev.Delta)
		return m, waitAgent(m.agentEvents)
	case msg.ev.Tool != nil:
		m = m.appendToolStep(msg.ev.Tool)
		return m, waitAgent(m.agentEvents)
	case msg.ev.Approval != nil:
		m.approval = msg.ev.Approval
		m.approvalCursor = 0
		m.state = "awaiting approval"
		return m, nil // wait for the user; do not re-subscribe yet
	case msg.ev.Notice != "":
		m.transcript = append(m.transcript, entry{"system", msg.ev.Notice})
		return m, waitAgent(m.agentEvents)
	}
	// terminal event: canonicalize the streamed line (or append the result), then settle.
	if msg.ev.Err != nil {
		if m.interrupted {
			m.transcript = append(m.transcript, entry{"system", "interrupted"})
		} else {
			m.transcript = append(m.transcript, entry{"agent", "error: " + msg.ev.Err.Error()})
		}
	} else if msg.ev.Final != "" {
		if m.streamIdx >= 0 && m.streamIdx < len(m.transcript) {
			m.transcript[m.streamIdx].text = msg.ev.Final // replace streamed text with the canonical final
		} else {
			m.transcript = append(m.transcript, entry{"agent", msg.ev.Final})
		}
	}
	// Accumulate reported token usage for the /status Usage tab and the status bar.
	m.inTokens += msg.ev.InTokens
	m.outTokens += msg.ev.OutTokens
	// A plan-mode turn that finished cleanly hands off to the approve/refine/reject overlay.
	completed := msg.ev.Err == nil && !m.interrupted && msg.ev.Final != ""
	planMode := m.modeOrDefault() == "plan"
	m = m.settleRun()
	if completed && planMode {
		m = m.openPlanReview()
	}
	return m, nil
}

// appendDelta streams a chunk of assistant text into the current agent line, starting a new line
// when none is open (e.g. the first delta, or the first delta after a tool step).
func (m Model) appendDelta(delta string) Model {
	if m.streamIdx < 0 || m.streamIdx >= len(m.transcript) {
		m.transcript = append(m.transcript, entry{"agent", ""})
		m.streamIdx = len(m.transcript) - 1
	}
	m.transcript[m.streamIdx].text += delta
	m.state = "streaming"
	return m
}

// appendToolStep renders a tool call as a subtle two-line log (a human action line, then the result
// folded under a "⎿" connector), in the style of Claude Code. A tool step closes the current
// streamed line so text after it starts fresh.
func (m Model) appendToolStep(t *ToolStep) Model {
	switch t.Phase {
	case "result":
		if m.toolIdx >= 0 && m.toolIdx < len(m.transcript) {
			s := toolResultSummary(t.Result)
			if m.showDetails {
				s = toolResultDetail(t.Result)
			}
			if s != "" {
				m.transcript[m.toolIdx].text += "\n  ⎿  " + s
			}
		}
		m.state = "thinking" // the tool is done; the model now reasons about the result
	default: // "call"
		label := toolCallLabel(t.Name, t.Input)
		if m.showDetails {
			if in := strings.TrimSpace(t.Input); in != "" && in != "{}" {
				label += "\n     ↳ " + oneLine(t.Input, 200) // full arguments in details mode
			}
		}
		m.transcript = append(m.transcript, entry{"tool", label})
		m.toolIdx = len(m.transcript) - 1
		m.state = "working" // a tool is executing
	}
	m.streamIdx = -1 // the next content delta starts a new agent line below the log
	return m
}

// toolVerbs maps built-in tool names to a human action verb for the log line.
var toolVerbs = map[string]string{
	"fs_read": "Read", "fs_search": "Search", "fs_diff": "Diff",
	"fs_write": "Write", "fs_replace": "Edit", "fs_patch": "Edit",
	"git_exec": "Git", "terminal_run": "Run", "process_control": "Process",
	"http_request": "Fetch", "sqlite_query": "Query",
}

// toolCallLabel renders a tool call as "<Verb> <subject>" (e.g. "Read internal/tui/model.go"),
// falling back to the raw tool name when the verb or subject is unknown.
func toolCallLabel(name, input string) string {
	verb := toolVerbs[name]
	if verb == "" {
		verb = name
	}
	if subj := toolSubject(input); subj != "" {
		return verb + " " + subj
	}
	return verb
}

// toolSubject pulls the most meaningful argument (path, command, query, url…) out of a tool's JSON
// input to display alongside the verb.
func toolSubject(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		return oneLine(input, 60)
	}
	for _, k := range []string{"command", "cmd", "url", "query", "pattern", "statement", "path", "file", "filename"} {
		if s, ok := m[k].(string); ok && strings.TrimSpace(s) != "" {
			return oneLine(s, 60)
		}
	}
	return ""
}

// toolResultSummary condenses a tool result to one subtle line: a friendly field from a JSON result
// (content/output/stdout…) or the raw text, with a "(+N lines)" hint when it spans several lines.
func toolResultSummary(result string) string {
	r := strings.TrimSpace(result)
	if r == "" {
		return ""
	}
	var m map[string]any
	if json.Unmarshal([]byte(r), &m) == nil {
		for _, k := range []string{"content", "output", "stdout", "result", "text", "message", "error"} {
			if s, ok := m[k].(string); ok {
				return summarizeText(s)
			}
		}
	}
	return summarizeText(r)
}

// toolResultDetail returns the tool's result content with a longer excerpt than toolResultSummary,
// for /details mode: it unwraps the same JSON fields but keeps up to 300 collapsed characters so the
// user can inspect the output without opening a file.
func toolResultDetail(result string) string {
	r := strings.TrimSpace(result)
	if r == "" {
		return ""
	}
	var m map[string]any
	if json.Unmarshal([]byte(r), &m) == nil {
		for _, k := range []string{"content", "output", "stdout", "result", "text", "message", "error"} {
			if s, ok := m[k].(string); ok {
				r = s
				break
			}
		}
	}
	return oneLine(r, 300)
}

// summarizeText returns the first line of s, appending "(+N lines)" when more follow.
func summarizeText(s string) string {
	s = strings.TrimRight(s, "\n")
	if strings.TrimSpace(s) == "" {
		return "(empty)"
	}
	lines := strings.Count(s, "\n") + 1
	first := s
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		first = s[:i]
	}
	first = oneLine(first, 70)
	if lines > 1 {
		return fmt.Sprintf("%s  (+%d lines)", first, lines-1)
	}
	return first
}

// settleRun resets streaming state when a run ends.
func (m Model) settleRun() Model {
	m.running = false
	m.agentEvents = nil
	m.cancelRun = nil
	m.interrupted = false
	m.streamIdx = -1
	m.toolIdx = -1
	m.state = "ready"
	return m
}

// oneLine collapses whitespace/newlines and truncates, so a tool arg/result stays a single row.
func oneLine(s string, limit int) string {
	s = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\r", ""))
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	if len(s) > limit {
		return s[:limit-1] + "…"
	}
	return s
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
