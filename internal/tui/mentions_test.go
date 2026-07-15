package tui

import (
	"strings"
	"testing"
)

func TestDollarToken(t *testing.T) {
	cases := []struct {
		in   string
		frag string
		ok   bool
	}{
		{"$rev", "rev", true},
		{"fix the bug with $ref", "ref", true},
		{"$", "", true},
		{"no token here", "", false},
		{"price is $5 today", "", false}, // trailing token is "today", not a $-token
		{"", "", false},
	}
	for _, c := range cases {
		frag, ok := dollarToken(c.in)
		if ok != c.ok || frag != c.frag {
			t.Errorf("dollarToken(%q) = (%q,%v), want (%q,%v)", c.in, frag, ok, c.frag, c.ok)
		}
	}
}

func TestMatchMentionsRanking(t *testing.T) {
	m := Model{mentionList: []Mention{
		{Kind: "skill", Name: "review", Description: "review a diff"},
		{Kind: "workflow", Name: "refactor", Description: "clean up code"},
		{Kind: "command", Name: "deploy", Description: "ship to prod, review gate"},
	}}
	got := m.matchMentions("re")
	if len(got) != 3 { // "review"/"refactor" by name prefix, "deploy" by description substring
		t.Fatalf("matchMentions(re) = %d entries: %+v", len(got), got)
	}
	if got[0].Name != "review" && got[0].Name != "refactor" {
		t.Fatalf("prefix matches should rank first, got %q", got[0].Name)
	}
	if got[2].Name != "deploy" {
		t.Fatalf("substring match should rank last, got %q", got[2].Name)
	}
}

// $-mentions cover every kind, and each is folded under a kind-appropriate header.
func TestExpandMentionsAcrossKinds(t *testing.T) {
	m := Model{mentionList: []Mention{
		{Kind: "skill", Name: "review", Body: "Check the diff carefully."},
		{Kind: "workflow", Name: "ship", Body: "1. build\n2. test\n3. release"},
		{Kind: "command", Name: "lint", Body: "Run the linters."},
		{Kind: "ontology", Name: "ontology", Body: "Use the .ttl map."},
	}}
	goal := "please $review then $ship and $lint using $ontology, ignore $unknown"
	out, used := m.expandMentions(goal)

	if len(used) != 4 {
		t.Fatalf("used = %d, want 4: %+v", len(used), used)
	}
	if used[0].Kind != "skill" || used[1].Kind != "workflow" || used[2].Kind != "command" || used[3].Kind != "ontology" {
		t.Fatalf("used kinds wrong: %+v", used)
	}
	for _, want := range []string{
		"## Skill: review", "Check the diff carefully.",
		"## Workflow: ship — follow these steps", "1. build",
		"## Command: lint", "Run the linters.",
		"## Workspace ontology map", "Use the .ttl map.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expanded goal missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "please $review then $ship") {
		t.Fatalf("expanded goal should retain the original text:\n%s", out)
	}
}

func TestExpandMentionsNoMatch(t *testing.T) {
	m := Model{mentionList: []Mention{{Kind: "skill", Name: "review", Body: "x"}}}
	out, used := m.expandMentions("just a normal goal, no mentions")
	if used != nil {
		t.Fatalf("used = %v, want nil", used)
	}
	if out != "just a normal goal, no mentions" {
		t.Fatalf("goal should be unchanged, got %q", out)
	}
}

// The menu labels each candidate with its kind so the type is visible at a glance.
func TestRenderMentionMenuShowsKind(t *testing.T) {
	m := Model{width: 80, input: "$", mentionList: []Mention{
		{Kind: "workflow", Name: "ship", Description: "release it"},
	}}
	out := m.renderMentionMenu()
	if !strings.Contains(out, "$ship") || !strings.Contains(out, "workflow") || !strings.Contains(out, "release it") {
		t.Fatalf("mention menu should show name + kind + description:\n%s", out)
	}
}
