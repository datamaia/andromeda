package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// cmdBackground (/background) launches an unattended agent for a goal in a detached process. The
// driver does the spawning; it may touch the filesystem/network, so it runs off the UI thread.
func cmdBackground(m Model, args string) (tea.Model, tea.Cmd) {
	if m.actions.Background == nil {
		return m.unavailable("background"), nil
	}
	a := strings.TrimSpace(args)
	fn := m.actions.Background
	return m.sys("launching a background task…"), func() tea.Msg {
		return noticeMsg{text: fn(context.Background(), a)}
	}
}

// autofixMsg carries the result of an /autofix-pr CI inspection: a fix goal to dispatch (when
// non-empty) and a status line to show.
type autofixMsg struct {
	goal   string
	status string
}

// cmdAutofixPR (/autofix-pr) inspects a PR's CI off the UI thread; the resulting autofixMsg either
// starts a fix run or just reports that there is nothing to fix.
func cmdAutofixPR(m Model, args string) (tea.Model, tea.Cmd) {
	if m.actions.AutofixPR == nil {
		return m.unavailable("autofix-pr"), nil
	}
	if m.running {
		return m.sys("finish or interrupt the current run before starting an autofix"), nil
	}
	a := strings.TrimSpace(args)
	fn := m.actions.AutofixPR
	return m.sys("checking the pull request's CI…"), func() tea.Msg {
		goal, status := fn(context.Background(), a)
		return autofixMsg{goal: goal, status: status}
	}
}

// applyAutofix shows the CI status and, when a fix goal came back, switches to agent mode and
// dispatches it as a normal run (so the fix streams and prompts for approval like any other turn).
func (m Model) applyAutofix(msg autofixMsg) (tea.Model, tea.Cmd) {
	m = m.sys(msg.status)
	if strings.TrimSpace(msg.goal) == "" || m.running {
		return m, nil
	}
	m.mode = "agent"
	m.transcript = append(m.transcript, entry{"user", msg.goal})
	return m.dispatchGoal(msg.goal, "agent")
}
