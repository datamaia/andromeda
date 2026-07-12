//go:build darwin

package sandbox

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/datamaia/andromeda/internal/ports"
)

// osIsolationSupported reports whether macOS Seatbelt (sandbox-exec) is available.
func osIsolationSupported() bool {
	_, err := exec.LookPath("sandbox-exec")
	return err == nil
}

// wrapOSIsolation wraps a command in sandbox-exec with a Seatbelt profile generated from the
// policy: reads are broadly allowed, writes are denied except under the policy's write paths,
// and network is denied when the policy denies it. Returns the wrapped program and args.
func wrapOSIsolation(policy ports.SandboxPolicy, program string, args []string) (string, []string) {
	profile := seatbeltProfile(policy)
	wrapped := append([]string{"-p", profile, program}, args...)
	return "sandbox-exec", wrapped
}

func seatbeltProfile(policy ports.SandboxPolicy) string {
	var b strings.Builder
	b.WriteString("(version 1)\n(allow default)\n")
	// Deny all writes, then re-allow the policy's write subpaths (resolved to their real paths
	// so they match what Seatbelt sees after macOS symlink resolution, e.g. /var → /private/var).
	b.WriteString(`(deny file-write* (subpath "/"))` + "\n")
	for _, p := range policy.WritePaths {
		b.WriteString(`(allow file-write* (subpath "` + escapeSeatbelt(realPath(p)) + `"))` + "\n")
	}
	// Processes commonly need to write to the null device; nothing else is broadly allowed.
	b.WriteString(`(allow file-write-data (literal "/dev/null"))` + "\n")
	if policy.NetworkPolicy == "deny" || policy.NetworkPolicy == "" {
		b.WriteString("(deny network*)\n")
	}
	return b.String()
}

func realPath(p string) string {
	if rp, err := filepath.EvalSymlinks(p); err == nil {
		return rp
	}
	return p
}

func escapeSeatbelt(p string) string {
	return strings.ReplaceAll(p, `"`, `\"`)
}
