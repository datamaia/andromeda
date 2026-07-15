package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// A navigable command menu is a stack of levels the user can drill into (enter / →) and back out of
// (esc / ←), with a breadcrumb across the top. Each row carries a description, and a level can show
// an empty-state hint. It is deliberately separate from the provider/model picker (menu.go), whose
// apply/afterPick flow is entangled with onboarding; command menus stay simple and self-contained.

// menuItem is one selectable row. run performs the action and returns the next model (+ optional
// command); it typically calls m.pushMenu to drill in, m.closeMenu to finish, or seeds m.input. A
// nil run marks a non-actionable info row (e.g. a file path).
type menuItem struct {
	label string
	desc  string
	run   func(m Model) (Model, tea.Cmd)
}

// menuLevel is one screen in the navigation stack: a titled list with its own cursor and an optional
// hint (context or empty-state) rendered under the title.
type menuLevel struct {
	title  string
	hint   string
	items  []menuItem
	cursor int
}

// openMenu opens a fresh navigable menu, replacing any existing stack.
func (m Model) openMenu(level menuLevel) (tea.Model, tea.Cmd) {
	m.menu = []menuLevel{level}
	return m, nil
}

// pushMenu drills into a submenu, preserving the breadcrumb back to the current level.
func (m Model) pushMenu(level menuLevel) Model {
	stack := make([]menuLevel, 0, len(m.menu)+1)
	stack = append(stack, m.menu...)
	m.menu = append(stack, level)
	return m
}

// closeMenu dismisses the entire menu stack.
func (m Model) closeMenu() Model {
	m.menu = nil
	return m
}

// menuOpen reports whether a navigable command menu is on screen.
func (m Model) menuOpen() bool { return len(m.menu) > 0 }

// handleMenuKey drives menu navigation: ↑/↓ move, enter/→ activate, esc/← go back one level (closing
// at the root). It copies the mutated level back so it never aliases a retained prior model.
func (m Model) handleMenuKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	i := len(m.menu) - 1
	top := m.menu[i]
	switch msg.Code {
	case tea.KeyEscape, tea.KeyLeft:
		m.menu = append([]menuLevel(nil), m.menu[:i]...) // pop without aliasing
		return m, nil
	case tea.KeyUp:
		if top.cursor > 0 {
			top.cursor--
		}
	case tea.KeyDown:
		if top.cursor < len(top.items)-1 {
			top.cursor++
		}
	case tea.KeyEnter, tea.KeyRight:
		if len(top.items) == 0 {
			return m, nil
		}
		it := top.items[clamp(top.cursor, len(top.items))]
		if it.run == nil {
			return m, nil // non-actionable info row
		}
		return it.run(m)
	default:
		return m, nil
	}
	stack := append([]menuLevel(nil), m.menu...)
	stack[i] = top
	m.menu = stack
	return m, nil
}

// renderMenu draws the active level: breadcrumb, title, optional hint, the rows (label + dimmed
// description) in a scrolling window, and a key legend.
func (m Model) renderMenu() string {
	if !m.menuOpen() {
		return ""
	}
	top := m.menu[len(m.menu)-1]
	var b strings.Builder
	b.WriteString("\n  " + m.styles.Title.Render(top.title) + "\n")
	if len(m.menu) > 1 {
		crumbs := make([]string, len(m.menu))
		for i, lv := range m.menu {
			crumbs[i] = lv.title
		}
		b.WriteString("  " + m.styles.Muted.Render(strings.Join(crumbs, " › ")) + "\n")
	}
	if top.hint != "" {
		b.WriteString("  " + m.styles.Muted.Render(top.hint) + "\n")
	}
	b.WriteString("\n")

	if len(top.items) == 0 {
		b.WriteString("    " + m.styles.Muted.Render("(empty — esc to go back)") + "\n\n")
	} else {
		labels := make([]string, len(top.items))
		for i, it := range top.items {
			if it.desc != "" {
				labels[i] = fmt.Sprintf("%-26s %s", it.label, it.desc)
			} else {
				labels[i] = it.label
			}
		}
		b.WriteString(m.renderMenuRows(labels, clamp(top.cursor, len(top.items))))
		b.WriteString("\n")
	}

	back := "esc close"
	if len(m.menu) > 1 {
		back = "esc/← back"
	}
	b.WriteString("  " + m.styles.Muted.Render("↑/↓ move · enter select · "+back) + "\n")
	return b.String()
}
