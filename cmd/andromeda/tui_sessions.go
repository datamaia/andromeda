package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/tui"
)

// nowRFC is the current instant as an RFC3339 UTC string, matching persistSession's format.
func nowRFC() string { return time.Now().UTC().Format(time.RFC3339) }

// branchAction (/branch) snapshots the current conversation as a new saved session and stays on the
// current one — a bookmark you can resume later without diverging from where you are now.
func (s *tuiSession) branchAction(_ context.Context) string {
	if len(s.history) == 0 {
		return "nothing to branch yet — send a message first"
	}
	forkID := app.NewSessionID()
	if err := app.SaveSession(app.StoredSession{
		ID:        forkID,
		Parent:    s.sessionID,
		Title:     app.SessionTitle(s.history),
		Provider:  s.cfg.provider,
		Model:     s.cfg.model,
		Mode:      "agent",
		UpdatedAt: nowRFC(),
		Messages:  append([]ports.Message(nil), s.history...),
	}); err != nil {
		return "branch failed: " + err.Error()
	}
	return fmt.Sprintf("branched this conversation → session %s\n"+
		"you stay on the current line; open the branch later with  andromeda --resume %s", forkID, forkID)
}

// cloneAction (/clone) freezes the current line under its own id and continues on a fresh copy, so
// the original is preserved at this point while new turns extend the clone.
func (s *tuiSession) cloneAction(_ context.Context) string {
	if len(s.history) == 0 {
		return "nothing to clone yet — send a message first"
	}
	parent := s.sessionID
	s.persistSession("agent") // freeze the line we're leaving under its current id
	s.sessionID = app.NewSessionID()
	if err := app.SaveSession(app.StoredSession{
		ID:        s.sessionID,
		Parent:    parent,
		Title:     app.SessionTitle(s.history),
		Provider:  s.cfg.provider,
		Model:     s.cfg.model,
		Mode:      "agent",
		UpdatedAt: nowRFC(),
		Messages:  append([]ports.Message(nil), s.history...),
	}); err != nil {
		return "clone failed: " + err.Error()
	}
	return fmt.Sprintf("cloned this conversation → now working on session %s\nthe original stays saved as %s", s.sessionID, parent)
}

// noteAction (/btw) queues an out-of-band note that is folded into the next message to the agent
// without triggering a reply now — the "by the way, keep this in mind" convention.
func (s *tuiSession) noteAction(_ context.Context, text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "note was empty"
	}
	s.pendingNotes = append(s.pendingNotes, text)
	return fmt.Sprintf("noted (%d pending) — folded into your next message to the agent", len(s.pendingNotes))
}

// sessionsAction (/sessions) lists or removes saved sessions. The resume subcommand is handled in
// the TUI (it re-seeds the transcript) via resumeSessionAction.
func (s *tuiSession) sessionsAction(_ context.Context, args string) string {
	sub, rest, _ := strings.Cut(strings.TrimSpace(args), " ")
	switch sub {
	case "rm", "remove", "delete":
		id := strings.TrimSpace(rest)
		if id == "" {
			return "usage: /sessions rm <id>"
		}
		if id == s.sessionID {
			return "refusing to remove the session you're in"
		}
		if err := app.RemoveSession(id); err != nil {
			return "remove failed: " + err.Error()
		}
		return "removed session " + id
	case "", "list", "ls":
		return s.sessionsList()
	default:
		return "usage: /sessions [list | resume <id> | rm <id>]"
	}
}

// sessionsList renders every saved session, newest first, marking the current one.
func (s *tuiSession) sessionsList() string {
	sessions, err := app.ListSessions()
	if err != nil {
		return "could not list sessions: " + err.Error()
	}
	if len(sessions) == 0 {
		return "no saved sessions yet"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d saved session(s):\n", len(sessions))
	for _, st := range sessions {
		marker := "  "
		if st.ID == s.sessionID {
			marker = "▸ "
		}
		fmt.Fprintf(&b, "%s%s · %d turns · %s · %s\n",
			marker, st.ID, app.CountTurns(st.Messages), shortStamp(st.UpdatedAt), st.Title)
	}
	b.WriteString("\nresume with  /sessions resume <id>   ·   remove with  /sessions rm <id>")
	return b.String()
}

// sessionTreeAction (/tree) renders the fork structure: each session under the parent it was
// branched from, so /branch and /clone lineage is visible at a glance.
func (s *tuiSession) sessionTreeAction(_ context.Context) string {
	sessions, err := app.ListSessions()
	if err != nil {
		return "could not read sessions: " + err.Error()
	}
	if len(sessions) == 0 {
		return "no saved sessions yet"
	}
	exists := map[string]bool{}
	for _, st := range sessions {
		exists[st.ID] = true
	}
	children := map[string][]app.StoredSession{}
	var roots []app.StoredSession
	for _, st := range sessions {
		if st.Parent != "" && exists[st.Parent] {
			children[st.Parent] = append(children[st.Parent], st)
		} else {
			roots = append(roots, st) // original, or parent no longer on disk
		}
	}
	var b strings.Builder
	b.WriteString("session tree (branches nested under their origin):\n")
	seen := map[string]bool{}
	var walk func(st app.StoredSession, depth int)
	walk = func(st app.StoredSession, depth int) {
		if seen[st.ID] { // defensive: never loop on a cyclic parent chain
			return
		}
		seen[st.ID] = true
		cur := ""
		if st.ID == s.sessionID {
			cur = "  ◂ current"
		}
		fmt.Fprintf(&b, "%s• %s · %d turns%s\n",
			strings.Repeat("  ", depth), st.ID, app.CountTurns(st.Messages), cur)
		for _, c := range children[st.ID] {
			walk(c, depth+1)
		}
	}
	for _, r := range roots {
		walk(r, 0)
	}
	return b.String()
}

// resumeSessionAction (/sessions resume <id>) swaps the live conversation to a saved session,
// persisting the one being left. The transcript is re-seeded from the returned entries. The current
// provider and model are kept (the transcript is provider-neutral, so the agent continues cleanly);
// a full provider switch happens on a cold `andromeda --resume`.
func (s *tuiSession) resumeSessionAction(_ context.Context, id string) ([]tui.HistoryEntry, bool, string) {
	st, err := app.LoadSession(id)
	if err != nil {
		return nil, false, "resume " + id + ": " + err.Error()
	}
	s.persistSession("agent") // don't lose the session we're leaving
	s.sessionID = st.ID
	s.history = st.Messages
	s.pendingNotes = nil
	return historyEntries(st.Messages), true, fmt.Sprintf(
		"resumed session %s · %d turns restored (continuing under %s · %s)",
		st.ID, app.CountTurns(st.Messages), s.cfg.provider, s.cfg.model)
}

// shortStamp trims an RFC3339 timestamp to "YYYY-MM-DD HH:MM" for compact listings.
func shortStamp(rfc string) string {
	if t, err := time.Parse(time.RFC3339, rfc); err == nil {
		return t.Local().Format("2006-01-02 15:04")
	}
	if len(rfc) >= 16 {
		return strings.Replace(rfc[:16], "T", " ", 1)
	}
	return rfc
}
