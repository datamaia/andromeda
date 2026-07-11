package memory

import (
	"context"
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/datamaia/andromeda/internal/streams"
)

// Store implements ports.MemoryStorePort over a workspace database.
type Store struct {
	db  *storage.DB
	now func() time.Time
}

// New returns a memory Store.
func New(db *storage.DB) *Store { return &Store{db: db, now: time.Now} }

var _ ports.MemoryStorePort = (*Store)(nil)

// Ingest writes memory records and returns their minted ULIDs. A batch is transactional.
func (s *Store) Ingest(ctx context.Context, drafts []ports.MemoryRecordDraft) ([]core.ULID, error) {
	if len(drafts) == 0 {
		return nil, nil
	}
	tx, err := s.db.SQL().BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	ids := make([]core.ULID, 0, len(drafts))
	ts := s.now().UTC().Format(time.RFC3339Nano)
	for _, d := range drafts {
		id := core.NewULID()
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO memory_records (id, layer, content, provenance, source, status, created_at)
			 VALUES (?, ?, ?, ?, ?, 'active', ?)`,
			id, d.Layer, d.Content, nz(d.Provenance), nz(d.Source), ts); err != nil {
			_ = tx.Rollback()
			return nil, memErr("E-MEM-001", "failed to write memory record", err)
		}
		ids = append(ids, id)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

// Retrieve queries by layer, text (case-insensitive substring), and limit. Only active records
// are returned.
func (s *Store) Retrieve(ctx context.Context, q ports.MemoryQuery) ([]ports.MemoryRecord, error) {
	sql := `SELECT id, layer, content, provenance, source, status, created_at
	        FROM memory_records WHERE status = 'active'`
	var args []any
	if len(q.Layers) > 0 {
		sql += " AND layer IN (" + placeholders(len(q.Layers)) + ")"
		for _, l := range q.Layers {
			args = append(args, l)
		}
	}
	if q.Text != "" {
		sql += " AND lower(content) LIKE ?"
		args = append(args, "%"+strings.ToLower(q.Text)+"%")
	}
	sql += " ORDER BY created_at DESC"
	if q.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, q.Limit)
	}
	rows, err := s.db.SQL().QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, memErr("E-MEM-002", "memory query failed", err)
	}
	defer rows.Close()
	return scanRecords(rows)
}

// Rank scores an explicit candidate set against a query by term overlap, highest first.
func (s *Store) Rank(ctx context.Context, q ports.MemoryQuery, candidates []core.ULID) ([]ports.RankedMemory, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	rows, err := s.db.SQL().QueryContext(ctx,
		`SELECT id, content FROM memory_records WHERE id IN (`+placeholders(len(candidates))+`)`,
		toArgs(candidates)...)
	if err != nil {
		return nil, memErr("E-MEM-002", "memory rank failed", err)
	}
	defer rows.Close()
	terms := tokenize(q.Text)
	var out []ports.RankedMemory
	for rows.Next() {
		var id, content string
		if err := rows.Scan(&id, &content); err != nil {
			return nil, err
		}
		out = append(out, ports.RankedMemory{ID: id, Score: overlapScore(terms, tokenize(content))})
	}
	sortByScore(out)
	return out, rows.Err()
}

// Expire marks records in a layer older than the policy's cutoff as expired.
func (s *Store) Expire(ctx context.Context, policy ports.ExpirePolicy) (ports.ExpireReport, error) {
	sql := `UPDATE memory_records SET status = 'expired' WHERE status = 'active'`
	var args []any
	if policy.Layer != "" {
		sql += " AND layer = ?"
		args = append(args, policy.Layer)
	}
	if policy.OlderThan != "" {
		sql += " AND created_at < ?"
		args = append(args, policy.OlderThan)
	}
	res, err := s.db.SQL().ExecContext(ctx, sql, args...)
	if err != nil {
		return ports.ExpireReport{}, memErr("E-MEM-003", "expire failed", err)
	}
	n, _ := res.RowsAffected()
	return ports.ExpireReport{Expired: int(n)}, nil
}

// Delete hard-deletes records by ID.
func (s *Store) Delete(ctx context.Context, ids []core.ULID) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.db.SQL().ExecContext(ctx,
		`DELETE FROM memory_records WHERE id IN (`+placeholders(len(ids))+`)`, toArgs(ids)...)
	if err != nil {
		return memErr("E-MEM-004", "delete failed", err)
	}
	return nil
}

// Export streams matching records as a slice-backed stream.
func (s *Store) Export(ctx context.Context, q ports.MemoryQuery) (ports.Stream[ports.MemoryRecord], error) {
	recs, err := s.Retrieve(ctx, q)
	if err != nil {
		return nil, err
	}
	return streams.Slice(recs), nil
}

func memErr(code, msg string, cause error) error {
	return &ports.PortError{Code: code, Category: "memory", Severity: "error", Message: msg, Detail: cause.Error(), Cause: cause}
}
