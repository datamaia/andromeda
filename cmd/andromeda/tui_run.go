package main

import (
	"context"
	"strings"

	"github.com/datamaia/andromeda/internal/agent"
	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/tui"
)

// startAgentRun implements tui.AgentRunner: it drives a goal on a background goroutine, streaming
// content deltas and tool steps to the TUI and pausing on each permission "ask" so the user can
// approve or deny. It returns a cancel the TUI calls to interrupt the run (Esc). Agent mode runs
// interactively (state-changing actions prompt); plan mode runs strictly read-only.
func (s *tuiSession) startAgentRun(goal, mode string) (<-chan tui.AgentEvent, func()) {
	goal = s.foldPendingNotes(goal)
	events := make(chan tui.AgentEvent, 16)
	runCtx, cancel := context.WithCancel(s.ctx)
	go func() {
		defer close(events)
		defer cancel()
		opts := app.RunAgentOptions{
			WorkspaceRoot: s.wd, Goal: goal, Model: s.cfg.model, Effort: s.cfg.effort,
			History: s.history, Provider: s.prov,
			Sink: func(ev agent.RunEvent) { forwardRunEvent(events, ev) },
		}
		if mode == "plan" {
			opts.System = planModeSystem // read-only: no Interactive, no capability grants
		} else {
			opts.System = agentModeSystem // act via tools; state-changing actions prompt for approval
			opts.Interactive = true
			opts.Approver = &channelApprover{events: events, ctx: runCtx}
		}
		res, err := app.RunAgent(runCtx, opts)
		if err != nil {
			events <- tui.AgentEvent{Err: err}
			return
		}
		// Carry the conversation forward for the next turn and persist it (best-effort). Writing before
		// the terminal send establishes happens-before with the next run's read of s.history.
		s.history = res.Messages
		s.persistSession(mode)
		events <- tui.AgentEvent{Final: res.FinalText, InTokens: res.InputTokens, OutTokens: res.OutputTokens}
	}()
	return events, cancel
}

// foldPendingNotes prepends any /btw notes to the goal as a short context preamble, then clears
// them, so a queued "by the way" reaches the agent with the user's next real message.
func (s *tuiSession) foldPendingNotes(goal string) string {
	if len(s.pendingNotes) == 0 {
		return goal
	}
	var b strings.Builder
	b.WriteString("Additional context to keep in mind:\n")
	for _, n := range s.pendingNotes {
		b.WriteString("- " + n + "\n")
	}
	b.WriteString("\n")
	b.WriteString(goal)
	s.pendingNotes = nil
	return b.String()
}

// forwardRunEvent translates an agent run event into the TUI's streaming event vocabulary.
func forwardRunEvent(events chan<- tui.AgentEvent, ev agent.RunEvent) {
	switch ev.Kind {
	case "content":
		events <- tui.AgentEvent{Delta: ev.Content}
	case "tool_call":
		events <- tui.AgentEvent{Tool: &tui.ToolStep{Phase: "call", Name: ev.ToolName, Input: string(ev.ToolInput)}}
	case "tool_result":
		events <- tui.AgentEvent{Tool: &tui.ToolStep{Phase: "result", Name: ev.ToolName, Result: ev.ToolResult}}
	}
}

// channelApprover bridges the permission Manager's approval callback to the TUI: it surfaces each
// "ask" as an ApprovalRequest on the run's event stream and blocks for the user's choice, mapping
// it to the outcome and decision kind the Manager persists (session/workspace allowlist, denylist).
type channelApprover struct {
	events chan<- tui.AgentEvent
	ctx    context.Context
}

func (a *channelApprover) Approve(ctx context.Context, req ports.PermissionRequest) (core.DecisionOutcome, core.PermissionDecisionKind, error) {
	reply := make(chan tui.ApprovalDecision, 1)
	ev := tui.AgentEvent{Approval: &tui.ApprovalRequest{
		Action:  actionLabel(req.Query.Permission),
		Subject: subjectLabel(req.Query.Subject),
		Detail:  approvalDetail(req.Query),
		Reply:   reply,
	}}
	select {
	case a.events <- ev:
	case <-ctx.Done():
		return core.OutcomeDeny, core.DecisionDenyOnce, ctx.Err()
	}
	select {
	case dec := <-reply:
		outcome, kind := mapChoice(dec.Choice)
		return outcome, kind, nil
	case <-ctx.Done():
		return core.OutcomeDeny, core.DecisionDenyOnce, ctx.Err()
	}
}

// mapChoice translates a TUI approval choice into the permission outcome and persisted decision.
func mapChoice(c tui.ApprovalChoice) (core.DecisionOutcome, core.PermissionDecisionKind) {
	switch c {
	case tui.ApproveOnce:
		return core.OutcomeAllow, core.DecisionAllowOnce
	case tui.ApproveSession:
		return core.OutcomeAllow, core.DecisionAllowForSession
	case tui.ApproveWorkspace:
		return core.OutcomeAllow, core.DecisionAllowForWorkspace
	case tui.AlwaysDeny:
		return core.OutcomeDeny, core.DecisionAlwaysDeny
	default: // RejectOnce, Discuss
		return core.OutcomeDeny, core.DecisionDenyOnce
	}
}

// actionLabel renders a permission as a short imperative the user recognizes.
func actionLabel(p core.Permission) string {
	switch p {
	case core.PermWrite:
		return "write"
	case core.PermExecute:
		return "run command"
	case core.PermGitMutation:
		return "git mutation"
	case core.PermProcessSpawn:
		return "spawn process"
	case core.PermNetwork:
		return "network request"
	case core.PermCredentialAccess:
		return "read credential"
	default:
		return string(p)
	}
}

// subjectLabel keeps the resource readable when a tool leaves it empty.
func subjectLabel(s string) string {
	if s == "" {
		return "(unspecified)"
	}
	return s
}

// approvalDetail is a one-line explanation of what the agent is asking to do.
func approvalDetail(q ports.PermissionQuery) string {
	return "The agent is requesting " + actionLabel(q.Permission) + " access to " + subjectLabel(q.Subject) + "."
}
