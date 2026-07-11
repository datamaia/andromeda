package sandbox

import (
	"path/filepath"
	"strings"

	"github.com/datamaia/andromeda/internal/logging"
	"github.com/datamaia/andromeda/internal/ports"
)

// filterEnv applies deny-by-default environment passthrough (ADR-021): only variables named in
// allow are kept, drawn from environ ("KEY=VALUE" entries). Sensitive-looking names are dropped
// even if allow-listed unless the name is listed verbatim, so a broad allowlist cannot leak a
// token by accident.
func filterEnv(environ []string, allow []string) []string {
	allowed := make(map[string]bool, len(allow))
	for _, a := range allow {
		allowed[a] = true
	}
	var out []string
	for _, kv := range environ {
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			continue
		}
		name := kv[:eq]
		if !allowed[name] {
			continue
		}
		if logging.IsSensitiveKey(name) && !containsExact(allow, name) {
			continue
		}
		out = append(out, kv)
	}
	return out
}

func containsExact(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// commandAllowed reports whether program passes the policy's allow/deny lists. Deny wins; an
// empty allowlist permits anything not denied; a non-empty allowlist permits only its members.
// Matching is on the command's base name.
func commandAllowed(policy ports.SandboxPolicy, program string) bool {
	base := filepath.Base(program)
	for _, d := range policy.CommandDeny {
		if d == base || d == program {
			return false
		}
	}
	if len(policy.CommandAllow) == 0 {
		return true
	}
	for _, a := range policy.CommandAllow {
		if a == base || a == program {
			return true
		}
	}
	return false
}

// pathWithin reports whether target is within one of roots (after cleaning). An empty roots
// list denies everything (deny-by-default), except that the sandbox working directory is always
// permitted by the caller.
func pathWithin(target string, roots []string) bool {
	tc := filepath.Clean(target)
	for _, r := range roots {
		rc := filepath.Clean(r)
		if tc == rc || strings.HasPrefix(tc, rc+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
