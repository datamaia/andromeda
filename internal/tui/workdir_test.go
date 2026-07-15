package tui

import (
	"context"
	"testing"
)

func TestCmdCdUpdatesHeader(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Cd: func(_ context.Context, _ string) (string, string, string) {
			return "/new/dir", "feature", "working directory → /new/dir"
		},
	})
	m.filesLoaded = true
	m.skillsLoaded = true
	nm, _ := cmdCd(m, "somewhere")
	got := nm.(Model)
	if got.workspaceRoot != "/new/dir" || got.branch != "feature" {
		t.Fatalf("/cd should update the header: root=%q branch=%q", got.workspaceRoot, got.branch)
	}
	if got.filesLoaded || got.skillsLoaded {
		t.Fatal("/cd should invalidate the file and skill caches")
	}
}

func TestCmdCdError(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Cd: func(_ context.Context, _ string) (string, string, string) { return "", "", "cd: not a directory: x" },
	})
	m.workspaceRoot = "/old"
	nm, _ := cmdCd(m, "x")
	if nm.(Model).workspaceRoot != "/old" {
		t.Fatal("/cd should leave the workspace unchanged on error")
	}
}

func TestClearResetsSession(t *testing.T) {
	called := false
	m := New("ollama", "llama3", nil).WithActions(Actions{
		ResetSession: func(_ context.Context) { called = true },
	})
	m.transcript = append(m.transcript, entry{"user", "hi"}, entry{"agent", "yo"})
	nm, _ := cmdClear(m, "")
	if !called {
		t.Fatal("/clear should reset the driver's cross-turn history, not just the display")
	}
	got := nm.(Model)
	if len(got.transcript) != 1 || got.transcript[0].role != "system" {
		t.Fatalf("/clear should leave a single system line, got %+v", got.transcript)
	}
}

func TestCmdAddDirUsage(t *testing.T) {
	m := New("ollama", "llama3", nil)
	nm, _ := cmdAddDir(m, "")
	if got := lastText(nm.(Model)); got != "usage: /add-dir <path>" {
		t.Fatalf("empty add-dir should show usage, got %q", got)
	}
}

// A run alternates thinking → working → responding; the header label and spinner must follow.
func TestThinkingPhase(t *testing.T) {
	m := Model{running: true, state: "running"}
	if label, thinking := m.runPhase(); label != "thinking" || !thinking {
		t.Fatalf("run start should be thinking, got %q/%v", label, thinking)
	}
	m.state = "working"
	if label, thinking := m.runPhase(); label != "working" || thinking {
		t.Fatalf("tool phase should be working, got %q/%v", label, thinking)
	}
	m.state = "streaming"
	if label, thinking := m.runPhase(); label != "responding" || thinking {
		t.Fatalf("stream phase should be responding, got %q/%v", label, thinking)
	}
	// The thinking animation must be visually distinct from the working one.
	thinkM := Model{running: true, state: "running"}
	workM := Model{running: true, state: "working"}
	if thinkM.spinnerFrame() == workM.spinnerFrame() {
		t.Fatal("thinking and working spinners should use different glyph sets")
	}
}
