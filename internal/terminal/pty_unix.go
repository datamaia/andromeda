//go:build unix

package terminal

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// ptySupported reports whether PTY allocation is available (Unix).
func ptySupported() bool { return true }

// ptyStart starts cmd attached to a new pseudoterminal and returns the master file.
func ptyStart(cmd *exec.Cmd) (*os.File, error) {
	return pty.Start(cmd)
}

// ptySetsize sets the PTY window size.
func ptySetsize(f *os.File, cols, rows int) error {
	return pty.Setsize(f, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
}
