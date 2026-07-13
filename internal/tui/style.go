package tui

import "charm.land/lipgloss/v2"

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

	// markdown rendering
	Heading   lipgloss.Style // # headings
	Bold      lipgloss.Style // **strong**
	Code      lipgloss.Style // `inline code`
	CodeBlock lipgloss.Style // fenced ``` blocks
	CodeLang  lipgloss.Style // the language tag above a fenced block
	Comment   lipgloss.Style // code comments (light syntax highlight)
	Tool      lipgloss.Style // inline tool-call cards
}

// DefaultStyles returns the brand styles.
func DefaultStyles() Styles {
	primary := lipgloss.Color(ColorPrimary)
	tertiary := lipgloss.Color(ColorTertiary)
	taupe := lipgloss.Color(ColorSecondary)
	return Styles{
		Title:     lipgloss.NewStyle().Foreground(primary).Bold(true),
		User:      lipgloss.NewStyle().Foreground(tertiary).Bold(true),
		Agent:     lipgloss.NewStyle().Foreground(primary),
		Prompt:    lipgloss.NewStyle().Foreground(primary).Bold(true),
		StatusBar: lipgloss.NewStyle().Foreground(lipgloss.Color(ColorNeutral)).Background(primary).Padding(0, 1),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDanger)),
		Muted:     lipgloss.NewStyle().Foreground(taupe),

		Heading:   lipgloss.NewStyle().Foreground(primary).Bold(true),
		Bold:      lipgloss.NewStyle().Foreground(tertiary).Bold(true),
		Code:      lipgloss.NewStyle().Foreground(tertiary).Background(lipgloss.Color("#2A2A33")),
		CodeBlock: lipgloss.NewStyle().Foreground(tertiary).Background(lipgloss.Color("#1B1B22")).Padding(0, 1),
		CodeLang:  lipgloss.NewStyle().Foreground(taupe).Italic(true),
		Comment:   lipgloss.NewStyle().Foreground(taupe).Italic(true),
		Tool:      lipgloss.NewStyle().Foreground(taupe),
	}
}

// LightStyles is the light-background variant selected by `/theme light`. It keeps the brand violet
// accent but swaps text/surface colors for legibility on a light terminal; the wordmark gradient
// and the violet primary are unchanged.
func LightStyles() Styles {
	primary := lipgloss.Color(ColorPrimary)    // brand violet (readable on light too)
	ink := lipgloss.Color("#1B1B22")           // near-black body text
	taupe := lipgloss.Color("#6B5D52")         // darker taupe for contrast on light
	neutralBg := lipgloss.Color(ColorTertiary) // off-white status-bar text over violet
	return Styles{
		Title:     lipgloss.NewStyle().Foreground(primary).Bold(true),
		User:      lipgloss.NewStyle().Foreground(ink).Bold(true),
		Agent:     lipgloss.NewStyle().Foreground(primary),
		Prompt:    lipgloss.NewStyle().Foreground(primary).Bold(true),
		StatusBar: lipgloss.NewStyle().Foreground(neutralBg).Background(primary).Padding(0, 1),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#C0392B")),
		Muted:     lipgloss.NewStyle().Foreground(taupe),

		Heading:   lipgloss.NewStyle().Foreground(primary).Bold(true),
		Bold:      lipgloss.NewStyle().Foreground(ink).Bold(true),
		Code:      lipgloss.NewStyle().Foreground(ink).Background(lipgloss.Color("#ECE8E1")),
		CodeBlock: lipgloss.NewStyle().Foreground(ink).Background(lipgloss.Color("#F0ECE4")).Padding(0, 1),
		CodeLang:  lipgloss.NewStyle().Foreground(taupe).Italic(true),
		Comment:   lipgloss.NewStyle().Foreground(taupe).Italic(true),
		Tool:      lipgloss.NewStyle().Foreground(taupe),
	}
}
