package main

import (
	"context"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/tui"
)

// startAgentRun implements tui.AgentRunner: it drives a goal on a background goroutine and streams
// events to the TUI, pausing on each permission "ask" so the user can approve or deny. Agent mode
// runs interactively (state-changing actions prompt); plan mode runs strictly read-only, so it
// streams straight to its result with nothing to approve.
func (s *tuiSession) startAgentRun(goal, mode string) <-chan tui.AgentEvent {
	events := make(chan tui.AgentEvent, 1)
	go func() {
		defer close(events)
		opts := app.RunAgentOptions{
			WorkspaceRoot: s.wd, Goal: goal, Model: s.cfg.model, Provider: s.prov,
		}
		if mode == "plan" {
			opts.System = planModeSystem // read-only: no Interactive, no capability grants
		} else {
			opts.Interactive = true
			opts.Approver = &channelApprover{events: events, ctx: s.ctx}
		}
		res, err := app.RunAgent(s.ctx, opts)
		if err != nil {
			events <- tui.AgentEvent{Err: err}
			return
		}
		events <- tui.AgentEvent{Final: res.FinalText}
	}()
	return events
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
