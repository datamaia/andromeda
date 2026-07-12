//go:build unix

package terminal

import (
	"os/exec"
	"syscall"

	"github.com/datamaia/andromeda/internal/ports"
)

// setpgid puts the child in its own process group so a later Signal can reach the whole process
// tree (the shell plus any command it forks or execs), not just the direct child. Without it, a
// grandchild that outlives its parent keeps the output pipes open, so the pump goroutines never
// see EOF and Wait blocks until the grandchild exits on its own. Call before cmd.Start.
func setpgid(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// sendSignal maps a portable signal name to a Unix signal and delivers it to the process group
// led by the command (see setpgid; PTY executions lead a session via pty.Start's Setsid), so a
// shell and everything it spawned all receive it. Falls back to signaling just the process if
// the group send fails (e.g. the child never became a group leader).
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
	// A negative PID targets the process group whose leader is this PID.
	if err := syscall.Kill(-cmd.Process.Pid, s); err == nil {
		return nil
	}
	return cmd.Process.Signal(s)
}
