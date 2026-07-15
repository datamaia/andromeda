package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// cmdDetails (/details) toggles verbose tool logging for the session. It is pure display state, so
// it needs no driver action. Grammar: on | off | (bare toggles).
func cmdDetails(m Model, args string) (tea.Model, tea.Cmd) {
	switch strings.TrimSpace(args) {
	case "on", "show":
		m.showDetails = true
	case "off", "hide":
		m.showDetails = false
	case "", "toggle":
		m.showDetails = !m.showDetails
	default:
		return m.sys("usage: /details [on | off]"), nil
	}
	if m.showDetails {
		return m.sys("tool details ON — each step shows its full arguments and a longer result excerpt"), nil
	}
	return m.sys("tool details OFF — tool steps show a compact one-line summary"), nil
}

// cmdEditor (/editor) opens $EDITOR to compose a prompt; on save the composed text is sent as a
// goal. The driver supplies the suspend-and-run command (tea.ExecProcess), so the TUI keeps no
// process/filesystem imports; the composer's current text seeds the editor buffer.
func cmdEditor(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.Editor == nil {
		return m.unavailable("editor"), nil
	}
	return m, m.actions.Editor(m.input)
}

// EditorMsg carries the result of an external $EDITOR session back to the Model.
type EditorMsg struct {
	Text string
	Err  error
}

// applyEditor handles the composed text: it reports an error, ignores an empty buffer, or submits
// the composed prompt as a goal (reusing the normal submit path, so skills and history apply).
func (m Model) applyEditor(msg EditorMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		return m.sys("editor: " + msg.Err.Error()), nil
	}
	if strings.TrimSpace(msg.Text) == "" {
		return m.sys("editor: nothing to send"), nil
	}
	if m.running {
		return m.sys("editor: finish the current run first"), nil
	}
	m.input = msg.Text
	return m.submit()
}
