package permission

import (
	"strings"

	"github.com/datamaia/andromeda/internal/core"
)

// allowlistGrantID labels the synthetic decision minted by a command-allowlist match so audit
// rows attribute the outcome to configuration rather than a stored grant.
const allowlistGrantID core.ULID = "policy-command-allowlist"

// CommandAllowlist pre-authorizes (or refuses) terminal_run commands by argv prefix, letting the
// agent run vetted commands without an approval prompt. Entries are whitespace-delimited argv
// prefixes exactly as written in andromeda.toml's [permission] section: "git status" matches
// "git status --short" but not "github" nor a bare "git". Deny is checked before allow, so a
// denied command is refused even if a broader allow entry would match.
type CommandAllowlist struct {
	allow [][]string
	deny  [][]string
}

// NewCommandAllowlist builds an allowlist from raw allow/deny entries. Blank entries are ignored.
func NewCommandAllowlist(allow, deny []string) CommandAllowlist {
	return CommandAllowlist{allow: tokenize(allow), deny: tokenize(deny)}
}

// Empty reports whether the allowlist carries no rules, so callers can skip wiring it.
func (a CommandAllowlist) Empty() bool { return len(a.allow) == 0 && len(a.deny) == 0 }

// Decide returns the effect for a command and whether any rule matched. Deny wins over allow; an
// unmatched command returns ("", false) so the normal ask/deny flow proceeds.
func (a CommandAllowlist) Decide(command string) (Effect, bool) {
	argv := strings.Fields(command)
	if len(argv) == 0 {
		return "", false
	}
	for _, d := range a.deny {
		if hasArgvPrefix(argv, d) {
			return EffectDeny, true
		}
	}
	for _, al := range a.allow {
		if hasArgvPrefix(argv, al) {
			return EffectAllow, true
		}
	}
	return "", false
}

// tokenize splits each entry into argv tokens on whitespace, dropping blank entries.
func tokenize(entries []string) [][]string {
	out := make([][]string, 0, len(entries))
	for _, e := range entries {
		if toks := strings.Fields(e); len(toks) > 0 {
			out = append(out, toks)
		}
	}
	return out
}

// hasArgvPrefix reports whether prefix is a leading, token-wise subsequence of argv.
func hasArgvPrefix(argv, prefix []string) bool {
	if len(prefix) > len(argv) {
		return false
	}
	for i := range prefix {
		if argv[i] != prefix[i] {
			return false
		}
	}
	return true
}
