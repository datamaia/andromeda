package tui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// /ontology with no argument opens the menu picker; the ontology ops are listed.
func TestOntologyOpensMenu(t *testing.T) {
	m := New("ollama", "llama3", nil)
	nm, _ := cmdOntology(m, "")
	got := nm.(Model)
	if !got.pickerOpen || got.pickerKind != "ontology" {
		t.Fatalf("/ontology should open the ontology picker, got open=%v kind=%q", got.pickerOpen, got.pickerKind)
	}
	ids := make([]string, len(got.pickerItems))
	for i, it := range got.pickerItems {
		ids[i] = it.id
	}
	if strings.Join(ids, ",") != "build,show,adjust,rm" {
		t.Errorf("ontology menu items = %v", ids)
	}
}

// A direct op (/ontology build) dispatches to the wired action without opening the menu.
func TestOntologyDirectOpRunsAction(t *testing.T) {
	called := ""
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Ontology: func(_ context.Context, op string) string { called = op; return "ontology · ok" },
	})
	nm, _ := cmdOntology(m, "build")
	got := nm.(Model)
	if got.pickerOpen {
		t.Error("a direct op should not open the menu")
	}
	if called != "build" {
		t.Errorf("action op = %q, want build", called)
	}
	tr := got.Transcript()
	if len(tr) == 0 || !strings.Contains(tr[len(tr)-1], "ontology · ok") {
		t.Errorf("transcript missing action output: %v", tr)
	}
}

// The "adjust" op seeds the prompt for the agent instead of calling the action.
func TestGraphAdjustSeedsPrompt(t *testing.T) {
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Graph: func(_ context.Context, _ string) string { t.Fatal("adjust must not call the graph action"); return "" },
	})
	got := m.runGraphOp("adjust")
	if !strings.HasPrefix(got.input, "Adjust the graph notes under .andromeda/graph/") {
		t.Errorf("adjust should seed the input, got %q", got.input)
	}
}

// An unwired action degrades to a helpful message rather than panicking.
func TestGraphUnavailableWhenUnwired(t *testing.T) {
	got := New("ollama", "llama3", nil).runGraphOp("open")
	tr := got.Transcript()
	if len(tr) == 0 || !strings.Contains(tr[len(tr)-1], "not available") {
		t.Errorf("expected an unavailable message, got %v", tr)
	}
}

// Selecting an item from the graph menu runs its op through the wired action.
func TestGraphMenuSelectionRunsOp(t *testing.T) {
	called := ""
	m := New("ollama", "llama3", nil).WithActions(Actions{
		Graph: func(_ context.Context, op string) string { called = op; return "graph · ok" },
	})
	nm, _ := cmdGraph(m, "")
	got := nm.(Model)
	// Move to "open" (second item) and select it.
	var tm tea.Model = got
	tm, _ = tm.Update(key(tea.KeyDown))
	tm, _ = tm.Update(key(tea.KeyEnter))
	if called != "open" {
		t.Errorf("selected op = %q, want open", called)
	}
	if tm.(Model).pickerOpen {
		t.Error("picker should close after selecting an op")
	}
}
