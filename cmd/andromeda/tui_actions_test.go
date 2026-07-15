package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/permstore"
	"github.com/datamaia/andromeda/internal/ports"
)

func TestPermissionActionAllowDenyListRemove(t *testing.T) {
	wd := t.TempDir()
	s := &tuiSession{wd: wd, ctx: context.Background()}

	if got := s.permissionAction(context.Background(), "allow git status"); !strings.Contains(got, "added to allow") {
		t.Fatalf("allow: %q", got)
	}
	if got := s.permissionAction(context.Background(), "deny rm -rf"); !strings.Contains(got, "added to deny") {
		t.Fatalf("deny: %q", got)
	}
	// Both lists must be persisted under the .andromeda store.
	r, err := permstore.Load(wd)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Allow) != 1 || len(r.Deny) != 1 {
		t.Fatalf("store = %+v", r)
	}
	if list := s.permissionAction(context.Background(), "list"); !strings.Contains(list, "git status") || !strings.Contains(list, "rm -rf") {
		t.Fatalf("list missing rules: %q", list)
	}
	if got := s.permissionAction(context.Background(), "rm allow git status"); !strings.Contains(got, "removed from allow") {
		t.Fatalf("rm: %q", got)
	}
	if r, _ = permstore.Load(wd); len(r.Allow) != 0 {
		t.Fatalf("allow not removed: %+v", r)
	}
}

func TestPermissionActionUsageAndUnknown(t *testing.T) {
	s := &tuiSession{wd: t.TempDir(), ctx: context.Background()}
	if got := s.permissionAction(context.Background(), "allow"); !strings.Contains(got, "usage") {
		t.Fatalf("allow no arg: %q", got)
	}
	if got := s.permissionAction(context.Background(), "rm bogus x"); !strings.Contains(got, "usage") {
		t.Fatalf("rm bad list: %q", got)
	}
	if got := s.permissionAction(context.Background(), "frobnicate"); !strings.Contains(got, "subcommands") {
		t.Fatalf("unknown: %q", got)
	}
	// A bare invocation lists the (empty) policy.
	if got := s.permissionAction(context.Background(), ""); !strings.Contains(got, "permission ·") {
		t.Fatalf("bare: %q", got)
	}
}

func TestPermissionView(t *testing.T) {
	wd := t.TempDir()
	s := &tuiSession{wd: wd, ctx: context.Background()}
	if _, err := permstore.Add(wd, permstore.Allow, "go test ./..."); err != nil {
		t.Fatal(err)
	}
	v := s.permissionView(context.Background())
	if len(v.Allow) != 1 || !v.Allow[0].Managed || v.Allow[0].Command != "go test ./..." {
		t.Fatalf("view.Allow = %+v", v.Allow)
	}
	if v.Path == "" {
		t.Fatal("view.Path should point at the store")
	}
}

func TestSkillCollectionAndListAction(t *testing.T) {
	wd := t.TempDir()
	dir := filepath.Join(wd, ".agents", "skills", "greeter")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	md := "---\nname: greeter\ndescription: says hi\n---\nSay hello warmly."
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(md), 0o600); err != nil {
		t.Fatal(err)
	}
	s := &tuiSession{wd: wd, ctx: context.Background()}

	cv := s.skillCollection()
	if len(cv.Entries) != 1 || cv.Entries[0].Title != "greeter" {
		t.Fatalf("skillCollection = %+v", cv.Entries)
	}
	notes := s.skillListAction(context.Background())
	if len(notes) != 1 || notes[0].Name != "greeter" || !strings.Contains(notes[0].Body, "hello") {
		t.Fatalf("skillListAction = %+v", notes)
	}
}

func TestWorkflowCollection(t *testing.T) {
	wd := t.TempDir()
	dir := filepath.Join(wd, ".agents", "workflows")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	md := "---\ndescription: ship it\n---\n1. build\n2. test\n3. release"
	if err := os.WriteFile(filepath.Join(dir, "release.md"), []byte(md), 0o600); err != nil {
		t.Fatal(err)
	}
	s := &tuiSession{wd: wd, ctx: context.Background()}
	cv := s.workflowCollection()
	if len(cv.Entries) != 1 || cv.Entries[0].Title != "release" {
		t.Fatalf("workflowCollection = %+v", cv.Entries)
	}
	if !strings.Contains(cv.Entries[0].Body, "build") || cv.Entries[0].Detail != "ship it" {
		t.Fatalf("entry = %+v", cv.Entries[0])
	}
}

func TestFallbackModelsKiro(t *testing.T) {
	// No provider built (s.prov nil) → modelsAction returns the catalog's curated list, so Kiro's
	// models are pickable before its gateway is running.
	s := &tuiSession{cfg: tuiConfig{provider: "kiro"}}
	got := s.modelsAction(context.Background())
	if len(got) == 0 {
		t.Fatal("expected Kiro fallback models")
	}
	var hasSonnet bool
	for _, m := range got {
		if m == "claude-sonnet-4-5" {
			hasSonnet = true
		}
	}
	if !hasSonnet {
		t.Fatalf("Kiro fallback missing claude-sonnet-4-5: %v", got)
	}
}

func TestAddDirAndCdActions(t *testing.T) {
	wd := t.TempDir()
	extra := t.TempDir()
	if err := os.WriteFile(filepath.Join(extra, "note.txt"), []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := &tuiSession{wd: wd}

	if msg := s.addDirAction(context.Background(), extra); !strings.Contains(msg, "added working directory") {
		t.Fatalf("add-dir: %q", msg)
	}
	if len(s.extraDirs) != 1 {
		t.Fatalf("extraDirs = %v", s.extraDirs)
	}
	if msg := s.addDirAction(context.Background(), extra); !strings.Contains(msg, "already added") {
		t.Fatalf("add-dir dedup: %q", msg)
	}
	if msg := s.addDirAction(context.Background(), filepath.Join(wd, "nope")); !strings.Contains(msg, "not a directory") {
		t.Fatalf("add-dir invalid: %q", msg)
	}
	// The extra directory's files join the @-mention list.
	var found bool
	for _, f := range s.listFiles(context.Background()) {
		if strings.HasSuffix(f, "note.txt") {
			found = true
		}
	}
	if !found {
		t.Fatal("listFiles should include the extra directory's files")
	}

	dir, _, status := s.cdAction(context.Background(), extra)
	if dir != extra || !strings.Contains(status, "working directory") {
		t.Fatalf("cd: dir=%q status=%q", dir, status)
	}
	if s.wd != extra {
		t.Fatalf("cd did not move wd: %q", s.wd)
	}
	if d, _, status := s.cdAction(context.Background(), filepath.Join(extra, "nope")); d != "" || !strings.Contains(status, "not a directory") {
		t.Fatalf("cd invalid: d=%q status=%q", d, status)
	}
}

func TestResetSessionAction(t *testing.T) {
	s := &tuiSession{history: []ports.Message{{Role: "user"}, {Role: "assistant"}}, sessionID: "old-id"}
	s.resetSessionAction(context.Background())
	if s.history != nil {
		t.Fatalf("history should be cleared, got %v", s.history)
	}
	if s.sessionID == "" || s.sessionID == "old-id" {
		t.Fatalf("resetSession should mint a fresh id, got %q", s.sessionID)
	}
}

func TestConfigStrings(t *testing.T) {
	cases := []struct {
		in   any
		want int
	}{
		{[]string{"a", " ", "b"}, 2},
		{[]any{"x", 3, "y"}, 2},
		{"single", 1},
		{"", 0},
		{42, 0},
		{nil, 0},
	}
	for _, c := range cases {
		if got := configStrings(c.in); len(got) != c.want {
			t.Errorf("configStrings(%v) = %v, want len %d", c.in, got, c.want)
		}
	}
}
