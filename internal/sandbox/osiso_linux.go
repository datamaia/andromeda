//go:build linux

package sandbox

import (
	"os/exec"

	"github.com/datamaia/andromeda/internal/ports"
)

// osIsolationSupported reports whether Linux OS-level isolation is available. bubblewrap (bwrap)
// is the ADR-021 mechanism: an unprivileged sandboxing wrapper that mirrors macOS sandbox-exec.
// Landlock (kernel ≥ 5.13) is a complementary in-process mechanism tracked as a follow-up.
func osIsolationSupported() bool {
	_, err := exec.LookPath("bwrap")
	return err == nil
}

// wrapOSIsolation wraps a command in bubblewrap with a filesystem and network policy derived
// from the SandboxPolicy: the root is bound read-only, the policy's write paths are bound
// read-write, /proc and a fresh /dev and /tmp are provided, and the network namespace is
// unshared when the policy denies network. Returns the wrapped program and args.
func wrapOSIsolation(policy ports.SandboxPolicy, program string, args []string) (string, []string) {
	bw := []string{
		"--ro-bind", "/", "/",
		"--dev", "/dev",
		"--proc", "/proc",
		"--tmpfs", "/tmp",
		"--die-with-parent",
	}
	for _, w := range policy.WritePaths {
		bw = append(bw, "--bind", w, w)
	}
	if policy.NetworkPolicy == "deny" || policy.NetworkPolicy == "" {
		bw = append(bw, "--unshare-net")
	}
	bw = append(bw, "--", program)
	bw = append(bw, args...)
	return "bwrap", bw
}
