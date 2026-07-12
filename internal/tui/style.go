package tui

import "github.com/charmbracelet/lipgloss"

// Brand design tokens (ADR-026).
const (
	ColorPrimary   = "#7C5CFF" // violet accent
	ColorSecondary = "#8C7B6E" // warm taupe
	ColorTertiary  = "#F5F2ED" // warm off-white
	ColorNeutral   = "#121417" // near-black
	ColorDanger    = "#FF6B6B" // accessible soft red (fixed in Volume 8)
)

// Styles holds the lipgloss styles derived from the brand tokens.
type Styles struct {
	Title     lipgloss.Style
	User      lipgloss.Style
	Agent     lipgloss.Style
	Prompt    lipgloss.Style
	StatusBar lipgloss.Style
	Error     lipgloss.Style
	Muted     lipgloss.Style
}

// DefaultStyles returns the brand styles.
func DefaultStyles() Styles {
	primary := lipgloss.Color(ColorPrimary)
	return Styles{
		Title:     lipgloss.NewStyle().Foreground(primary).Bold(true),
		User:      lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTertiary)).Bold(true),
		Agent:     lipgloss.NewStyle().Foreground(primary),
		Prompt:    lipgloss.NewStyle().Foreground(primary).Bold(true),
		StatusBar: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorNeutral)).Background(primary).Padding(0, 1),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDanger)),
		Muted:     lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondary)),
	}
}
