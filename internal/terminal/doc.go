// Package terminal is layer L3: the Terminal Engine implementing ports.TerminalPort (Volume 6).
// This MVP uses pipe-based execution with streaming capture, stdin write, and signals; PTY mode
// and full sandbox integration are later increments. Output capture is bounded and truncation
// is marked, never silent. Callers hold a permission decision before running commands; the
// terminal.run built-in tool routes through the Tool Runtime for that mediation.
package terminal
