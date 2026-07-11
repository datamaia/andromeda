//go:build unix

package terminal

import (
	"os/exec"
	"syscall"

	"github.com/datamaia/andromeda/internal/ports"
)

// sendSignal maps a portable signal name to a Unix signal and delivers it to the process.
func sendSignal(cmd *exec.Cmd, sig ports.SignalName) error {
	if cmd.Process == nil {
		return termErr("E-TOOL-014", "process not running", nil)
	}
	var s syscall.Signal
	switch sig {
	case ports.SignalInterrupt:
		s = syscall.SIGINT
	case ports.SignalTerminate:
		s = syscall.SIGTERM
	case ports.SignalKill:
		s = syscall.SIGKILL
	default:
		s = syscall.SIGTERM
	}
	return cmd.Process.Signal(s)
}
