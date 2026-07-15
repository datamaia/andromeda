package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// /add-dir and /cd manage the session's working directories. Both delegate to the driver (which owns
// the filesystem) through Actions and then refresh the TUI's cached file/skill lists so @-mention and
// $-invocation reflect the change immediately.

// cmdAddDir adds an extra working directory whose files join @-mention completion.
func cmdAddDir(m Model, args string) (tea.Model, tea.Cmd) {
	path := strings.TrimSpace(args)
	if path == "" {
		return m.sys("usage: /add-dir <path>"), nil
	}
	if m.actions.AddDir == nil {
		return m.unavailable("add-dir"), nil
	}
	msg := m.actions.AddDir(context.Background(), path)
	m.filesLoaded = false // re-read the file list so the new directory's files appear in @-mentions
	return m.sys(msg), nil
}

// cmdCd changes the session working directory, updating the header (workspace, branch) and refreshing
// the file and skill caches so subsequent runs and completions use the new root.
func cmdCd(m Model, args string) (tea.Model, tea.Cmd) {
	path := strings.TrimSpace(args)
	if path == "" {
		return m.sys("usage: /cd <path>"), nil
	}
	if m.actions.Cd == nil {
		return m.unavailable("cd"), nil
	}
	dir, branch, status := m.actions.Cd(context.Background(), path)
	if dir != "" {
		m.workspaceRoot = dir
		m.branch = branch
		m.filesLoaded = false
		m.mentionsLoaded = false
	}
	return m.sys(status), nil
}
