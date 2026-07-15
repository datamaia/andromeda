package tui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// With no entries, a collection menu shows a friendly empty state and still offers "Create".
func TestCollectionEmptyState(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Collection: func(_ context.Context, kind string) CollectionView {
			return CollectionView{Empty: "No skills yet.", Create: "they live under .agents/skills/"}
		},
	})
	nm, _ := cmdSkills(m, "")
	got := nm.(Model)
	if !got.menuOpen() {
		t.Fatal("/skills should open a menu")
	}
	top := got.menu[len(got.menu)-1]
	if !strings.Contains(top.hint, "No skills yet") {
		t.Errorf("empty-state hint = %q", top.hint)
	}
	if len(top.items) != 1 || !strings.Contains(top.items[0].label, "Create a new skill") {
		t.Errorf("empty collection should offer exactly a Create item, got %+v", top.items)
	}
	// The rendered view surfaces the empty state to the user.
	view := stripANSI(got.View().Content)
	if !strings.Contains(view, "No skills yet") {
		t.Errorf("empty state not visible on screen:\n%s", view)
	}
}

// Existing entries are listed and can be drilled into (breadcrumb + detail), and esc goes back.
func TestCollectionListAndDrillIn(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Collection: func(_ context.Context, kind string) CollectionView {
			return CollectionView{Entries: []CollectionEntry{
				{Title: "web-search@1.0.0", Detail: "search the web", Path: ".agents/skills/web-search"},
			}}
		},
	})
	nm, _ := cmdSkills(m, "")
	var tm tea.Model = nm
	// First item is the entry; drill in.
	tm, _ = tm.Update(key(tea.KeyEnter))
	got := tm.(Model)
	if len(got.menu) != 2 {
		t.Fatalf("drilling into an entry should push a detail level, stack depth = %d", len(got.menu))
	}
	detail := got.menu[len(got.menu)-1]
	if detail.title != "web-search@1.0.0" {
		t.Errorf("detail title = %q", detail.title)
	}
	view := stripANSI(got.View().Content)
	if !strings.Contains(view, ".agents/skills/web-search") || !strings.Contains(view, "Edit via chat") {
		t.Errorf("detail should show the path and an edit action:\n%s", view)
	}
	// Esc pops back to the list.
	tm, _ = tm.Update(key(tea.KeyEscape))
	if len(tm.(Model).menu) != 1 {
		t.Errorf("esc should pop one level, depth = %d", len(tm.(Model).menu))
	}
	// Esc again closes the menu.
	tm, _ = tm.Update(key(tea.KeyEscape))
	if tm.(Model).menuOpen() {
		t.Error("esc at the root should close the menu")
	}
}

// "Create via chat" seeds the prompt for the agent and closes the menu.
func TestCollectionCreateSeedsPrompt(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Collection: func(_ context.Context, kind string) CollectionView { return CollectionView{} },
	})
	nm, _ := cmdSkills(m, "")
	// The only item is Create; select it.
	tm, _ := nm.Update(key(tea.KeyEnter))
	got := tm.(Model)
	if got.menuOpen() {
		t.Error("selecting Create should close the menu")
	}
	if !strings.HasPrefix(got.input, "Create a new skill") {
		t.Errorf("Create should seed the prompt, got %q", got.input)
	}
}
