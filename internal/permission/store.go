package permission

import (
	"context"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/storage"
)

// Store persists grants and permission audit records in a workspace database.
type Store struct {
	db *storage.DB
}

// NewStore returns a Store over the given database (schema v3+).
func NewStore(db *storage.DB) *Store { return &Store{db: db} }

// SaveGrant inserts a grant row.
func (s *Store) SaveGrant(ctx context.Context, g Grant) error {
	_, err := s.db.SQL().ExecContext(ctx, `
INSERT INTO grants (id, permission, scope, selector, effect, subject_session, subject_workspace, valid_until, revoked, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		g.ID, g.Permission, g.Scope, g.Selector, g.Effect,
		nz(g.SubjectSession), nz(g.SubjectWorkspace), nz(g.ValidUntil), boolToInt(g.Revoked), g.CreatedAt)
	return err
}

// ActiveGrantsFor returns unrevoked grants for a permission. Expiry is evaluated in memory
// against the caller's clock so the model stays testable and network-free.
func (s *Store) ActiveGrantsFor(ctx context.Context, perm core.Permission) ([]Grant, error) {
	rows, err := s.db.SQL().QueryContext(ctx, `
SELECT id, permission, scope, selector, effect, subject_session, subject_workspace, valid_until, revoked, created_at
FROM grants WHERE permission = ? AND revoked = 0`, perm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Grant
	for rows.Next() {
		var g Grant
		var sess, ws, valid *string
		var revoked int
		if err := rows.Scan(&g.ID, &g.Permission, &g.Scope, &g.Selector, &g.Effect,
			&sess, &ws, &valid, &revoked, &g.CreatedAt); err != nil {
			return nil, err
		}
		g.SubjectSession = deref(sess)
		g.SubjectWorkspace = deref(ws)
		g.ValidUntil = deref(valid)
		g.Revoked = revoked != 0
		out = append(out, g)
	}
	return out, rows.Err()
}

// RevokeGrant marks a grant revoked.
func (s *Store) RevokeGrant(ctx context.Context, id core.ULID) error {
	_, err := s.db.SQL().ExecContext(ctx, `UPDATE grants SET revoked = 1 WHERE id = ?`, id)
	return err
}

// ListGrants returns all grants (revoked included) for inspection.
func (s *Store) ListGrants(ctx context.Context) ([]Grant, error) {
	rows, err := s.db.SQL().QueryContext(ctx, `
SELECT id, permission, scope, selector, effect, subject_session, subject_workspace, valid_until, revoked, created_at
FROM grants ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Grant
	for rows.Next() {
		var g Grant
		var sess, ws, valid *string
		var revoked int
		if err := rows.Scan(&g.ID, &g.Permission, &g.Scope, &g.Selector, &g.Effect,
			&sess, &ws, &valid, &revoked, &g.CreatedAt); err != nil {
			return nil, err
		}
		g.SubjectSession, g.SubjectWorkspace, g.ValidUntil = deref(sess), deref(ws), deref(valid)
		g.Revoked = revoked != 0
		out = append(out, g)
	}
	return out, rows.Err()
}

// AuditRecord is one persisted decision record.
type AuditRecord struct {
	Permission core.Permission
	Scope      core.PermissionScope
	Subject    string
	Outcome    core.DecisionOutcome
	Deciding   string
	Actor      string
	Timestamp  string
}

// AppendAudit persists one decision record. A failed audit write blocks the action (E-SEC-014,
// chapter 08): the caller treats a returned error as fail-closed.
func (s *Store) AppendAudit(ctx context.Context, r AuditRecord) error {
	_, err := s.db.SQL().ExecContext(ctx, `
INSERT INTO permission_audit (id, permission, scope, subject, outcome, deciding, actor, ts)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		core.NewULID(), r.Permission, r.Scope, r.Subject, r.Outcome, r.Deciding, r.Actor, r.Timestamp)
	return err
}

// AuditCount returns the number of audit rows (diagnostic helper).
func (s *Store) AuditCount(ctx context.Context) (int, error) {
	var n int
	err := s.db.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM permission_audit`).Scan(&n)
	return n, err
}

func nz(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
