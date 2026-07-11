package permission

import (
	"context"
	"testing"
	"time"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
)

func newTestManager(t *testing.T, opts ...Option) *Manager {
	t.Helper()
	db, err := storage.OpenWorkspaceDB(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return NewManager(NewStore(db), opts...)
}

func TestCheckDefaultsToAsk(t *testing.T) {
	m := newTestManager(t)
	d, err := m.Check(context.Background(), ports.PermissionQuery{
		Permission: core.PermWrite, Scope: core.ScopePath, Subject: "/repo/file.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.Outcome != core.OutcomeAsk {
		t.Errorf("default outcome = %s, want ask", d.Outcome)
	}
}

func TestDenyOverridesAllow(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t)
	_, _ = m.GrantPermission(ctx, Grant{Permission: core.PermNetwork, Scope: core.ScopeDomain, Selector: "*", Effect: EffectAllow})
	_, _ = m.GrantPermission(ctx, Grant{Permission: core.PermNetwork, Scope: core.ScopeDomain, Selector: "evil.example", Effect: EffectDeny})

	d, _ := m.Check(ctx, ports.PermissionQuery{Permission: core.PermNetwork, Scope: core.ScopeDomain, Subject: "evil.example"})
	if d.Outcome != core.OutcomeDeny {
		t.Errorf("deny must override allow, got %s", d.Outcome)
	}
	// A different domain gets the allow.
	d2, _ := m.Check(ctx, ports.PermissionQuery{Permission: core.PermNetwork, Scope: core.ScopeDomain, Subject: "good.example"})
	if d2.Outcome != core.OutcomeAllow {
		t.Errorf("expected allow for good.example, got %s", d2.Outcome)
	}
}

func TestUnknownPermissionDeniesWithE_SEC_002(t *testing.T) {
	m := newTestManager(t)
	d, err := m.Check(context.Background(), ports.PermissionQuery{Permission: "teleport", Subject: "x"})
	if d.Outcome != core.OutcomeDeny {
		t.Errorf("unknown permission must deny, got %s", d.Outcome)
	}
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != "E-SEC-002" {
		t.Errorf("want E-SEC-002, got %v", err)
	}
}

func TestExpiredGrantIgnored(t *testing.T) {
	ctx := context.Background()
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	m := newTestManager(t)
	_, _ = m.GrantPermission(ctx, Grant{
		Permission: core.PermRead, Scope: core.ScopePath, Selector: "*", Effect: EffectAllow, ValidUntil: past,
	})
	d, _ := m.Check(ctx, ports.PermissionQuery{Permission: core.PermRead, Scope: core.ScopePath, Subject: "/x"})
	if d.Outcome != core.OutcomeAsk {
		t.Errorf("expired grant should not apply; got %s", d.Outcome)
	}
}

func TestPathGlobSelector(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t)
	_, _ = m.GrantPermission(ctx, Grant{Permission: core.PermWrite, Scope: core.ScopePath, Selector: "/repo/**", Effect: EffectAllow})
	allow, _ := m.Check(ctx, ports.PermissionQuery{Permission: core.PermWrite, Scope: core.ScopePath, Subject: "/repo/pkg/x.go"})
	if allow.Outcome != core.OutcomeAllow {
		t.Errorf("glob should allow nested path, got %s", allow.Outcome)
	}
	outside, _ := m.Check(ctx, ports.PermissionQuery{Permission: core.PermWrite, Scope: core.ScopePath, Subject: "/etc/passwd"})
	if outside.Outcome != core.OutcomeAsk {
		t.Errorf("path outside glob should not be allowed, got %s", outside.Outcome)
	}
}

func TestNonInteractiveAskDenies(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t)
	d, _ := m.Request(ctx, ports.PermissionRequest{
		Query:       ports.PermissionQuery{Permission: core.PermExecute, Scope: core.ScopeCommand, Subject: "rm"},
		Interactive: false,
	})
	if d.Outcome != core.OutcomeDeny {
		t.Errorf("non-interactive ask must deny, got %s", d.Outcome)
	}
}

// grantingApprover approves with allow_for_session.
type grantingApprover struct{}

func (grantingApprover) Approve(_ context.Context, _ ports.PermissionRequest) (core.DecisionOutcome, core.PermissionDecisionKind, error) {
	return core.OutcomeAllow, core.DecisionAllowForSession, nil
}

func TestInteractiveApprovalMintsSessionGrant(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t, WithApprover(grantingApprover{}))
	sid := core.NewULID()
	q := ports.PermissionQuery{Permission: core.PermWrite, Scope: core.ScopePath, Subject: "/repo/a.go", SessionID: sid}
	d, err := m.Request(ctx, ports.PermissionRequest{Query: q, Interactive: true})
	if err != nil {
		t.Fatal(err)
	}
	if d.Outcome != core.OutcomeAllow || d.Kind != core.DecisionAllowForSession {
		t.Fatalf("approval = %+v", d)
	}
	// A subsequent Check in the same session is now allowed without prompting.
	d2, _ := m.Check(ctx, q)
	if d2.Outcome != core.OutcomeAllow {
		t.Errorf("session grant not applied on re-check, got %s", d2.Outcome)
	}
	// A different session does not inherit the grant.
	q3 := q
	q3.SessionID = core.NewULID()
	d3, _ := m.Check(ctx, q3)
	if d3.Outcome != core.OutcomeAsk {
		t.Errorf("grant leaked across sessions, got %s", d3.Outcome)
	}
}

func TestEveryDecisionIsAudited(t *testing.T) {
	ctx := context.Background()
	db, err := storage.OpenWorkspaceDB(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	m := NewManager(store)
	_, _ = m.Check(ctx, ports.PermissionQuery{Permission: core.PermRead, Subject: "/x"})
	_, _ = m.Check(ctx, ports.PermissionQuery{Permission: core.PermWrite, Subject: "/y"})
	n, _ := store.AuditCount(ctx)
	if n != 2 {
		t.Errorf("audit count = %d, want 2", n)
	}
}

func TestRevoke(t *testing.T) {
	ctx := context.Background()
	m := newTestManager(t)
	id, _ := m.GrantPermission(ctx, Grant{Permission: core.PermRead, Scope: core.ScopePath, Selector: "*", Effect: EffectAllow})
	if d, _ := m.Check(ctx, ports.PermissionQuery{Permission: core.PermRead, Subject: "/z"}); d.Outcome != core.OutcomeAllow {
		t.Fatal("expected allow before revoke")
	}
	if err := m.Revoke(ctx, id); err != nil {
		t.Fatal(err)
	}
	if d, _ := m.Check(ctx, ports.PermissionQuery{Permission: core.PermRead, Subject: "/z"}); d.Outcome != core.OutcomeAsk {
		t.Error("expected ask after revoke")
	}
}
