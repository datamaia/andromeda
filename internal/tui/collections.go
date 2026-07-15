package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

// A "collection" is a manageable set of workspace capabilities — skills, MCP servers, workflows,
// plugins — surfaced as an interactive menu: list what exists (with descriptions), show a friendly
// empty state when there is nothing, and offer to create or edit an entry. The driver supplies the
// data through Actions.Collection so this package stays free of filesystem/config imports.

// CollectionEntry is one existing item in a collection.
type CollectionEntry struct {
	Title  string // display name (e.g. "web-search@1.2.0")
	Detail string // one-line description
	Path   string // file backing the entry, if any (shown so the user can find/edit it)
	Body   string // runnable content (a workflow recipe); when set, the detail view offers "Run"
}

// CollectionView is a collection's contents plus the copy for its empty and create affordances.
type CollectionView struct {
	Empty   string            // shown when there are no entries, e.g. "No skills yet."
	Create  string            // guidance for creating one, e.g. "add one under .agents/skills/…"
	Entries []CollectionEntry // existing items
}

// collectionMeta describes how one collection kind is presented.
type collectionMeta struct {
	kind     string // key passed to Actions.Collection
	title    string // menu title
	singular string // used in "Create a new <singular>" / edit prompts
	seed     string // chat prompt prefix for "create via chat"
}

var collectionKinds = map[string]collectionMeta{
	"skills":    {"skills", "Skills", "skill", "Create a skill at .agents/skills/<name>/SKILL.md with YAML frontmatter (name, description) and the instructions in the body. The skill should "},
	"mcp":       {"mcp", "MCP servers", "MCP server", "Add an MCP server to andromeda.toml that "},
	"workflows": {"workflows", "Workflows", "workflow", "Create a workflow at .agents/workflows/<name>.md with YAML frontmatter (description) and numbered step-by-step instructions in the body. The workflow should "},
	"plugins":   {"plugins", "Plugins", "plugin", "Add a plugin to andromeda.toml that "},
}

// openCollectionMenu builds the interactive menu for a collection kind: existing entries (each
// drilling into a detail level), a friendly empty state, and a "create via chat" action.
func (m Model) openCollectionMenu(kind string) (tea.Model, tea.Cmd) {
	meta, ok := collectionKinds[kind]
	if !ok {
		return m, nil
	}
	var view CollectionView
	if m.actions.Collection != nil {
		view = m.actions.Collection(context.Background(), kind)
	}

	lvl := menuLevel{title: meta.title}
	for _, e := range view.Entries {
		e := e
		lvl.items = append(lvl.items, menuItem{
			label: e.Title,
			desc:  e.Detail,
			run:   func(mm Model) (Model, tea.Cmd) { return mm.pushCollectionDetail(meta, e), nil },
		})
	}
	if len(view.Entries) == 0 {
		lvl.hint = firstNonEmpty(view.Empty, "Nothing here yet.")
		if view.Create != "" {
			lvl.hint += "  " + view.Create
		}
	}
	lvl.items = append(lvl.items, menuItem{
		label: "＋ Create a new " + meta.singular,
		desc:  "describe it — the agent will create it",
		run: func(mm Model) (Model, tea.Cmd) {
			mm = mm.closeMenu()
			mm.input = meta.seed
			return mm.sys("describe the " + meta.singular + " and press enter — the agent will create it"), nil
		},
	})
	return m.openMenu(lvl)
}

// pushCollectionDetail drills into one entry: its description, its backing file (info row), and an
// "edit via chat" action.
func (m Model) pushCollectionDetail(meta collectionMeta, e CollectionEntry) Model {
	lvl := menuLevel{title: e.Title, hint: e.Detail}
	// A runnable entry (a workflow recipe) leads with "Run": it closes the menu and sends the recipe
	// to the agent as a goal, echoing only a short label so the transcript stays readable.
	if e.Body != "" {
		lvl.items = append(lvl.items, menuItem{
			label: "▶ Run " + meta.singular,
			desc:  "send this recipe to the agent now",
			run: func(mm Model) (Model, tea.Cmd) {
				mm = mm.closeMenu()
				if mm.running {
					return mm, nil
				}
				mm.scrollOffset = 0
				mm.transcript = append(mm.transcript, entry{"user", "▶ " + meta.singular + ": " + e.Title})
				tm, cmd := mm.dispatchGoal(e.Body, mm.modeOrDefault())
				return tm.(Model), cmd
			},
		})
	}
	if e.Path != "" {
		lvl.items = append(lvl.items, menuItem{label: "File", desc: e.Path}) // info row (nil run)
	}
	lvl.items = append(lvl.items, menuItem{
		label: "Edit via chat",
		desc:  "describe a change — the agent edits it",
		run: func(mm Model) (Model, tea.Cmd) {
			mm = mm.closeMenu()
			mm.input = "Modify the " + meta.singular + " " + e.Title + " to "
			return mm.sys("describe the change and press enter — the agent will edit it"), nil
		},
	})
	return m.pushMenu(lvl)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// --- command handlers (repoint the collection slash commands to their menus) ---

func cmdSkills(m Model, _ string) (tea.Model, tea.Cmd)    { return m.openCollectionMenu("skills") }
func cmdMCP(m Model, _ string) (tea.Model, tea.Cmd)       { return m.openCollectionMenu("mcp") }
func cmdWorkflows(m Model, _ string) (tea.Model, tea.Cmd) { return m.openCollectionMenu("workflows") }
func cmdPlugins(m Model, _ string) (tea.Model, tea.Cmd)   { return m.openCollectionMenu("plugins") }
