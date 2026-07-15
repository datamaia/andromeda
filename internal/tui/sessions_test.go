package tui

import (
	"context"
	"strings"
	"testing"
)

func TestCmdBtwQueuesNote(t *testing.T) {
	var got string
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Note: func(_ context.Context, text string) string { got = text; return "noted (1 pending)" },
	})
	nm, cmd := cmdBtw(m, "remember the cache TTL")
	if cmd != nil {
		t.Fatal("/btw must not start a run")
	}
	if got != "remember the cache TTL" {
		t.Fatalf("note text = %q", got)
	}
	body := lastText(nm.(Model))
	if !strings.Contains(body, "noted") {
		t.Fatalf("transcript missing confirmation: %q", body)
	}
	// The queued note is echoed so the user sees what was captured.
	var echoed bool
	for _, e := range nm.(Model).transcript {
		if strings.Contains(e.text, "remember the cache TTL") {
			echoed = true
		}
	}
	if !echoed {
		t.Fatal("/btw should echo the note into the transcript")
	}
}

func TestCmdBtwUsage(t *testing.T) {
	m := New("ollama", "llama3", nil)
	nm, _ := cmdBtw(m, "   ")
	if !strings.Contains(lastText(nm.(Model)), "usage") {
		t.Fatalf("blank /btw should show usage: %q", lastText(nm.(Model)))
	}
}

func TestCmdSessionsResumeReseeds(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		ResumeSession: func(_ context.Context, id string) ([]HistoryEntry, bool, string) {
			return []HistoryEntry{{Role: "user", Text: "earlier goal"}, {Role: "agent", Text: "earlier reply"}},
				true, "resumed session " + id + " · 1 turns restored"
		},
	})
	m.transcript = append(m.transcript, entry{"user", "stale line"})
	nm, _ := cmdSessions(m, "resume 20260101-000000-abc")
	got := nm.(Model)
	// The stale transcript is replaced by the resumed header + restored turns.
	if got.transcript[0].role != "system" || !strings.Contains(got.transcript[0].text, "resumed session") {
		t.Fatalf("first line should be the resume header: %+v", got.transcript[0])
	}
	joined := ""
	for _, e := range got.transcript {
		joined += e.text + "\n"
	}
	if strings.Contains(joined, "stale line") {
		t.Fatal("resume should clear the prior transcript")
	}
	if !strings.Contains(joined, "earlier goal") || !strings.Contains(joined, "earlier reply") {
		t.Fatalf("resume did not re-seed restored turns: %q", joined)
	}
}

func TestCmdSessionsResumeGuards(t *testing.T) {
	// Refused mid-run.
	running := New("ollama", "llama3", nil).WithActions(Actions{
		ResumeSession: func(_ context.Context, _ string) ([]HistoryEntry, bool, string) {
			t.Fatal("resume must not run while a turn is active")
			return nil, false, ""
		},
	})
	running.running = true
	nm, _ := cmdSessions(running, "resume x")
	if !strings.Contains(lastText(nm.(Model)), "interrupt the current run") {
		t.Fatalf("running guard: %q", lastText(nm.(Model)))
	}
	// Missing id -> usage.
	m := New("ollama", "llama3", nil)
	nm, _ = cmdSessions(m, "resume")
	if !strings.Contains(lastText(nm.(Model)), "usage") {
		t.Fatalf("resume usage: %q", lastText(nm.(Model)))
	}
	// Unknown id (ok=false) leaves the transcript in place and shows the error.
	m2 := New("ollama", "llama3", nil).WithActions(Actions{
		ResumeSession: func(_ context.Context, _ string) ([]HistoryEntry, bool, string) {
			return nil, false, "resume z: not found"
		},
	})
	nm, _ = cmdSessions(m2, "resume z")
	if !strings.Contains(lastText(nm.(Model)), "not found") {
		t.Fatalf("resume failure: %q", lastText(nm.(Model)))
	}
}

func TestCmdSessionsBranchCloneTree(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Sessions:    func(_ context.Context, _ string) string { return "1 saved session(s):" },
		Branch:      func(_ context.Context) string { return "branched this conversation" },
		Clone:       func(_ context.Context) string { return "cloned this conversation" },
		SessionTree: func(_ context.Context) string { return "session tree" },
	})
	if nm, _ := cmdSessions(m, "list"); !strings.Contains(lastText(nm.(Model)), "saved session") {
		t.Fatalf("list: %q", lastText(nm.(Model)))
	}
	if nm, _ := cmdBranch(m, ""); !strings.Contains(lastText(nm.(Model)), "branched") {
		t.Fatalf("branch: %q", lastText(nm.(Model)))
	}
	if nm, _ := cmdClone(m, ""); !strings.Contains(lastText(nm.(Model)), "cloned") {
		t.Fatalf("clone: %q", lastText(nm.(Model)))
	}
	if nm, _ := cmdTree(m, ""); !strings.Contains(lastText(nm.(Model)), "session tree") {
		t.Fatalf("tree: %q", lastText(nm.(Model)))
	}
}

func TestCmdSessionsUnavailable(t *testing.T) {
	m := New("ollama", "llama3", nil) // no actions wired
	if nm, _ := cmdBranch(m, ""); !strings.Contains(lastText(nm.(Model)), "unavailable") {
		t.Fatalf("branch unavailable: %q", lastText(nm.(Model)))
	}
	if nm, _ := cmdSessions(m, "list"); !strings.Contains(lastText(nm.(Model)), "unavailable") {
		t.Fatalf("sessions unavailable: %q", lastText(nm.(Model)))
	}
}
