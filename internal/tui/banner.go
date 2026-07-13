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

// smallCat is the minimalist mascot shown on every start screen.
var smallCat = []string{
	`  /\_/\`,
	` ( o.o )`,
	`  > ^ <`,
}

// wordmark is "ANDROMEDA" as block letters (figlet "standard").
var wordmark = []string{
	`    _    _   _ ____  ____   ___  __  __ _____ ____    _`,
	`   / \  | \ | |  _ \|  _ \ / _ \|  \/  | ____|  _ \  / \`,
	`  / _ \ |  \| | | | | |_) | | | | |\/| |  _| | | | |/ _ \`,
	` / ___ \| |\  | |_| |  _ <| |_| | |  | | |___| |_| / ___ \`,
	`/_/   \_\_| \_|____/|_| \_\\___/|_|  |_|_____|____/_/   \_\`,
}

// wordmarkWidth is the width of the widest wordmark line.
const wordmarkWidth = 59

// Splash renders the start-screen banner centered within width.
func (m Model) Splash(width int) string {
	if width <= 0 {
		width = m.width
	}
	sparkle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTertiary)).Render("✦")
	cat := m.styles.Agent // violet

	var b strings.Builder
	b.WriteString("\n")
	// minimalist cat, centered, with a sparkle beside its ear
	for i, line := range smallCat {
		row := cat.Render(line)
		if i == 0 {
			row += "   " + sparkle
		}
		b.WriteString(center(row, width) + "\n")
	}
	b.WriteString("\n")
	// ANDROMEDA wordmark (block letters, or a plain fallback when the terminal is too narrow)
	if width < wordmarkWidth+2 {
		b.WriteString(center(m.styles.Title.Render("a n d r o m e d a"), width) + "\n")
	} else {
		indent := strings.Repeat(" ", (width-wordmarkWidth)/2)
		for _, line := range wordmark {
			b.WriteString(indent + m.styles.Title.Render(line) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(center(m.styles.Muted.Render(Tagline)+" "+sparkle, width) + "\n")
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
