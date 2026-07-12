package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// The start-screen splash is the terminal-native rendering of the brand banner
// (docs/brand/banner-sketch.png, ADR-026): a front-facing sitting cat mascot with sparkle
// accents, the lowercase wordmark, and the tagline. The mascot art is pure ASCII so it
// renders in any terminal; the violet coloring and the four-pointed sparkles degrade to
// plain monochrome under the no-color tier (Volume 8 degradation tiers).

// mascot is the ASCII-art cat evoking the sketch: tall pointed ears, large round eyes,
// whiskers, and a seated body with two front paws. It is centered as a block (one shared
// indent) so the cat keeps its shape regardless of per-line width.
var mascot = []string{
	`     /\     /\`,
	`    /  \___/  \`,
	`   (           )`,
	`   (  (O) (O)  )`,
	` ~~(     Y     )~~`,
	`   (    \_/    )`,
	`    \         /`,
	`     \_______/`,
	`     /       \`,
	`    |  |   |  |`,
	`     \_|   |_/`,
	`     (_)   (_)`,
}

// scattered sparkles echo the sketch's stars: a few small ones around the cat, in the side
// gutters, plus the large one beside the wordmark. Keyed by mascot row index.
var (
	leftSparkleRow  = map[int]bool{1: true, 6: true, 9: true}
	rightSparkleRow = map[int]bool{3: true, 10: true}
)

// gutter is the blank margin reserved on each side of the mascot for scattered sparkles, so
// the cat itself stays centered on the given width.
const gutter = 6

// Splash renders the start-screen banner: the sparkled mascot, the wordmark, and the tagline,
// centered within width. It is shown while the session has no exchanges yet.
func (m Model) Splash(width int) string {
	if width <= 0 {
		width = m.width
	}
	big := m.styles.Title.Render("✦")
	small := m.styles.Muted.Render("✧")
	cat := m.styles.Agent // violet

	widest := 0
	for _, line := range mascot {
		if w := lipgloss.Width(line); w > widest {
			widest = w
		}
	}
	block := widest + gutter*2
	indent := 0
	if width > block {
		indent = (width - block) / 2
	}
	pad := strings.Repeat(" ", indent)

	var b strings.Builder
	b.WriteString("\n")
	for i, line := range mascot {
		var row strings.Builder
		row.WriteString(pad)
		// left gutter, optionally holding a small sparkle
		if leftSparkleRow[i] {
			row.WriteString(strings.Repeat(" ", gutter-2) + small + " ")
		} else {
			row.WriteString(strings.Repeat(" ", gutter))
		}
		row.WriteString(cat.Render(line))
		// right gutter, optionally holding a sparkle aligned past the widest line
		if rightSparkleRow[i] {
			row.WriteString(strings.Repeat(" ", widest-lipgloss.Width(line)+2) + small)
		}
		b.WriteString(strings.TrimRight(row.String(), " ") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(center(m.styles.Title.Render("andromeda")+"  "+big, width) + "\n")
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
