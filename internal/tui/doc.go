// Package tui is layer L5: the terminal user interface (Volume 8). It presents an interactive
// session — a start-screen splash (the ASCII cat mascot, wordmark, and tagline, ADR-026), a
// scrollable transcript, a prompt input, and a status bar — styled with the brand design tokens.
// The Model is a Bubble Tea model whose Update/render logic is unit-testable headlessly; only the
// program runner needs a terminal.
//
// Per ADR-006 this uses the Charm v2 stack via the charm.land vanity import paths
// (charm.land/bubbletea/v2, charm.land/lipgloss/v2). v2 keyboard input is tea.KeyPressMsg, View
// returns a tea.View, and the alternate screen is requested declaratively on that View. The v1
// github.com/charmbracelet/* paths MUST NOT appear in the tree (enforced by make lint-charm).
package tui
