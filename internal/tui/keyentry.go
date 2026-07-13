package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// ProviderKeyEnvFunc reports the environment variable a provider needs a key in, or "" when no key
// is required or one is already present. ProviderSetKeyFunc stores a pasted key for the session so
// the provider can be built. Both are injected by the driver.
type ProviderKeyEnvFunc func(id string) string

// ProviderSetKeyFunc records a pasted API key for a provider (e.g. into the process environment).
type ProviderSetKeyFunc func(id, key string)

// WithProviderKeyEntry wires in-TUI API-key entry: a provider whose key is missing prompts the user
// to paste it, then the key is stored and the provider activated.
func (m Model) WithProviderKeyEntry(need ProviderKeyEnvFunc, set ProviderSetKeyFunc) Model {
	m.providerKeyEnv = need
	m.setProviderKey = set
	return m
}

// handleKeyEntryKey drives the API-key prompt: typing builds the key, enter stores it and activates
// the provider, esc cancels (back to the provider picker during onboarding, else to the chat).
func (m Model) handleKeyEntryKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Code == tea.KeyEscape:
		id := m.keyProvider
		m.keyEntry = false
		m.keyInput = ""
		m.keyProvider = ""
		m.keyEnvName = ""
		if m.onboarding {
			return m.openProviderPicker()
		}
		m.transcript = append(m.transcript, entry{"system", "cancelled sign-in for " + id})
		return m, nil
	case msg.Code == tea.KeyEnter:
		key := strings.TrimSpace(m.keyInput)
		if key == "" {
			return m, nil // nothing pasted yet
		}
		id := m.keyProvider
		if m.setProviderKey != nil {
			m.setProviderKey(id, key)
		}
		m.keyEntry = false
		m.keyInput = ""
		m.keyProvider = ""
		m.keyEnvName = ""
		return m.activateProvider(id)
	case msg.Code == tea.KeyBackspace:
		if n := len(m.keyInput); n > 0 {
			m.keyInput = m.keyInput[:n-1]
		}
		return m, nil
	case msg.Text != "":
		m.keyInput += msg.Text
		return m, nil
	}
	return m, nil
}

// renderKeyEntry draws the API-key paste prompt with the key masked.
func (m Model) renderKeyEntry() string {
	var b strings.Builder
	b.WriteString("\n  " + m.styles.Title.Render("Sign in to "+m.keyProvider) + "\n\n")
	b.WriteString("  " + m.styles.Muted.Render("Paste your API key for "+m.keyEnvName+" and press enter.") + "\n")
	b.WriteString("  " + m.styles.Muted.Render("It is kept only for this session; nothing is written to disk.") + "\n\n")
	masked := strings.Repeat("•", len([]rune(m.keyInput)))
	b.WriteString("  " + m.styles.Prompt.Render("key ❯ ") + masked + "▏\n")
	b.WriteString("\n  " + m.styles.Muted.Render("enter save · esc cancel") + "\n")
	return b.String()
}
