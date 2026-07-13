package tui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestAtToken(t *testing.T) {
	cases := map[string]struct {
		frag string
		ok   bool
	}{
		"open @READ":     {"READ", true},
		"@src/main":      {"src/main", true},
		"just words":     {"", false},
		"email a@b done": {"", false}, // token ends after the space
		"@":              {"", true},
	}
	for in, want := range cases {
		frag, ok := atToken(in)
		if ok != want.ok || frag != want.frag {
			t.Errorf("atToken(%q) = (%q,%v), want (%q,%v)", in, frag, ok, want.frag, want.ok)
		}
	}
}

// Typing "@frag" opens a ranked file menu; Enter inserts "@path ".
func TestAtMentionInsertsPath(t *testing.T) {
	files := []string{"README.md", "internal/tui/model.go", "cmd/andromeda/main.go"}
	m := New("p", "m", func(string, string) string { return "ok" }).
		WithActions(Actions{Files: func(context.Context) []string { return files }})
	var tm tea.Model = m
	tm = typeString(tm, "look at @model")
	got := tm.(Model)
	if !got.atActive() {
		t.Fatal("@model should activate the mention menu")
	}
	if match := got.matchFiles("model"); len(match) == 0 || match[0] != "internal/tui/model.go" {
		t.Fatalf("matchFiles(model) = %v, want model.go first", match)
	}
	tm, _ = tm.Update(key(tea.KeyEnter)) // insert highlighted path
	if in := tm.(Model).input; in != "look at @internal/tui/model.go " {
		t.Fatalf("after accept, input = %q", in)
	}
}

// Ranking: basename prefix beats a mid-path substring.
func TestMatchFilesRanking(t *testing.T) {
	m := New("p", "m", nil)
	m.fileList = []string{"docs/api.md", "api.go", "internal/rapid.go"}
	m.filesLoaded = true
	got := m.matchFiles("api")
	if len(got) < 2 || got[0] != "api.go" {
		t.Fatalf("matchFiles(api) = %v, want api.go ranked first", got)
	}
}

// With no file source, "@" does not trap navigation keys.
func TestAtInactiveWithoutFiles(t *testing.T) {
	var m tea.Model = New("p", "m", func(string, string) string { return "ok" })
	m = typeString(m, "@x")
	if m.(Model).atActive() {
		t.Error("@ should be inert when no files are available")
	}
}
