//go:build unix

package main

import (
	"os/exec"
	"syscall"
)

// detach puts the child in its own process group so the TUI's terminal signals do not reach it and
// it keeps running in the background independently.
func detach(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}
