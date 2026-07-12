package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// ProviderChoice is one selectable entry in the provider menu.
type ProviderChoice struct {
	ID      string
	Display string
	Auth    string // env var holding the key, or "no key" for local providers
	Note    string
}

// ProviderSelectFunc applies a provider selection and returns the model now in use. It is wired
// by the composition root to rebuild the live provider, so the TUI stays free of provider imports.
type ProviderSelectFunc func(id string) (model string, err error)

// WithProviderMenu configures the provider picker. Without it, the ctrl+p menu is inert.
func (m Model) WithProviderMenu(choices []ProviderChoice, onSelect ProviderSelectFunc) Model {
	m.providers = choices
	m.onSelectProvider = onSelect
	return m
}

// openMenu shows the picker, starting the cursor on the current provider.
func (m Model) openMenu() (tea.Model, tea.Cmd) {
	m.menuOpen = true
	m.menuErr = ""
	m.menuCursor = 0
	for i, c := range m.providers {
		if c.ID == m.provider {
			m.menuCursor = i
			break
		}
	}
	return m, nil
}

// handleMenuKey drives the picker: arrows (or j/k) move, enter selects, esc goes back unchanged.
func (m Model) handleMenuKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Code == tea.KeyEscape:
		m.menuOpen = false
		m.menuErr = ""
	case msg.Code == tea.KeyUp || msg.Text == "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case msg.Code == tea.KeyDown || msg.Text == "j":
		if m.menuCursor < len(m.providers)-1 {
			m.menuCursor++
		}
	case msg.Code == tea.KeyEnter:
		choice := m.providers[m.menuCursor]
		if m.onSelectProvider != nil {
			model, err := m.onSelectProvider(choice.ID)
			if err != nil {
				m.menuErr = err.Error() // stay open so the user can pick another
				return m, nil
			}
			m.provider = choice.ID
			m.model = model
		}
		m.menuOpen = false
	}
	return m, nil
}

// renderMenu draws the provider picker with the cursor highlighted and any selection error.
func (m Model) renderMenu() string {
	var b strings.Builder
	b.WriteString("\n  " + m.styles.Title.Render("Select a provider") + "\n\n")
	for i, c := range m.providers {
		label := fmt.Sprintf("%-12s %-26s %s", c.ID, c.Display, c.Auth)
		if i == m.menuCursor {
			b.WriteString("  " + m.styles.Agent.Render("▸ "+label) + "\n")
		} else {
			b.WriteString("    " + m.styles.Muted.Render(label) + "\n")
		}
	}
	b.WriteString("\n")
	if m.menuErr != "" {
		b.WriteString("  " + m.styles.User.Render("⚠ "+m.menuErr) + "\n\n")
	}
	b.WriteString("  " + m.styles.Muted.Render("↑/↓ move · enter select · esc back") + "\n")
	return b.String()
}
