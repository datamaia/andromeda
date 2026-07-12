package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// The start-screen splash is the terminal-native rendering of the brand banner
// (docs/brand/banner-sketch.png, ADR-026): the ASCII cat mascot with sparkle accents, the
// lowercase wordmark, and the tagline. Glyphs are ASCII-safe so it renders in any terminal; the
// styling degrades to plain text with no color (Volume 8 degradation tiers).

// mascot is the ASCII-art cat — a front-facing sitting cat evoking the banner sketch.
var mascot = []string{
	`      /\_/\`,
	`     ( o.o )`,
	`     =( " )=`,
	`      )   (`,
	`     (  Y  )`,
	`     /_/ \_\`,
}

// Splash renders the start-screen banner: sparkles, the mascot, the wordmark, and the tagline,
// centered within width. It is shown while the session has no exchanges yet. The mascot is
// centered as a block (one shared indent) so the cat keeps its shape regardless of per-line width.
func (m Model) Splash(width int) string {
	if width <= 0 {
		width = m.width
	}
	star := m.styles.Title.Render("✦")
	cat := m.styles.Agent // violet

	// Block indent from the widest mascot line, so every line shares the same left margin.
	widest := 0
	for _, line := range mascot {
		if len(line) > widest {
			widest = len(line)
		}
	}
	indent := 0
	if width > widest {
		indent = (width - widest) / 2
	}
	pad := strings.Repeat(" ", indent)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(pad + "        " + star + "\n") // sparkle above, over the right ear
	for i, line := range mascot {
		row := pad + cat.Render(line)
		if i == 1 {
			row += "      " + star // sparkle beside the face
		}
		b.WriteString(row + "\n")
	}
	b.WriteString("\n")
	b.WriteString(center(m.styles.Title.Render("andromeda")+" "+star, width) + "\n")
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
