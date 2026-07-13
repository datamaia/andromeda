package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// The plan-review handoff closes the loop between plan mode and agent mode: when a plan-mode turn
// finishes, an overlay offers to approve the proposal (switch to agent mode and execute it), refine
// it, or reject it. This mirrors how other agent CLIs turn a plan into action without the user
// having to remember to type /agent.

// planReviewKind is the chosen disposition of a proposed plan.
type planReviewKind int

const (
	planApprove planReviewKind = iota
	planRefine
	planReject
)

type planReviewOption struct {
	kind  planReviewKind
	label string
}

var planReviewOptions = []planReviewOption{
	{planApprove, "Approve & execute (switch to agent mode)"},
	{planRefine, "Refine the plan (stay in plan mode)"},
	{planReject, "Reject"},
}

// planBuildPrompt is sent when a plan is approved, so the agent carries out the plan it just wrote.
const planBuildPrompt = "The plan above is approved. Execute it now, making the necessary changes."

// openPlanReview shows the approve/refine/reject overlay after a completed plan-mode turn.
func (m Model) openPlanReview() Model {
	m.planReview = true
	m.planReviewCursor = 0
	return m
}

// handlePlanReviewKey drives the overlay: arrows move, enter chooses, esc keeps planning.
func (m Model) handlePlanReviewKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Code == tea.KeyEscape:
		m.planReview = false
		return m.sys("keeping the plan open — refine your request (still in plan mode)"), nil
	case msg.Code == tea.KeyUp || msg.Text == "k":
		if m.planReviewCursor > 0 {
			m.planReviewCursor--
		}
		return m, nil
	case msg.Code == tea.KeyDown || msg.Text == "j":
		if m.planReviewCursor < len(planReviewOptions)-1 {
			m.planReviewCursor++
		}
		return m, nil
	case msg.Code == tea.KeyEnter:
		return m.choosePlanReview(planReviewOptions[clamp(m.planReviewCursor, len(planReviewOptions))].kind)
	}
	return m, nil
}

// choosePlanReview applies the disposition: approve switches to agent mode and submits the build
// prompt; refine/reject leave a note and stay in plan mode.
func (m Model) choosePlanReview(kind planReviewKind) (tea.Model, tea.Cmd) {
	m.planReview = false
	switch kind {
	case planApprove:
		m.mode = "agent"
		m = m.sys("plan approved → executing in agent mode")
		m.input = planBuildPrompt
		return m.submit()
	case planRefine:
		return m.sys("refine the plan — tell me what to change (still in plan mode)"), nil
	default: // planReject
		return m.sys("plan rejected — still in plan mode"), nil
	}
}

// renderPlanReview draws the overlay above the prompt, keeping the proposed plan visible above it.
func (m Model) renderPlanReview() string {
	var b strings.Builder
	b.WriteString("  " + m.styles.Title.Render("Plan ready") +
		m.styles.Muted.Render(" — review the proposal above") + "\n")
	for i, opt := range planReviewOptions {
		if i == m.planReviewCursor {
			b.WriteString("  " + m.styles.Agent.Render("▸ "+opt.label) + "\n")
		} else {
			b.WriteString("    " + m.styles.Muted.Render(opt.label) + "\n")
		}
	}
	b.WriteString("  " + m.styles.Muted.Render("↑/↓ move · enter choose · esc keep planning") + "\n")
	return b.String()
}
