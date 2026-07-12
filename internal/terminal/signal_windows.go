//go:build windows

package terminal

import (
	"os/exec"
	"strconv"

	"github.com/datamaia/andromeda/internal/ports"
)

// setpgid is a no-op on Windows: process-tree termination is handled by taskkill /T in
// sendSignal, so no process-group setup is needed at start time.
func setpgid(_ *exec.Cmd) {}

// sendSignal delivers a signal to the process. Windows lacks POSIX signals; interrupt and
// terminate map to a forced process-tree termination via taskkill, as does kill.
func sendSignal(cmd *exec.Cmd, _ ports.SignalName) error {
	if cmd.Process == nil {
		return termErr("E-TOOL-014", "process not running", nil)
	}
	return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
}
