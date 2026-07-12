//go:build !darwin

package sandbox

import "github.com/datamaia/andromeda/internal/ports"

// osIsolationSupported reports whether OS-level isolation is available on this platform. Linux
// Landlock/bubblewrap wiring is a follow-up (ADR-021, PENDING VALIDATION); until then only the
// process-level layer is offered here, so the effective level is honestly reported as process.
func osIsolationSupported() bool { return false }

// osIsolationName identifies the platform mechanism.
const osIsolationName = "none"

// wrapOSIsolation is a passthrough on platforms without an OS-level layer.
func wrapOSIsolation(_ ports.SandboxPolicy, program string, args []string) (string, []string) {
	return program, args
}
