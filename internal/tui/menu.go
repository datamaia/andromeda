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

// pickerItem is one row in a modal selection list (providers, models, …).
type pickerItem struct {
	id      string
	display string
	note    string // right-aligned hint (env var, capability, …)
}

// openPicker opens a modal selection list titled title over items, starting the cursor on current.
// apply performs the side effect of a choice and returns the updated Model, or an error to display
// while keeping the picker open so the user can choose again.
func (m Model) openPicker(title string, items []pickerItem, current string, apply func(Model, string) (Model, error)) (tea.Model, tea.Cmd) {
	m.pickerOpen = true
	m.pickerTitle = title
	m.pickerItems = items
	m.pickerErr = ""
	m.pickerCursor = 0
	for i, it := range items {
		if it.id == current {
			m.pickerCursor = i
			break
		}
	}
	m.pickerApply = apply
	return m, nil
}

// openProviderPicker opens the provider list (ctrl+p / "/provider"), rebuilding the live provider
// and adopting its default model on selection.
func (m Model) openProviderPicker() (tea.Model, tea.Cmd) {
	items := make([]pickerItem, 0, len(m.providers))
	for _, c := range m.providers {
		items = append(items, pickerItem{id: c.ID, display: c.Display, note: c.Auth})
	}
	return m.openPicker("Select a provider", items, m.provider, func(mm Model, id string) (Model, error) {
		if mm.onSelectProvider == nil {
			mm.provider = id
			return mm, nil
		}
		model, err := mm.onSelectProvider(id)
		if err != nil {
			return mm, err
		}
		mm.provider = id
		mm.model = model
		return mm, nil
	})
}

// handlePickerKey drives the picker: arrows (or j/k) move, enter selects, esc goes back unchanged.
func (m Model) handlePickerKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Code == tea.KeyEscape:
		m.pickerOpen = false
		m.pickerErr = ""
	case msg.Code == tea.KeyUp || msg.Text == "k":
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
	case msg.Code == tea.KeyDown || msg.Text == "j":
		if m.pickerCursor < len(m.pickerItems)-1 {
			m.pickerCursor++
		}
	case msg.Code == tea.KeyEnter:
		if len(m.pickerItems) == 0 {
			m.pickerOpen = false
			return m, nil
		}
		id := m.pickerItems[m.pickerCursor].id
		if m.pickerApply != nil {
			nm, err := m.pickerApply(m, id)
			if err != nil {
				m.pickerErr = err.Error() // stay open so the user can pick another
				return m, nil
			}
			nm.pickerOpen = false
			nm.pickerErr = ""
			return nm, nil
		}
		m.pickerOpen = false
	}
	return m, nil
}

// renderPicker draws the active modal list with the cursor highlighted and any selection error.
func (m Model) renderPicker() string {
	var b strings.Builder
	b.WriteString("\n  " + m.styles.Title.Render(m.pickerTitle) + "\n\n")
	if len(m.pickerItems) == 0 {
		b.WriteString("    " + m.styles.Muted.Render("(nothing to choose)") + "\n\n")
	}
	for i, it := range m.pickerItems {
		label := it.display
		if it.note != "" {
			label = fmt.Sprintf("%-30s %s", it.display, it.note)
		}
		if i == m.pickerCursor {
			b.WriteString("  " + m.styles.Agent.Render("▸ "+label) + "\n")
		} else {
			b.WriteString("    " + m.styles.Muted.Render(label) + "\n")
		}
	}
	b.WriteString("\n")
	if m.pickerErr != "" {
		b.WriteString("  " + m.styles.User.Render("⚠ "+m.pickerErr) + "\n\n")
	}
	b.WriteString("  " + m.styles.Muted.Render("↑/↓ move · enter select · esc back") + "\n")
	return b.String()
}
