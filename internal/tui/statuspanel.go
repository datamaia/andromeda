package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// The /status panel is a tabbed session dashboard (Overview | Usage | Tools | Context). It replaces
// the transcript while open; ←/→ or Tab switch tabs, number keys jump, Esc/q close. statusTabNames
// (shared with the palette's argument completion) is the tab order.

// openStatusPanel shows the dashboard, optionally jumping to a named tab (e.g. /status usage).
func (m Model) openStatusPanel(tabArg string) Model {
	m.statusPanel = true
	m.statusTab = 0
	if tabArg != "" {
		for i, name := range statusTabNames {
			if name == strings.ToLower(strings.TrimSpace(tabArg)) {
				m.statusTab = i
				break
			}
		}
	}
	return m
}

// handleStatusKey drives the panel: ←/→ or Shift+Tab/Tab move, 1-4 jump, Esc/q close.
func (m Model) handleStatusKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	n := len(statusTabNames)
	switch {
	case msg.Code == tea.KeyEscape || msg.Text == "q":
		m.statusPanel = false
		return m, nil
	case msg.Code == tea.KeyLeft || (msg.Code == tea.KeyTab && msg.Mod&tea.ModShift != 0):
		m.statusTab = (m.statusTab - 1 + n) % n
	case msg.Code == tea.KeyRight || msg.Code == tea.KeyTab:
		m.statusTab = (m.statusTab + 1) % n
	case msg.Text >= "1" && msg.Text <= "4":
		if i := int(msg.Text[0] - '1'); i < n {
			m.statusTab = i
		}
	}
	return m, nil
}

// renderStatusPanel draws the tab strip and the active tab's body.
func (m Model) renderStatusPanel() string {
	var b strings.Builder
	b.WriteString("\n  " + m.styles.Title.Render("Status") + "\n\n  ")
	for i, name := range statusTabNames {
		label := " " + capitalize(name) + " "
		if i == m.statusTab {
			b.WriteString(m.styles.StatusBar.Render(label))
		} else {
			b.WriteString(m.styles.Muted.Render(label))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n\n")
	b.WriteString(m.statusTabBody())
	b.WriteString("\n  " + m.styles.Muted.Render("←/→ or Tab switch · 1-4 jump · esc close") + "\n")
	return b.String()
}

// statusTabBody renders the body of the active tab.
func (m Model) statusTabBody() string {
	switch statusTabNames[m.statusTab] {
	case "usage":
		return m.statusUsageBody()
	case "tools":
		return m.statusToolsBody()
	case "context":
		return m.statusContextBody()
	default:
		return m.statusOverviewBody()
	}
}

func (m Model) kv(key, val string) string {
	return "  " + m.styles.Muted.Render(fmt.Sprintf("%-12s", key)) + val + "\n"
}

func (m Model) statusOverviewBody() string {
	effort := m.effort
	if effort == "" {
		effort = "default"
	}
	loop := "off"
	if m.loop {
		loop = "on"
	}
	var b strings.Builder
	b.WriteString(m.kv("provider", m.provider))
	b.WriteString(m.kv("model", m.model))
	b.WriteString(m.kv("mode", m.modeOrDefault()))
	b.WriteString(m.kv("effort", effort))
	b.WriteString(m.kv("theme", m.themeName()))
	b.WriteString(m.kv("loop", loop))
	b.WriteString(m.kv("uptime", m.uptime()))
	b.WriteString(m.kv("state", m.state))
	return b.String()
}

func (m Model) statusUsageBody() string {
	var b strings.Builder
	b.WriteString(m.kv("input tok", humanCount(m.inTokens)))
	b.WriteString(m.kv("output tok", humanCount(m.outTokens)))
	b.WriteString(m.kv("total tok", humanCount(m.inTokens+m.outTokens)))
	b.WriteString(m.kv("turns", fmt.Sprintf("%d", m.turnCount())))
	if m.inTokens == 0 && m.outTokens == 0 {
		b.WriteString("\n  " + m.styles.Muted.Render("(token usage appears after the first run that reports it)") + "\n")
	}
	return b.String()
}

// statusToolsBody summarizes what the agent may do in the active mode (the concrete grants are
// enforced by the permission prompt).
func (m Model) statusToolsBody() string {
	var lines []string
	switch m.modeOrDefault() {
	case "plan":
		lines = []string{"read-only — no changes are made", "read · search · diff"}
	case "shell":
		lines = []string{"runs your shell commands directly (not agent tools)"}
	default:
		lines = []string{
			"read · search · diff",
			"write · edit · patch   (approval-gated)",
			"git · shell · process  (approval-gated)",
			"http request           (approval-gated)",
		}
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString("  " + m.styles.Muted.Render(l) + "\n")
	}
	return b.String()
}

func (m Model) statusContextBody() string {
	var b strings.Builder
	if m.actions.Context != nil {
		for _, l := range m.actions.Context(context.Background()) {
			b.WriteString("  " + l + "\n")
		}
		return b.String()
	}
	if m.workspaceRoot != "" {
		b.WriteString(m.kv("workspace", m.workspaceRoot))
	}
	if m.branch != "" {
		b.WriteString(m.kv("branch", m.branch))
	}
	if b.Len() == 0 {
		b.WriteString("  " + m.styles.Muted.Render("(workspace context unavailable)") + "\n")
	}
	return b.String()
}

// capitalize upper-cases the first letter of a tab name for the tab strip.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// turnCount is the number of user turns in the transcript.
func (m Model) turnCount() int {
	n := 0
	for _, e := range m.transcript {
		if e.role == "user" {
			n++
		}
	}
	return n
}
