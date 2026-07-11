package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (ADR-007), registered as "sqlite"
)

// ErrFutureSchema is returned when a database was written by a newer schema version than this
// build knows. Opening it is refused (ADR-029); the caller maps it to exit code 9.
var ErrFutureSchema = errors.New("storage: database schema is newer than this build supports")

// ErrIntegrity is returned when a post-migration integrity check fails (ADR-029, exit code 9).
var ErrIntegrity = errors.New("storage: integrity check failed")

// DB is an open Andromeda database with its migration set applied.
type DB struct {
	sql  *sql.DB
	path string
	role string // "workspace" | "global"
}

// dsn builds a modernc.org/sqlite DSN with WAL mode, a busy timeout, and foreign keys on.
func dsn(path string) string {
	return "file:" + path +
		"?_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)" +
		"&_pragma=foreign_keys(ON)"
}

// open opens (creating if needed) the database at path, applies migrations, and verifies
// integrity. role is used only for diagnostics.
func open(ctx context.Context, path, role string, migs []Migration) (*DB, error) {
	sqldb, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("open %s db: %w", role, err)
	}
	// modernc.org/sqlite is a single logical connection engine; one open connection avoids
	// WAL/locking surprises for an embedded single-writer store.
	sqldb.SetMaxOpenConns(1)
	if err := sqldb.PingContext(ctx); err != nil {
		_ = sqldb.Close()
		return nil, fmt.Errorf("ping %s db: %w", role, err)
	}
	db := &DB{sql: sqldb, path: path, role: role}
	if err := db.migrate(ctx, migs); err != nil {
		_ = sqldb.Close()
		return nil, err
	}
	return db, nil
}

// SQL exposes the underlying *sql.DB for schema-owning packages in higher layers.
func (d *DB) SQL() *sql.DB { return d.sql }

// Path returns the database file path.
func (d *DB) Path() string { return d.path }

// SchemaVersion returns the database's current user_version.
func (d *DB) SchemaVersion(ctx context.Context) (int, error) {
	var v int
	err := d.sql.QueryRowContext(ctx, "PRAGMA user_version").Scan(&v)
	return v, err
}

// Close closes the database.
func (d *DB) Close() error { return d.sql.Close() }
