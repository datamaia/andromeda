package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFrontMatter(t *testing.T) {
	desc, body := parseFrontMatter("---\ndescription: Greet a user\n---\nSay hello to $1")
	if desc != "Greet a user" {
		t.Errorf("desc = %q", desc)
	}
	if body != "Say hello to $1" {
		t.Errorf("body = %q", body)
	}
	// No front matter: whole string is the body.
	if d, b := parseFrontMatter("just a body"); d != "" || b != "just a body" {
		t.Errorf("no-frontmatter parse = (%q,%q)", d, b)
	}
}

func TestDiscoverCustomCommands(t *testing.T) {
	wd := t.TempDir()
	mkCmd := func(dir, name, content string) {
		full := filepath.Join(wd, dir)
		if err := os.MkdirAll(full, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(full, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mkCmd(".agents/commands", "review.md", "---\ndescription: Review the diff\n---\nReview $ARGUMENTS")
	// Same name in a lower-precedence dir must be ignored.
	mkCmd(".claude/commands", "review.md", "SHOULD NOT WIN")
	mkCmd(".claude/commands", "ship.md", "Ship it")

	cmds := discoverCustomCommands(wd)
	if len(cmds) != 2 {
		t.Fatalf("discovered %d commands, want 2: %+v", len(cmds), cmds)
	}
	// sorted by name: review, ship
	if cmds[0].Name != "review" || cmds[0].Desc != "Review the diff" || cmds[0].Template != "Review $ARGUMENTS" {
		t.Errorf("review command = %+v", cmds[0])
	}
	if cmds[1].Name != "ship" || cmds[1].Desc != "custom command" {
		t.Errorf("ship command = %+v", cmds[1])
	}
}
