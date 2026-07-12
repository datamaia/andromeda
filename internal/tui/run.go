package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

// Run starts a default interactive TUI session and blocks until the user quits.
func Run(ctx context.Context, provider, model string, respond Responder) error {
	return RunModel(ctx, New(provider, model, respond))
}

// RunModel starts the interactive TUI program for a fully-configured Model (e.g. one carrying a
// provider menu) and blocks until the user quits. It requires a terminal; the Model's Update
// logic is unit-tested separately without one. The alternate screen is requested declaratively by
// the Model's View (Bubble Tea v2, ADR-006).
func RunModel(ctx context.Context, m Model) error {
	p := tea.NewProgram(m, tea.WithContext(ctx))
	_, err := p.Run()
	return err
}
