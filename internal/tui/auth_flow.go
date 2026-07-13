package tui

import (
	tea "charm.land/bubbletea/v2"
)

// AuthEvent is one step of an interactive provider sign-in (browser OAuth). The flow emits the
// browser URL first (so the UI can show it), then a terminal Done or Err.
type AuthEvent struct {
	URL  string // the browser URL to visit (emitted first)
	Done bool   // sign-in completed and the credential was stored
	Err  error  // sign-in failed
}

// ProviderAuthFunc starts interactive sign-in for a provider and returns a stream of progress, or
// nil when the provider needs no interactive auth (already signed in, or a key/keyless provider).
// Injected by the driver so the TUI keeps no auth imports.
type ProviderAuthFunc func(id string) <-chan AuthEvent

// WithProviderAuth wires interactive provider sign-in (e.g. ChatGPT OAuth).
func (m Model) WithProviderAuth(f ProviderAuthFunc) Model {
	m.providerAuth = f
	return m
}

// authEventMsg carries one AuthEvent (or a closed-stream signal) back into Update.
type authEventMsg struct {
	ev     AuthEvent
	closed bool
}

// waitAuth returns a command that reads the next event from a sign-in stream.
func waitAuth(ch <-chan AuthEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return authEventMsg{closed: true}
		}
		return authEventMsg{ev: ev}
	}
}

// handleAuthEvent advances an interactive sign-in: it surfaces the URL, then on success activates
// the provider (and continues onboarding), or reports failure.
func (m Model) handleAuthEvent(msg authEventMsg) (tea.Model, tea.Cmd) {
	if msg.closed {
		if m.authing {
			m.authing = false
			m.authEvents = nil
			m.state = "ready"
		}
		return m, nil
	}
	ev := msg.ev
	switch {
	case ev.URL != "":
		m.authURL = ev.URL
		m.transcript = append(m.transcript, entry{"system",
			"Opening your browser to sign in. If it doesn't open, visit:\n  " + ev.URL})
		return m, waitAuth(m.authEvents) // keep waiting for completion
	case ev.Err != nil:
		m.authing = false
		m.authEvents = nil
		m.state = "ready"
		m.transcript = append(m.transcript, entry{"system", "sign-in failed: " + ev.Err.Error()})
		if m.onboarding {
			return m.openProviderPicker() // let the user choose again
		}
		return m, nil
	case ev.Done:
		m.authing = false
		m.authEvents = nil
		m.state = "ready"
		m.transcript = append(m.transcript, entry{"system", "signed in ✓"})
		return m.activateProvider(m.authProvider)
	}
	return m, waitAuth(m.authEvents)
}
