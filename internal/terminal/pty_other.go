//go:build !unix

package terminal

import (
	"errors"
	"os"
	"os/exec"
)

// ptySupported reports whether PTY allocation is available (not on non-Unix here).
func ptySupported() bool { return false }

func ptyStart(*exec.Cmd) (*os.File, error) {
	return nil, errors.New("pty not supported on this platform")
}

func ptySetsize(*os.File, int, int) error {
	return errors.New("pty not supported on this platform")
}
