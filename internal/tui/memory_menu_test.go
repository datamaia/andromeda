package tui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Bare /memory opens the CRUD menu with a friendly empty state and add/search actions.
func TestMemoryMenuEmptyState(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		MemoryList: func(context.Context) []MemoryNote { return nil },
	})
	nm, _ := cmdMemory(m, "")
	got := nm.(Model)
	if !got.menuOpen() {
		t.Fatal("bare /memory should open the menu")
	}
	top := got.menu[len(got.menu)-1]
	if !strings.Contains(top.hint, "No memories yet") {
		t.Errorf("empty hint = %q", top.hint)
	}
	labels := itemLabels(top)
	if !strings.Contains(strings.Join(labels, "|"), "Add a note") || !strings.Contains(strings.Join(labels, "|"), "Search") {
		t.Errorf("menu should offer Add and Search, got %v", labels)
	}
}

// A text subcommand (/memory add …) goes straight to the Memory action, not the menu.
func TestMemoryTextSubcommand(t *testing.T) {
	got := ""
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Memory: func(_ context.Context, args string) string { got = args; return "remembered 0001" },
	})
	nm, _ := cmdMemory(m, "add Fix the bug #bug")
	if got != "add Fix the bug #bug" {
		t.Errorf("memory action args = %q", got)
	}
	tr := nm.(Model).Transcript()
	if !strings.Contains(tr[len(tr)-1], "remembered 0001") {
		t.Errorf("transcript = %v", tr)
	}
}

// Notes are listed and drill into a detail level with delete/edit; deleting refreshes the list.
func TestMemoryMenuListDrillAndDelete(t *testing.T) {
	notes := []MemoryNote{{ID: "0001", Title: "Auth chunking", Tags: []string{"auth"}, Created: "2026-07-15", Path: ".andromeda/memory/0001-auth-chunking.md", Preview: "chunk secrets"}}
	deleted := ""
	m := New("ollama", "llama3", nil).WithActions(Actions{
		MemoryList: func(context.Context) []MemoryNote {
			if deleted == "0001" {
				return nil
			}
			return notes
		},
		Memory: func(_ context.Context, args string) string {
			deleted = strings.TrimPrefix(args, "rm ")
			return "deleted memory 0001"
		},
	})
	nm, _ := cmdMemory(m, "")
	tm := nm
	tm, _ = tm.Update(key(tea.KeyEnter)) // drill into the note
	got := tm.(Model)
	if len(got.menu) != 2 || got.menu[1].title != "Auth chunking" {
		t.Fatalf("should drill into the note, stack=%d", len(got.menu))
	}
	view := stripANSI(got.View().Content)
	if !strings.Contains(view, "chunk secrets") || !strings.Contains(view, "Delete") {
		t.Errorf("detail view missing preview/delete:\n%s", view)
	}
	// Navigate to Delete (Preview, File, Edit via chat, Delete) and activate it.
	for i := 0; i < 3; i++ {
		tm, _ = tm.Update(key(tea.KeyDown))
	}
	tm, _ = tm.Update(key(tea.KeyEnter))
	final := tm.(Model)
	if deleted != "0001" {
		t.Errorf("delete not invoked, deleted=%q", deleted)
	}
	// After delete, the refreshed root list is empty (shows the empty state).
	if !final.menuOpen() || !strings.Contains(final.menu[len(final.menu)-1].hint, "No memories yet") {
		t.Errorf("after delete the list should refresh to empty, menu=%+v", final.menu)
	}
}

func itemLabels(l menuLevel) []string {
	out := make([]string, len(l.items))
	for i, it := range l.items {
		out[i] = it.label
	}
	return out
}
