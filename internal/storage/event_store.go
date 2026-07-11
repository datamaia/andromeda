package storage

import (
	"context"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// EventStore persists enveloped Event records (Volume 2) to the workspace database, providing
// the durable trail behind the at-most-once in-process event bus (ADR-012). An event ID is
// minted on append when the envelope does not carry one in its correlation slot.
type EventStore struct {
	db *DB
}

// NewEventStore returns an EventStore over the given database.
func NewEventStore(db *DB) *EventStore { return &EventStore{db: db} }

// Append writes one event record and returns its stored ID.
func (s *EventStore) Append(ctx context.Context, e ports.Event) (core.ULID, error) {
	id := core.NewULID()
	_, err := s.db.sql.ExecContext(ctx, `
INSERT INTO events (id, name, version, producer, correlation_id, session_id, run_id, ts, payload)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, e.Name, e.Version, e.Producer,
		nullable(e.CorrelationID), nullable(e.SessionID), nullable(e.RunID),
		e.Timestamp, e.Payload)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Count returns the number of stored events (diagnostic helper).
func (s *EventStore) Count(ctx context.Context) (int, error) {
	var n int
	err := s.db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM events`).Scan(&n)
	return n, err
}

// QueryByCorrelation returns stored events sharing a correlation ID, in insertion order.
func (s *EventStore) QueryByCorrelation(ctx context.Context, id core.ULID) ([]ports.Event, error) {
	rows, err := s.db.sql.QueryContext(ctx, `
SELECT name, version, producer, correlation_id, session_id, run_id, ts, payload
FROM events WHERE correlation_id = ? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.Event
	for rows.Next() {
		var e ports.Event
		var corr, sess, run *string
		if err := rows.Scan(&e.Name, &e.Version, &e.Producer, &corr, &sess, &run, &e.Timestamp, &e.Payload); err != nil {
			return nil, err
		}
		e.CorrelationID = deref(corr)
		e.SessionID = deref(sess)
		e.RunID = deref(run)
		out = append(out, e)
	}
	return out, rows.Err()
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
