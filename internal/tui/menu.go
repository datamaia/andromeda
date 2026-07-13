package tui

import (
	"context"
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

// openProviderPicker opens the provider list (ctrl+p / "/provider"). Selecting a provider either
// switches immediately, or — when the provider needs interactive setup — defers to a browser
// sign-in (OAuth) or an API-key prompt before activating. The apply closure only records what is
// needed (pendingAuth / pendingKeyEnv); handlePickerKey starts the matching flow.
func (m Model) openProviderPicker() (tea.Model, tea.Cmd) {
	items := make([]pickerItem, 0, len(m.providers))
	for _, c := range m.providers {
		items = append(items, pickerItem{id: c.ID, display: c.Display, note: c.Auth})
	}
	m.pickerKind = "provider"
	return m.openPicker("Select a provider", items, m.provider, func(mm Model, id string) (Model, error) {
		// Browser sign-in (e.g. ChatGPT) — defer the switch until sign-in completes.
		if mm.providerAuth != nil {
			if ch := mm.providerAuth(id); ch != nil {
				mm.authEvents = ch
				mm.pendingAuth = id
				return mm, nil
			}
		}
		// Missing API key — prompt the user to paste it, then activate.
		if mm.providerKeyEnv != nil {
			if env := mm.providerKeyEnv(id); env != "" {
				mm.pendingKeyEnv = env
				mm.pendingKeyProvider = id
				return mm, nil
			}
		}
		// Ready to use — switch now. An error here keeps the picker open so another can be chosen.
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

// openModelPicker opens the model list for the current provider, discovering models when possible
// and always offering at least the current model so a choice can be made (used during onboarding
// and by "/model").
func (m Model) openModelPicker() (tea.Model, tea.Cmd) {
	var models []string
	if m.actions.Models != nil {
		models = m.actions.Models(context.Background())
	}
	if len(models) == 0 && m.model != "" {
		models = []string{m.model}
	}
	items := make([]pickerItem, 0, len(models))
	for _, id := range models {
		items = append(items, pickerItem{id: id, display: id})
	}
	m.pickerKind = "model"
	return m.openPicker("Select a model", items, m.model, func(mm Model, id string) (Model, error) {
		mm.model = id
		return mm, nil
	})
}

// activateProvider builds and switches to a provider once it is usable (signed in / key present),
// then continues onboarding (to the model picker) when appropriate.
func (m Model) activateProvider(id string) (tea.Model, tea.Cmd) {
	if m.onSelectProvider != nil {
		model, err := m.onSelectProvider(id)
		if err != nil {
			m.transcript = append(m.transcript, entry{"system", "could not use " + id + ": " + err.Error()})
			if m.onboarding {
				return m.openProviderPicker()
			}
			return m, nil
		}
		m.provider = id
		m.model = model
	} else {
		m.provider = id
	}
	return m.afterPick()
}

// afterPick advances the onboarding sequence: after a provider comes the model picker; after a
// model, onboarding is complete. Outside onboarding it is a no-op (a plain switch).
func (m Model) afterPick() (tea.Model, tea.Cmd) {
	if !m.onboarding {
		return m, nil
	}
	switch m.pickerKind {
	case "provider":
		return m.openModelPicker()
	default: // "model"
		m.onboarding = false
		m.transcript = append(m.transcript, entry{"system",
			"ready · " + m.provider + " · " + m.model + " — type a goal, enter to send"})
		return m, nil
	}
}

// handlePickerKey drives the picker: arrows (or j/k) move, enter selects, esc goes back. During
// onboarding esc on the provider list quits (a provider is required); esc on the model list steps
// back to provider choice.
func (m Model) handlePickerKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Code == tea.KeyEscape:
		if m.onboarding {
			if m.pickerKind == "model" {
				return m.openProviderPicker()
			}
			m.quitting = true
			return m, tea.Quit
		}
		m.pickerOpen = false
		m.pickerErr = ""
		return m, nil
	case msg.Code == tea.KeyUp || msg.Text == "k":
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
		return m, nil
	case msg.Code == tea.KeyDown || msg.Text == "j":
		if m.pickerCursor < len(m.pickerItems)-1 {
			m.pickerCursor++
		}
		return m, nil
	case msg.Code == tea.KeyEnter:
		if len(m.pickerItems) == 0 {
			if m.onboarding && m.pickerKind == "model" {
				m.transcript = append(m.transcript, entry{"system",
					"no models available for " + m.provider + " — choose another provider"})
				return m.openProviderPicker()
			}
			m.pickerOpen = false
			return m, nil
		}
		if m.pickerApply == nil {
			m.pickerOpen = false
			return m, nil
		}
		id := m.pickerItems[m.pickerCursor].id
		nm, err := m.pickerApply(m, id)
		if err != nil {
			m.pickerErr = err.Error() // stay open so the user can pick another
			return m, nil
		}
		return nm.afterPickerApply()
	}
	return m, nil
}

// afterPickerApply closes the picker and dispatches whatever the selection deferred: a browser
// sign-in, an API-key prompt, provider activation, or the next onboarding step.
func (m Model) afterPickerApply() (tea.Model, tea.Cmd) {
	m.pickerOpen = false
	m.pickerErr = ""
	switch {
	case m.pendingAuth != "":
		m.authProvider = m.pendingAuth
		m.pendingAuth = ""
		m.authing = true
		m.authURL = ""
		m.state = "signing in"
		return m, waitAuth(m.authEvents)
	case m.pendingKeyEnv != "":
		m.keyEntry = true
		m.keyProvider = m.pendingKeyProvider
		m.keyEnvName = m.pendingKeyEnv
		m.keyInput = ""
		m.pendingKeyEnv = ""
		m.pendingKeyProvider = ""
		return m, nil
	default:
		// A provider chosen inline (or a model) — advance the onboarding sequence.
		return m.afterPick()
	}
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
