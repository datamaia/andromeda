package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// MemoryNote is one workspace memory note surfaced in the interactive /memory menu. The driver builds
// these from the file-based note store so this package stays free of filesystem imports.
type MemoryNote struct {
	ID      string
	Title   string
	Tags    []string
	Created string
	Preview string // first line of the body, if any
	Path    string // workspace-relative file path
}

// cmdMemory dispatches the text subcommands (/memory add|search|rm|list) to the Memory action, and
// opens the interactive CRUD menu on a bare /memory.
func cmdMemory(m Model, args string) (tea.Model, tea.Cmd) {
	if strings.TrimSpace(args) != "" {
		if m.actions.Memory == nil {
			return m.unavailable("memory"), nil
		}
		return m.sys(m.actions.Memory(context.Background(), args)), nil
	}
	return m.openMemoryMenu()
}

// openMemoryMenu lists the workspace notes (each drilling into a detail view with delete / edit) and
// offers add / search actions, with a friendly empty state.
func (m Model) openMemoryMenu() (tea.Model, tea.Cmd) {
	var notes []MemoryNote
	if m.actions.MemoryList != nil {
		notes = m.actions.MemoryList(context.Background())
	}
	lvl := menuLevel{title: "Memory · workspace notes"}
	for _, n := range notes {
		n := n
		lvl.items = append(lvl.items, menuItem{
			label: n.Title,
			desc:  memoryDesc(n),
			run:   func(mm Model) (Model, tea.Cmd) { return mm.pushMemoryDetail(n), nil },
		})
	}
	if len(notes) == 0 {
		lvl.hint = "No memories yet — Andromeda remembers facts here, alongside AGENTS.md."
	}
	lvl.items = append(lvl.items,
		menuItem{
			label: "＋ Add a note",
			desc:  "remember a fact (add #tags inline)",
			run: func(mm Model) (Model, tea.Cmd) {
				mm = mm.closeMenu()
				mm.input = "/memory add "
				return mm, nil
			},
		},
		menuItem{
			label: "🔍 Search",
			desc:  "find notes by title, tag, or text",
			run: func(mm Model) (Model, tea.Cmd) {
				mm = mm.closeMenu()
				mm.input = "/memory search "
				return mm, nil
			},
		},
	)
	return m.openMenu(lvl)
}

// pushMemoryDetail drills into one note: a preview, its file path, and edit / delete actions.
func (m Model) pushMemoryDetail(n MemoryNote) Model {
	lvl := menuLevel{title: n.Title, hint: memoryDesc(n)}
	if n.Preview != "" {
		lvl.items = append(lvl.items, menuItem{label: "Preview", desc: n.Preview})
	}
	if n.Path != "" {
		lvl.items = append(lvl.items, menuItem{label: "File", desc: n.Path})
	}
	lvl.items = append(lvl.items,
		menuItem{
			label: "Edit via chat",
			desc:  "describe a change — the agent edits it",
			run: func(mm Model) (Model, tea.Cmd) {
				mm = mm.closeMenu()
				mm.input = "Update memory note " + n.ID + " (" + n.Title + "): "
				return mm.sys("describe the change and press enter — the agent will edit the note"), nil
			},
		},
		menuItem{
			label: "Delete",
			desc:  "remove this note",
			run: func(mm Model) (Model, tea.Cmd) {
				if mm.actions.Memory != nil {
					mm = mm.sys(mm.actions.Memory(context.Background(), "rm "+n.ID))
				}
				nm, cmd := mm.openMemoryMenu() // reopen the refreshed list (the note is gone)
				return nm.(Model), cmd
			},
		},
	)
	return m.pushMenu(lvl)
}

// memoryDesc is the one-line descriptor for a note row: id · tags · date.
func memoryDesc(n MemoryNote) string {
	parts := []string{n.ID}
	if len(n.Tags) > 0 {
		parts = append(parts, strings.Join(n.Tags, ", "))
	}
	if n.Created != "" {
		parts = append(parts, n.Created)
	}
	return strings.Join(parts, " · ")
}
