package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// cmdBranch (/branch) snapshots the conversation into a new saved session, leaving the user where
// they are. cmdClone (/clone) instead continues on a fresh copy so the original is preserved.
func cmdBranch(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.Branch == nil {
		return m.sys("branching is unavailable"), nil
	}
	return m.sys(m.actions.Branch(context.Background())), nil
}

func cmdClone(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.Clone == nil {
		return m.sys("cloning is unavailable"), nil
	}
	return m.sys(m.actions.Clone(context.Background())), nil
}

// cmdTree (/tree) renders the fork lineage of saved sessions.
func cmdTree(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.SessionTree == nil {
		return m.sys("session tree is unavailable"), nil
	}
	return m.sys(m.actions.SessionTree(context.Background())), nil
}

// cmdBtw (/btw) queues an out-of-band note. It does not start a run: the note is folded into the
// user's next real message, and the composer confirms it inline.
func cmdBtw(m Model, args string) (tea.Model, tea.Cmd) {
	args = strings.TrimSpace(args)
	if args == "" {
		return m.sys("usage: /btw <note> — the agent sees it on your next message"), nil
	}
	if m.actions.Note == nil {
		return m.sys("notes are unavailable"), nil
	}
	status := m.actions.Note(context.Background(), args)
	m.transcript = append(m.transcript, entry{"system", "btw: " + args}, entry{"system", status})
	return m, nil
}

// cmdSessions (/sessions) lists or removes saved sessions, and resumes one in place. Resuming swaps
// the live conversation and re-seeds the visible transcript; it is refused mid-run.
func cmdSessions(m Model, args string) (tea.Model, tea.Cmd) {
	args = strings.TrimSpace(args)
	if sub, rest, _ := strings.Cut(args, " "); sub == "resume" {
		id := strings.TrimSpace(rest)
		if id == "" {
			return m.sys("usage: /sessions resume <id>"), nil
		}
		if m.running {
			return m.sys("finish or interrupt the current run before resuming another session"), nil
		}
		if m.actions.ResumeSession == nil {
			return m.sys("resuming is unavailable"), nil
		}
		entries, ok, status := m.actions.ResumeSession(context.Background(), id)
		if !ok {
			return m.sys(status), nil // e.g. unknown id — leave the transcript untouched
		}
		m.transcript = seedTranscript(status, entries)
		m.scrollOffset = 0
		return m, nil
	}
	if m.actions.Sessions == nil {
		return m.sys("sessions are unavailable"), nil
	}
	return m.sys(m.actions.Sessions(context.Background(), args)), nil
}

// cmdAdvisor (/advisor) consults a second opinion. Config/usage replies are synchronous; a real
// question hits the provider, so it runs off the UI thread and reports back as a notice.
func cmdAdvisor(m Model, args string) (tea.Model, tea.Cmd) {
	if m.actions.Advisor == nil {
		return m.unavailable("advisor"), nil
	}
	q := strings.TrimSpace(args)
	fn := m.actions.Advisor
	if q == "" || q == "model" || strings.HasPrefix(q, "model ") {
		return m.sys(fn(context.Background(), q)), nil
	}
	return m.sys("consulting the advisor…"), func() tea.Msg {
		return noticeMsg{text: fn(context.Background(), q)}
	}
}

// cmdShare (/share) uploads the transcript as a secret gist off the UI thread.
func cmdShare(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.Share == nil {
		return m.unavailable("share"), nil
	}
	lines := m.Transcript()
	fn := m.actions.Share
	return m.sys("uploading transcript to a secret gist…"), func() tea.Msg {
		return noticeMsg{text: fn(lines)}
	}
}

// cmdUnshare (/unshare) deletes the gist created by /share off the UI thread.
func cmdUnshare(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.Unshare == nil {
		return m.unavailable("unshare"), nil
	}
	fn := m.actions.Unshare
	return m.sys("removing shared gist…"), func() tea.Msg {
		return noticeMsg{text: fn(context.Background())}
	}
}
