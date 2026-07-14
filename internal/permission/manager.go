package permission

import (
	"context"
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// Approver renders an interactive Approval decision for an "ask" outcome. Drivers (CLI/TUI)
// implement it. A nil approver means non-interactive: "ask" resolves to deny (PRD-009).
type Approver interface {
	Approve(ctx context.Context, req ports.PermissionRequest) (core.DecisionOutcome, core.PermissionDecisionKind, error)
}

// Clock returns the current time; injectable for tests.
type Clock func() time.Time

// Manager implements ports.PermissionPort.
type Manager struct {
	store     *Store
	policy    []Grant          // effect-bearing rules from resolved configuration
	allowlist CommandAllowlist // terminal-command allow/deny from [permission] config
	approver  Approver
	now       Clock
	actor     string
}

// Option configures a Manager.
type Option func(*Manager)

// WithApprover sets the interactive approver.
func WithApprover(a Approver) Option { return func(m *Manager) { m.approver = a } }

// WithClock sets the clock source.
func WithClock(c Clock) Option { return func(m *Manager) { m.now = c } }

// WithPolicy sets standing policy rules (from configuration).
func WithPolicy(rules []Grant) Option { return func(m *Manager) { m.policy = rules } }

// WithCommandAllowlist sets the terminal-command allowlist consulted during evaluation, so
// allow-listed commands resolve without prompting and deny-listed commands are refused.
func WithCommandAllowlist(a CommandAllowlist) Option { return func(m *Manager) { m.allowlist = a } }

// WithActor sets the actor label recorded in audit rows.
func WithActor(a string) Option { return func(m *Manager) { m.actor = a } }

// NewManager returns a Manager over the given store.
func NewManager(store *Store, opts ...Option) *Manager {
	m := &Manager{store: store, now: time.Now, actor: "system"}
	for _, o := range opts {
		o(m)
	}
	return m
}

var _ ports.PermissionPort = (*Manager)(nil)

// knownPermission reports whether p is one of the 13 closed permission values.
func knownPermission(p core.Permission) bool {
	switch p {
	case core.PermRead, core.PermWrite, core.PermExecute, core.PermNetwork,
		core.PermCredentialAccess, core.PermGitMutation, core.PermProcessSpawn,
		core.PermContainerAccess, core.PermExternalServiceAccess, core.PermClipboard,
		core.PermNotifications, core.PermPackageInstallation, core.PermSystemModification:
		return true
	}
	return false
}

// evaluate runs the collect/filter/resolve steps and returns the outcome and deciding IDs.
func (m *Manager) evaluate(ctx context.Context, q ports.PermissionQuery) (core.DecisionOutcome, []core.ULID, error) {
	if !knownPermission(q.Permission) {
		// Unknown permission: E-SEC-002, resolved as deny (fail-closed).
		return core.OutcomeDeny, nil, &ports.PortError{
			Code: "E-SEC-002", Category: "security", Severity: "error",
			Message: "unknown permission", Detail: string(q.Permission),
		}
	}
	now := m.now()
	stored, err := m.store.ActiveGrantsFor(ctx, q.Permission)
	if err != nil {
		return core.OutcomeDeny, nil, &ports.PortError{
			Code: "E-SEC-002", Category: "security", Severity: "error",
			Message: "permission evaluation failed", Detail: err.Error(), Cause: err,
		}
	}
	candidates := make([]Grant, 0, len(stored)+len(m.policy))
	for _, g := range stored {
		if g.matches(q, now) {
			candidates = append(candidates, g)
		}
	}
	for _, r := range m.policy {
		if r.matches(q, now) {
			candidates = append(candidates, r)
		}
	}
	// Command allowlist ([permission] in andromeda.toml): pre-authorize or refuse vetted terminal
	// commands by argv prefix. Appended as a synthetic candidate so the normal precedence
	// (deny > ask > allow) applies — an explicit stored deny still overrides an allow-listed command.
	if q.Scope == core.ScopeCommand && q.Command != "" && !m.allowlist.Empty() {
		if eff, matched := m.allowlist.Decide(q.Command); matched {
			candidates = append(candidates, Grant{ID: allowlistGrantID, Effect: eff})
		}
	}
	outcome, deciding := resolve(candidates)
	return outcome, deciding, nil
}

// Check performs non-interactive evaluation (never prompts).
func (m *Manager) Check(ctx context.Context, q ports.PermissionQuery) (ports.Decision, error) {
	outcome, deciding, evalErr := m.evaluate(ctx, q)
	if err := m.audit(ctx, q, outcome, deciding); err != nil {
		// Fail-closed: an audit write failure denies the action (E-SEC-014).
		return ports.Decision{Outcome: core.OutcomeDeny}, err
	}
	return ports.Decision{Outcome: outcome}, evalErr
}

// Request performs the full decision flow, raising an Approval on "ask" when interaction is
// permitted and an approver is present. Non-interactive "ask" resolves to deny.
func (m *Manager) Request(ctx context.Context, req ports.PermissionRequest) (ports.Decision, error) {
	outcome, deciding, evalErr := m.evaluate(ctx, req.Query)
	if evalErr != nil {
		_ = m.audit(ctx, req.Query, core.OutcomeDeny, deciding)
		return ports.Decision{Outcome: core.OutcomeDeny}, evalErr
	}
	if outcome != core.OutcomeAsk {
		if err := m.audit(ctx, req.Query, outcome, deciding); err != nil {
			return ports.Decision{Outcome: core.OutcomeDeny}, err
		}
		return ports.Decision{Outcome: outcome}, nil
	}

	// outcome == ask
	if !req.Interactive || m.approver == nil {
		_ = m.audit(ctx, req.Query, core.OutcomeDeny, deciding)
		return ports.Decision{Outcome: core.OutcomeDeny}, nil
	}
	decided, kind, err := m.approver.Approve(ctx, req)
	if err != nil {
		_ = m.audit(ctx, req.Query, core.OutcomeDeny, nil)
		return ports.Decision{Outcome: core.OutcomeDeny}, err
	}
	approvalID := core.NewULID()
	if g := m.mintGrantFor(kind, req.Query, decided); g != nil {
		if err := m.store.SaveGrant(ctx, *g); err != nil {
			return ports.Decision{Outcome: core.OutcomeDeny}, err
		}
	}
	if err := m.audit(ctx, req.Query, decided, []core.ULID{approvalID}); err != nil {
		return ports.Decision{Outcome: core.OutcomeDeny}, err
	}
	return ports.Decision{Outcome: decided, Kind: kind, ApprovalID: approvalID}, nil
}

// RecordDecision persists a decision produced elsewhere (e.g. policy pre-resolution).
func (m *Manager) RecordDecision(ctx context.Context, rec ports.DecisionRecord) error {
	actor := rec.Actor
	if actor == "" {
		actor = m.actor
	}
	return m.store.AppendAudit(ctx, AuditRecord{
		Permission: rec.Query.Permission,
		Scope:      rec.Query.Scope,
		Subject:    rec.Query.Subject,
		Outcome:    rec.Decision.Outcome,
		Deciding:   string(rec.Decision.Kind),
		Actor:      actor,
		Timestamp:  m.now().UTC().Format(time.RFC3339Nano),
	})
}

// GrantPermission mints and persists a grant directly (used by policy setup and the CLI).
func (m *Manager) GrantPermission(ctx context.Context, g Grant) (core.ULID, error) {
	if g.ID == "" {
		g.ID = core.NewULID()
	}
	if g.CreatedAt == "" {
		g.CreatedAt = m.now().UTC().Format(time.RFC3339Nano)
	}
	if g.Effect == "" {
		g.Effect = EffectAllow
	}
	return g.ID, m.store.SaveGrant(ctx, g)
}

// Revoke revokes a grant by ID.
func (m *Manager) Revoke(ctx context.Context, id core.ULID) error {
	return m.store.RevokeGrant(ctx, id)
}

// ListGrants returns all grants.
func (m *Manager) ListGrants(ctx context.Context) ([]Grant, error) { return m.store.ListGrants(ctx) }

func (m *Manager) audit(ctx context.Context, q ports.PermissionQuery, outcome core.DecisionOutcome, deciding []core.ULID) error {
	return m.store.AppendAudit(ctx, AuditRecord{
		Permission: q.Permission,
		Scope:      q.Scope,
		Subject:    q.Subject,
		Outcome:    outcome,
		Deciding:   strings.Join(deciding, ","),
		Actor:      m.actor,
		Timestamp:  m.now().UTC().Format(time.RFC3339Nano),
	})
}

func (m *Manager) mintGrantFor(kind core.PermissionDecisionKind, q ports.PermissionQuery, _ core.DecisionOutcome) *Grant {
	base := Grant{
		ID:         core.NewULID(),
		Permission: q.Permission,
		Scope:      q.Scope,
		Selector:   selectorFor(q.Subject),
		CreatedAt:  m.now().UTC().Format(time.RFC3339Nano),
	}
	switch kind {
	case core.DecisionAllowForSession:
		base.Effect, base.Scope, base.SubjectSession = EffectAllow, core.ScopeSession, q.SessionID
		return &base
	case core.DecisionAllowForWorkspace, core.DecisionAlwaysAllowPolicy:
		base.Effect, base.Scope, base.SubjectWorkspace = EffectAllow, core.ScopeWorkspace, q.WorkspaceID
		return &base
	case core.DecisionAlwaysDeny:
		base.Effect = EffectDeny
		return &base
	default:
		// allow_once, deny_once, ask_every_time — no persisted grant.
		return nil
	}
}

// selectorFor turns a concrete subject into an exact selector; an empty subject becomes "*".
func selectorFor(subject string) string {
	if subject == "" {
		return "*"
	}
	return subject
}
