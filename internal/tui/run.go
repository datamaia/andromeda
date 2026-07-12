package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

// Run starts the interactive TUI program and blocks until the user quits. It requires a
// terminal; the Model's Update logic is unit-tested separately without one. The alternate screen
// is requested declaratively by the Model's View (Bubble Tea v2, ADR-006).
func Run(ctx context.Context, provider, model string, respond Responder) error {
	p := tea.NewProgram(New(provider, model, respond), tea.WithContext(ctx))
	_, err := p.Run()
	return err
}
