package main

import "os"

// interactive reports whether this process is attached to a usable interactive terminal,
// per FR-CLI-003 branch 1 (ADR-102 interactivity resolution). It requires both standard
// streams to be character devices (a TTY, not a pipe/file/CI buffer) and a capable TERM —
// an unset or "dumb" TERM resolves non-interactive so the full-screen TUI never takes over
// a stream that cannot render it.
func interactive() bool {
	switch os.Getenv("TERM") {
	case "", "dumb":
		return false
	}
	return isCharDevice(os.Stdin) && isCharDevice(os.Stdout)
}

// isCharDevice reports whether f is a character device (a terminal).
func isCharDevice(f *os.File) bool {
	if f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
