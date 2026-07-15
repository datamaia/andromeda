package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Ontology and graph are Andromeda's context-engineering surfaces: deterministic, regenerable maps
// of the workspace written under .andromeda/. Both are exposed as slash commands that open a small
// menu (build / show / adjust-via-chat / delete) or take a direct op argument (e.g. /graph open).
// The heavy lifting lives behind Actions.Ontology / Actions.Graph so this package stays free of any
// filesystem/app imports.

// cmdOntology opens the ontology menu, or runs an op directly (/ontology build|show|adjust|rm).
func cmdOntology(m Model, args string) (tea.Model, tea.Cmd) {
	if op := firstWord(args); op != "" {
		return m.runOntologyOp(op), nil
	}
	return m.openOntologyMenu()
}

// cmdGraph opens the graph menu, or runs an op directly (/graph build|open|show|adjust|rm).
func cmdGraph(m Model, args string) (tea.Model, tea.Cmd) {
	if op := firstWord(args); op != "" {
		return m.runGraphOp(op), nil
	}
	return m.openGraphMenu()
}

func (m Model) openOntologyMenu() (tea.Model, tea.Cmd) {
	return m.openMenu(menuLevel{
		title: "Ontology · workspace map (.ttl)",
		items: []menuItem{
			{label: "Build", desc: "scan & (re)write the .ttl ontology", run: ontologyOpItem("build")},
			{label: "Show", desc: "print the current ontology", run: ontologyOpItem("show")},
			{label: "Adjust via chat", desc: "describe a change for the agent", run: ontologyOpItem("adjust")},
			{label: "Delete", desc: "remove the generated ontology", run: ontologyOpItem("rm")},
		},
	})
}

func (m Model) openGraphMenu() (tea.Model, tea.Cmd) {
	return m.openMenu(menuLevel{
		title: "Graph · visual workspace map",
		items: []menuItem{
			{label: "Build", desc: "scan & (re)write the graph", run: graphOpItem("build")},
			{label: "Open viewer", desc: "serve the interactive graph in your browser", run: graphOpItem("open")},
			{label: "Show", desc: "print the graph overview", run: graphOpItem("show")},
			{label: "Adjust via chat", desc: "describe a change for the agent", run: graphOpItem("adjust")},
			{label: "Delete", desc: "remove the generated graph", run: graphOpItem("rm")},
		},
	})
}

// ontologyOpItem / graphOpItem adapt a context-engineering op into a menu action: close the menu,
// run the op (which appends its result or seeds a prompt), and stay put.
func ontologyOpItem(op string) func(Model) (Model, tea.Cmd) {
	return func(m Model) (Model, tea.Cmd) { return m.closeMenu().runOntologyOp(op), nil }
}

func graphOpItem(op string) func(Model) (Model, tea.Cmd) {
	return func(m Model) (Model, tea.Cmd) { return m.closeMenu().runGraphOp(op), nil }
}

// runOntologyOp runs one ontology operation. "adjust" seeds the prompt with an editable goal so the
// agent can refine the artifact; the rest are backed by Actions.Ontology.
func (m Model) runOntologyOp(op string) Model {
	switch op {
	case "adjust":
		m.input = "Adjust the ontology under .andromeda/ontology/ to "
		return m.sys("describe the change and press enter — the agent will edit the ontology")
	case "build", "show", "rm":
		if m.actions.Ontology == nil {
			return m.unavailable("ontology")
		}
		return m.sys(m.actions.Ontology(context.Background(), op))
	default:
		return m.sys("usage: /ontology build | show | adjust | rm")
	}
}

// runGraphOp runs one graph operation, mirroring runOntologyOp with an extra "open" that serves the
// interactive viewer.
func (m Model) runGraphOp(op string) Model {
	switch op {
	case "adjust":
		m.input = "Adjust the graph notes under .andromeda/graph/ to "
		return m.sys("describe the change and press enter — the agent will edit the graph notes")
	case "build", "show", "rm", "open":
		if m.actions.Graph == nil {
			return m.unavailable("graph")
		}
		return m.sys(m.actions.Graph(context.Background(), op))
	default:
		return m.sys("usage: /graph build | open | show | adjust | rm")
	}
}

// firstWord returns the lowercased first whitespace-delimited token of s ("" if none).
func firstWord(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return strings.ToLower(f[0])
}
