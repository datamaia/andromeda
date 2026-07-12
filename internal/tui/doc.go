// Package tui is layer L5: the terminal user interface (Volume 8). It presents an interactive
// session — a scrollable transcript, a prompt input, and a status bar — styled with the brand
// design tokens (ADR-026). The Model is a Bubble Tea model whose Update logic is unit-testable
// headlessly; only the program runner needs a terminal.
//
// Deviation from ADR-006: this MVP uses Bubble Tea v1 (github.com/charmbracelet/bubbletea) for a
// stable, working implementation; migration to the v2 charm.land stack is tracked in STATUS.md.
package tui
