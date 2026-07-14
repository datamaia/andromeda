package permission

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

func TestCommandAllowlistDecide(t *testing.T) {
	al := NewCommandAllowlist(
		[]string{"git status", "git diff", "go test ./...", "ls"},
		[]string{"git push --force", "rm -rf"},
	)
	cases := []struct {
		command string
		want    Effect
		matched bool
	}{
		{"git status --short", EffectAllow, true},          // argv prefix beyond the entry
		{"git status", EffectAllow, true},                  // exact
		{"go test ./...", EffectAllow, true},               // multi-token exact
		{"ls -la", EffectAllow, true},                      // single-token entry matches any args
		{"git push --force origin main", EffectDeny, true}, // deny matches prefix
		{"rm -rf /tmp/x", EffectDeny, true},                // deny matches prefix
		{"git", "", false},                                 // bare binary: allow entry is longer, no match
		{"github status", "", false},                       // token boundary: not the same as "git"
		{"go build ./...", "", false},                      // not listed
		{"", "", false},                                    // empty command
	}
	for _, c := range cases {
		gotEff, gotMatched := al.Decide(c.command)
		if gotEff != c.want || gotMatched != c.matched {
			t.Errorf("Decide(%q) = (%q,%v), want (%q,%v)", c.command, gotEff, gotMatched, c.want, c.matched)
		}
	}
}

// Deny beats allow even when an allow entry would also match the command.
func TestCommandAllowlistDenyBeatsAllow(t *testing.T) {
	al := NewCommandAllowlist([]string{"git"}, []string{"git push --force"})
	if eff, _ := al.Decide("git push --force origin"); eff != EffectDeny {
		t.Errorf("deny should win over a broader allow, got %q", eff)
	}
	// The same broad allow still applies to a non-denied git command.
	if eff, _ := al.Decide("git status"); eff != EffectAllow {
		t.Errorf("allow should apply to a non-denied command, got %q", eff)
	}
}

func TestCommandAllowlistEmpty(t *testing.T) {
	if !NewCommandAllowlist(nil, nil).Empty() {
		t.Error("nil allow/deny should be Empty")
	}
	if !NewCommandAllowlist([]string{"  ", ""}, nil).Empty() {
		t.Error("blank-only entries should be Empty")
	}
	if NewCommandAllowlist([]string{"ls"}, nil).Empty() {
		t.Error("a real entry should not be Empty")
	}
}

// An allow-listed command resolves to allow through the Manager without any stored grant or
// approver — the config allowlist is the deciding policy. A non-listed command still asks.
func TestManagerAllowlistAllows(t *testing.T) {
	ctx := context.Background()
	al := NewCommandAllowlist([]string{"git status", "go test"}, nil)
	m := newTestManager(t, WithCommandAllowlist(al))

	query := func(cmd string) ports.PermissionQuery {
		return ports.PermissionQuery{Permission: core.PermExecute, Scope: core.ScopeCommand, Subject: "git", Command: cmd}
	}
	if d, _ := m.Check(ctx, query("git status --short")); d.Outcome != core.OutcomeAllow {
		t.Errorf("allow-listed command = %s, want allow", d.Outcome)
	}
	if d, _ := m.Check(ctx, query("git push")); d.Outcome != core.OutcomeAsk {
		t.Errorf("non-listed command = %s, want ask", d.Outcome)
	}
}

// A deny-listed command resolves to deny through the Manager, even under an interactive request
// (it is refused without ever reaching the approver).
func TestManagerAllowlistDenies(t *testing.T) {
	ctx := context.Background()
	al := NewCommandAllowlist([]string{"git"}, []string{"git push --force"})
	m := newTestManager(t, WithCommandAllowlist(al), WithApprover(grantingApprover{}))

	d, _ := m.Request(ctx, ports.PermissionRequest{
		Query:       ports.PermissionQuery{Permission: core.PermExecute, Scope: core.ScopeCommand, Subject: "git", Command: "git push --force origin"},
		Interactive: true,
	})
	if d.Outcome != core.OutcomeDeny {
		t.Errorf("deny-listed command = %s, want deny", d.Outcome)
	}
}

// A stored deny grant still overrides an allow-listed command (precedence: deny wins).
func TestManagerStoredDenyOverridesAllowlist(t *testing.T) {
	ctx := context.Background()
	al := NewCommandAllowlist([]string{"git status"}, nil)
	m := newTestManager(t, WithCommandAllowlist(al))
	_, _ = m.GrantPermission(ctx, Grant{
		Permission: core.PermExecute, Scope: core.ScopeCommand, Selector: "git", Effect: EffectDeny,
	})
	d, _ := m.Check(ctx, ports.PermissionQuery{
		Permission: core.PermExecute, Scope: core.ScopeCommand, Subject: "git", Command: "git status",
	})
	if d.Outcome != core.OutcomeDeny {
		t.Errorf("stored deny should override the allowlist, got %s", d.Outcome)
	}
}
