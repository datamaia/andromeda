//go:build windows

package main

import "os/exec"

// detach is a no-op on Windows, which has no process groups in the Unix sense; the child still runs
// concurrently.
func detach(cmd *exec.Cmd) {}
