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

func TestMatchSkillsRanking(t *testing.T) {
	m := Model{skillList: []SkillNote{
		{Name: "review", Description: "review a diff"},
		{Name: "refactor", Description: "clean up code"},
		{Name: "deploy", Description: "ship to prod, review gate"},
	}}
	got := m.matchSkills("re")
	if len(got) != 3 { // "review"/"refactor" by name prefix, "deploy" by description substring
		t.Fatalf("matchSkills(re) = %d entries: %+v", len(got), got)
	}
	if got[0].Name != "review" && got[0].Name != "refactor" {
		t.Fatalf("prefix matches should rank first, got %q", got[0].Name)
	}
	if got[2].Name != "deploy" {
		t.Fatalf("substring match should rank last, got %q", got[2].Name)
	}
}

func TestExpandSkillMentions(t *testing.T) {
	m := Model{skillList: []SkillNote{
		{Name: "review", Body: "Check the diff carefully."},
		{Name: "lint", Body: "Run the linters."},
	}}
	goal := "please $review and $lint this, then $unknown"
	out, used := m.expandSkillMentions(goal)
	if len(used) != 2 || used[0] != "review" || used[1] != "lint" {
		t.Fatalf("used = %v, want [review lint]", used)
	}
	if !strings.Contains(out, "Check the diff carefully.") || !strings.Contains(out, "Run the linters.") {
		t.Fatalf("expanded goal missing skill bodies:\n%s", out)
	}
	if !strings.Contains(out, "please $review and $lint this") {
		t.Fatalf("expanded goal should retain the original text:\n%s", out)
	}
}

func TestExpandSkillMentionsNoMatch(t *testing.T) {
	m := Model{skillList: []SkillNote{{Name: "review", Body: "x"}}}
	out, used := m.expandSkillMentions("just a normal goal, no mentions")
	if used != nil {
		t.Fatalf("used = %v, want nil", used)
	}
	if out != "just a normal goal, no mentions" {
		t.Fatalf("goal should be unchanged, got %q", out)
	}
}
