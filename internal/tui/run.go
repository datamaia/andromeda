package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the interactive TUI program and blocks until the user quits. It requires a
// terminal; the Model's Update logic is unit-tested separately without one.
func Run(ctx context.Context, provider, model string, respond Responder) error {
	p := tea.NewProgram(New(provider, model, respond), tea.WithContext(ctx), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
