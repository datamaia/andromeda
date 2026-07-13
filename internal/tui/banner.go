package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// The start-screen splash (ADR-026) is the small minimalist cat mascot above the ANDROMEDA
// wordmark and the tagline, in the brand palette (violet mascot + letters, off-white sparkles,
// taupe tagline). The art is pure ASCII so it renders anywhere; it degrades to a plain centered
// wordmark on terminals too narrow for the block letters, and to monochrome under no-color. The
// large detailed cat is reserved for the `about` easter egg (see bigcat.go).

// smallCat is the minimalist mascot shown on every start screen. All three lines are the same
// visible width so they stay vertically aligned when each is centered (the ears sit over the face).
var smallCat = []string{
	` /\_/\ `,
	`( o.o )`,
	` > ^ < `,
}

// wordmark is "ANDROMEDA" as bold block letters (the "ANSI Shadow" style: solid blocks with
// box-drawing edges). Every row is padded to the same width so it renders as a clean rectangle.
var wordmark = []string{
	` █████╗ ███╗   ██╗██████╗ ██████╗  ██████╗ ███╗   ███╗███████╗██████╗  █████╗ `,
	`██╔══██╗████╗  ██║██╔══██╗██╔══██╗██╔═══██╗████╗ ████║██╔════╝██╔══██╗██╔══██╗`,
	`███████║██╔██╗ ██║██║  ██║██████╔╝██║   ██║██╔████╔██║█████╗  ██║  ██║███████║`,
	`██╔══██║██║╚██╗██║██║  ██║██╔══██╗██║   ██║██║╚██╔╝██║██╔══╝  ██║  ██║██╔══██║`,
	`██║  ██║██║ ╚████║██████╔╝██║  ██║╚██████╔╝██║ ╚═╝ ██║███████╗██████╔╝██║  ██║`,
	`╚═╝  ╚═╝╚═╝  ╚═══╝╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝╚═════╝ ╚═╝  ╚═╝`,
}

// wordmarkWidth is the width of the widest wordmark line.
const wordmarkWidth = 78

// wordmarkGradient is the top-to-bottom violet ramp painted per wordmark row, all within the brand
// family (a light tint above fading to a deeper shade below). One hex color per wordmark line.
var wordmarkGradient = []string{
	"#A78BFF",
	"#9575FF",
	ColorPrimary, // #7C5CFF, the brand violet
	"#6B4BF5",
	"#5B3DE0",
	"#4E33C7",
}

// Splash renders the start-screen banner centered within width.
func (m Model) Splash(width int) string {
	if width <= 0 {
		width = m.width
	}
	cat := m.styles.Agent // violet

	var b strings.Builder
	b.WriteString("\n")
	// minimalist cat, centered as a block (equal-width lines keep the ears over the face)
	for _, line := range smallCat {
		b.WriteString(center(cat.Render(line), width) + "\n")
	}
	b.WriteString("\n")
	// ANDROMEDA wordmark (block letters with a violet gradient, or a plain fallback when the
	// terminal is too narrow for the art).
	if width < wordmarkWidth+2 {
		b.WriteString(center(m.styles.Title.Render("a n d r o m e d a"), width) + "\n")
	} else {
		indent := strings.Repeat(" ", (width-wordmarkWidth)/2)
		for i, line := range wordmark {
			c := lipgloss.Color(wordmarkGradient[i%len(wordmarkGradient)])
			b.WriteString(indent + lipgloss.NewStyle().Foreground(c).Render(line) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(center(m.styles.Muted.Render(Tagline), width) + "\n")
	return b.String()
}

// center pads s so its visible content sits roughly in the middle of width. It measures the
// printable width (ignoring ANSI styling) so styled and plain strings center alike.
func center(s string, width int) string {
	vis := lipgloss.Width(s)
	if vis >= width {
		return s
	}
	pad := (width - vis) / 2
	return strings.Repeat(" ", pad) + s
}
