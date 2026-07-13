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

// WithModelSelect wires a callback that propagates a model choice to the driver, so the agent runs
// on the model the user actually picked (not just the provider's default).
func (m Model) WithModelSelect(onSelect func(string)) Model {
	m.onSelectModel = onSelect
	return m
}

// setModel records a model choice on both the view and (via the driver) the running agent.
func (m Model) setModel(id string) Model {
	m.model = id
	if m.onSelectModel != nil {
		m.onSelectModel(id)
	}
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
	m.pickerFilter = ""
	m.pickerTop = 0
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

// filteredItems returns the items matching the current type-to-filter query (case-insensitive
// substring over id and display). An empty query returns every item.
func (m Model) filteredItems() []pickerItem {
	if m.pickerFilter == "" {
		return m.pickerItems
	}
	q := strings.ToLower(m.pickerFilter)
	out := make([]pickerItem, 0, len(m.pickerItems))
	for _, it := range m.pickerItems {
		if strings.Contains(strings.ToLower(it.id), q) || strings.Contains(strings.ToLower(it.display), q) {
			out = append(out, it)
		}
	}
	return out
}

// pickerViewport is the number of item rows shown at once; the list scrolls within it so a huge
// catalogue (OpenRouter lists 400+ models) never overflows the screen or hides the cursor.
func (m Model) pickerViewport() int {
	h := m.height - 9 // title, filter line, error, help, status bar, padding
	if h < 5 {
		return 5
	}
	if h > 15 {
		return 15
	}
	return h
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

// openModelPicker discovers the provider's models off the UI thread (a slow or hanging provider
// must never freeze the interface) and opens the picker when they arrive. It always offers at least
// the current model so a choice can be made. Used during onboarding and by "/model".
func (m Model) openModelPicker() (tea.Model, tea.Cmd) {
	m.discovering = true
	m.state = "discovering models"
	act := m.actions.Models
	current := m.model
	return m, func() tea.Msg {
		var models []string
		if act != nil {
			models = act(context.Background())
		}
		if len(models) == 0 && current != "" {
			models = []string{current}
		}
		return modelsMsg{models: models}
	}
}

// showModelPicker opens the model list once discovery has returned.
func (m Model) showModelPicker(models []string) (tea.Model, tea.Cmd) {
	m.discovering = false
	if m.state == "discovering models" {
		m.state = "ready"
	}
	items := make([]pickerItem, 0, len(models))
	for _, id := range models {
		items = append(items, pickerItem{id: id, display: id})
	}
	m.pickerKind = "model"
	return m.openPicker("Select a model", items, m.model, func(mm Model, id string) (Model, error) {
		return mm.setModel(id), nil
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

// afterPick advances the selection sequence: choosing a provider always leads to the live model
// picker (during onboarding AND a mid-session /provider switch), so the user picks a currently
// available model instead of silently inheriting a possibly-retired catalog default. Choosing a
// model finishes onboarding or confirms the mid-session switch.
func (m Model) afterPick() (tea.Model, tea.Cmd) {
	switch m.pickerKind {
	case "provider":
		return m.openModelPicker()
	case "model":
		if m.onboarding {
			m.onboarding = false
			m.transcript = append(m.transcript, entry{"system",
				"ready · " + m.provider + " · " + m.model + " — type a goal, enter to send"})
			return m, nil
		}
		m.transcript = append(m.transcript, entry{"system", "now using " + m.provider + " · " + m.model})
		return m, nil
	default: // effort / theme / other one-shot pickers — the apply closure already applied + messaged
		return m, nil
	}
}

// handlePickerKey drives the picker: arrows (or j/k) move, enter selects, esc goes back. During
// onboarding esc on the provider list quits (a provider is required); esc on the model list steps
// back to provider choice.
func (m Model) handlePickerKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	items := m.filteredItems()
	switch {
	case msg.Code == tea.KeyEscape:
		if m.pickerFilter != "" {
			m.pickerFilter = "" // first esc clears the filter, revealing the full list again
			m.pickerCursor, m.pickerTop = 0, 0
			return m, nil
		}
		if m.onboarding {
			if m.pickerKind == "model" {
				return m.openProviderPicker() // step back to provider choice
			}
			return m, nil // a provider is required; exit is ctrl+c twice
		}
		m.pickerOpen = false
		m.pickerErr = ""
		return m, nil
	case msg.Code == tea.KeyUp:
		if m.pickerCursor > 0 {
			m.pickerCursor--
		}
		return m.clampViewport(), nil
	case msg.Code == tea.KeyDown:
		if m.pickerCursor < len(items)-1 {
			m.pickerCursor++
		}
		return m.clampViewport(), nil
	case msg.Code == tea.KeyEnter:
		if len(items) == 0 {
			if m.pickerFilter != "" { // no matches for the query — let the user edit it
				return m, nil
			}
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
		id := items[m.pickerCursor].id
		nm, err := m.pickerApply(m, id)
		if err != nil {
			m.pickerErr = err.Error() // stay open so the user can pick another
			return m, nil
		}
		return nm.afterPickerApply()
	case msg.Code == tea.KeyBackspace:
		if n := len(m.pickerFilter); n > 0 {
			m.pickerFilter = m.pickerFilter[:n-1]
			m.pickerCursor, m.pickerTop = 0, 0
		}
		return m, nil
	case msg.Text != "":
		// Any printable key narrows the list (type-to-filter), so long catalogues stay navigable.
		m.pickerFilter += msg.Text
		m.pickerCursor, m.pickerTop = 0, 0
		return m, nil
	}
	return m, nil
}

// clampViewport keeps the cursor within the visible window, scrolling the viewport as needed.
func (m Model) clampViewport() Model {
	vp := m.pickerViewport()
	if m.pickerCursor < m.pickerTop {
		m.pickerTop = m.pickerCursor
	}
	if m.pickerCursor >= m.pickerTop+vp {
		m.pickerTop = m.pickerCursor - vp + 1
	}
	if m.pickerTop < 0 {
		m.pickerTop = 0
	}
	return m
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

// renderPicker draws the active modal list: title, a type-to-filter line, a scrolling viewport of
// items with the cursor highlighted, scroll hints, and any selection error.
func (m Model) renderPicker() string {
	items := m.filteredItems()
	var b strings.Builder
	b.WriteString("\n  " + m.styles.Title.Render(m.pickerTitle) + "\n")

	// Filter line: show what's typed, plus the match count for context on long lists.
	filter := m.pickerFilter
	if filter == "" {
		filter = m.styles.Muted.Render("type to filter")
	}
	count := fmt.Sprintf("%d", len(items))
	if len(items) != len(m.pickerItems) {
		count = fmt.Sprintf("%d/%d", len(items), len(m.pickerItems))
	}
	b.WriteString("  " + m.styles.Muted.Render("search: ") + filter +
		m.styles.Muted.Render("   ("+count+")") + "\n\n")

	if len(items) == 0 {
		msg := "(nothing to choose)"
		if m.pickerFilter != "" {
			msg = "(no matches — backspace to widen)"
		}
		b.WriteString("    " + m.styles.Muted.Render(msg) + "\n")
	}

	// Viewport window around the cursor.
	vp := m.pickerViewport()
	top := m.pickerTop
	if top > len(items)-vp {
		top = len(items) - vp
	}
	if top < 0 {
		top = 0
	}
	end := top + vp
	if end > len(items) {
		end = len(items)
	}
	if top > 0 {
		b.WriteString("    " + m.styles.Muted.Render(fmt.Sprintf("↑ %d more", top)) + "\n")
	}
	for i := top; i < end; i++ {
		it := items[i]
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
	if end < len(items) {
		b.WriteString("    " + m.styles.Muted.Render(fmt.Sprintf("↓ %d more", len(items)-end)) + "\n")
	}

	b.WriteString("\n")
	if m.pickerErr != "" {
		b.WriteString("  " + m.styles.User.Render("⚠ "+m.pickerErr) + "\n\n")
	}
	b.WriteString("  " + m.styles.Muted.Render("↑/↓ move · type to filter · enter select · esc back") + "\n")
	return b.String()
}
