package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// SessionStore implements ports.SessionStorePort over a workspace database. Storage mechanics
// are Volume 10's; run/turn semantics are Volume 4's. This is the durable substrate behind
// PRD-010 (recoverable work).
type SessionStore struct {
	db *DB
}

// NewSessionStore returns a SessionStore over the given workspace database.
func NewSessionStore(db *DB) *SessionStore { return &SessionStore{db: db} }

var _ ports.SessionStorePort = (*SessionStore)(nil)

// ErrRevisionConflict indicates optimistic-concurrency failure on SaveSession.
var ErrRevisionConflict = errors.New("storage: session revision conflict")

// ErrNotFound indicates a requested row does not exist.
var ErrNotFound = errors.New("storage: not found")

func (s *SessionStore) SaveSession(ctx context.Context, snap ports.SessionSnapshot) error {
	if snap.ID == "" {
		return fmt.Errorf("SaveSession: empty session id")
	}
	if snap.CreatedAt == "" {
		snap.CreatedAt = nowUTC()
	}
	// Insert new, or update with optimistic concurrency on revision.
	res, err := s.db.sql.ExecContext(ctx, `
INSERT INTO sessions (id, state, revision, created_at, data)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  state = excluded.state,
  revision = sessions.revision + 1,
  data = excluded.data
WHERE sessions.revision = ?`,
		snap.ID, snap.State, snap.Revision, snap.CreatedAt, snap.Data, snap.Revision)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrRevisionConflict
	}
	return nil
}

func (s *SessionStore) LoadSession(ctx context.Context, id core.ULID) (ports.SessionSnapshot, error) {
	var snap ports.SessionSnapshot
	err := s.db.sql.QueryRowContext(ctx,
		`SELECT id, state, revision, created_at, data FROM sessions WHERE id = ?`, id).
		Scan(&snap.ID, &snap.State, &snap.Revision, &snap.CreatedAt, &snap.Data)
	if errors.Is(err, sql.ErrNoRows) {
		return snap, ErrNotFound
	}
	return snap, err
}

func (s *SessionStore) ListSessions(ctx context.Context, f ports.SessionFilter) ([]ports.SessionSummary, error) {
	q := `SELECT id, state, created_at FROM sessions`
	var args []any
	if f.State != "" {
		q += ` WHERE state = ?`
		args = append(args, f.State)
	}
	q += ` ORDER BY id DESC`
	if f.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, f.Limit)
	}
	rows, err := s.db.sql.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.SessionSummary
	for rows.Next() {
		var su ports.SessionSummary
		if err := rows.Scan(&su.ID, &su.State, &su.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, su)
	}
	return out, rows.Err()
}

func (s *SessionStore) AppendRunRecords(ctx context.Context, runID core.ULID, batch []ports.RunRecord) error {
	if len(batch) == 0 {
		return nil
	}
	tx, err := s.db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Continue the per-run sequence from the current maximum.
	var maxSeq sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(seq) FROM run_records WHERE run_id = ?`, runID).Scan(&maxSeq); err != nil {
		_ = tx.Rollback()
		return err
	}
	seq := maxSeq.Int64
	for _, rec := range batch {
		seq++
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO run_records (id, run_id, seq, kind, payload, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			core.NewULID(), runID, seq, rec.Kind, rec.Payload, nowUTC()); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *SessionStore) LoadRun(ctx context.Context, id core.ULID) (ports.RunSnapshot, error) {
	var snap ports.RunSnapshot
	err := s.db.sql.QueryRowContext(ctx, `SELECT id, state FROM runs WHERE id = ?`, id).
		Scan(&snap.ID, &snap.State)
	if errors.Is(err, sql.ErrNoRows) {
		return snap, ErrNotFound
	}
	if err != nil {
		return snap, err
	}
	rows, err := s.db.sql.QueryContext(ctx,
		`SELECT kind, payload FROM run_records WHERE run_id = ? ORDER BY seq`, id)
	if err != nil {
		return snap, err
	}
	defer rows.Close()
	for rows.Next() {
		var rec ports.RunRecord
		if err := rows.Scan(&rec.Kind, &rec.Payload); err != nil {
			return snap, err
		}
		snap.Records = append(snap.Records, rec)
	}
	return snap, rows.Err()
}

func (s *SessionStore) ListRuns(ctx context.Context, f ports.RunFilter) ([]ports.RunSummary, error) {
	q := `SELECT id, session_id, state FROM runs`
	var conds []string
	var args []any
	if f.SessionID != "" {
		conds = append(conds, "session_id = ?")
		args = append(args, f.SessionID)
	}
	if f.State != "" {
		conds = append(conds, "state = ?")
		args = append(args, f.State)
	}
	for i, c := range conds {
		if i == 0 {
			q += " WHERE "
		} else {
			q += " AND "
		}
		q += c
	}
	q += " ORDER BY id DESC"
	if f.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, f.Limit)
	}
	rows, err := s.db.sql.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.RunSummary
	for rows.Next() {
		var ru ports.RunSummary
		if err := rows.Scan(&ru.ID, &ru.SessionID, &ru.State); err != nil {
			return nil, err
		}
		out = append(out, ru)
	}
	return out, rows.Err()
}

// MarkInterrupted atomically marks every non-terminal run in scope as interrupted and returns
// their IDs (crash recovery, PRD-010). Terminal states are never rewritten.
func (s *SessionStore) MarkInterrupted(ctx context.Context, scope ports.InterruptScope) ([]core.ULID, error) {
	const terminal = `('completed','failed','cancelled')`
	tx, err := s.db.sql.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	q := `SELECT id FROM runs WHERE state NOT IN ` + terminal
	var args []any
	if scope.SessionID != "" {
		q += ` AND session_id = ?`
		args = append(args, scope.SessionID)
	}
	rows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	var ids []core.ULID
	for rows.Next() {
		var id core.ULID
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			_ = tx.Rollback()
			return nil, err
		}
		ids = append(ids, id)
	}
	_ = rows.Close()
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, `UPDATE runs SET state = 'interrupted' WHERE id = ?`, id); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

// CreateRun inserts a run row; a convenience used by higher layers and tests until the Run
// aggregate's writer (Volume 4) lands.
func (s *SessionStore) CreateRun(ctx context.Context, runID, sessionID core.ULID, state string) error {
	_, err := s.db.sql.ExecContext(ctx,
		`INSERT INTO runs (id, session_id, state, created_at) VALUES (?, ?, ?, ?)`,
		runID, sessionID, state, nowUTC())
	return err
}
