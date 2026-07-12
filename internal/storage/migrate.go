package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Migration is one forward-only schema step. Version numbers are contiguous starting at 1;
// SQL runs inside a transaction. Migrations are append-only: once released, a migration's SQL
// is never edited (ADR-029), only new higher-versioned migrations are added.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// migrate applies every migration whose version exceeds the database's current user_version.
// Before applying any, it backs up the database file (when it exists on disk). After applying,
// it runs an integrity check. A database whose user_version exceeds the highest known
// migration is refused with ErrFutureSchema.
func (d *DB) migrate(ctx context.Context, migs []Migration) error {
	current, err := d.SchemaVersion(ctx)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	highest := 0
	for _, m := range migs {
		if m.Version > highest {
			highest = m.Version
		}
	}
	if current > highest {
		return fmt.Errorf("%w: db=%d build=%d", ErrFutureSchema, current, highest)
	}
	if current == highest {
		return nil // up to date
	}

	// Back up before mutating an existing on-disk database (ADR-029 recovery path).
	if current > 0 {
		if err := d.backup(); err != nil {
			return fmt.Errorf("pre-migration backup: %w", err)
		}
	}

	for _, m := range migs {
		if m.Version <= current {
			continue
		}
		if err := d.applyOne(ctx, m); err != nil {
			return fmt.Errorf("apply migration %d (%s): %w", m.Version, m.Name, err)
		}
	}

	if err := d.integrityCheck(ctx); err != nil {
		return err
	}
	return nil
}

func (d *DB) applyOne(ctx context.Context, m Migration) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		_ = tx.Rollback()
		return err
	}
	// user_version cannot be parameterized; the value is an int we control, not user input.
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", m.Version)); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (d *DB) integrityCheck(ctx context.Context) error {
	var result string
	if err := d.sql.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result); err != nil {
		return fmt.Errorf("%w: %v", ErrIntegrity, err)
	}
	if result != "ok" {
		return fmt.Errorf("%w: %s", ErrIntegrity, result)
	}
	if _, err := d.sql.ExecContext(ctx, "PRAGMA foreign_key_check"); err != nil {
		return fmt.Errorf("%w: foreign_key_check: %v", ErrIntegrity, err)
	}
	return nil
}

// backup copies the database file (and, best-effort, its WAL sidecar) to a timestamped file
// beside it. In-memory databases have no on-disk file and are skipped.
func (d *DB) backup() error {
	if d.path == "" || d.path == ":memory:" {
		return nil
	}
	if _, err := os.Stat(d.path); err != nil {
		return nil //nolint:nilerr // nothing on disk yet to back up
	}
	// Checkpoint the WAL so the main file is complete before copying.
	_, _ = d.sql.Exec("PRAGMA wal_checkpoint(FULL)")

	stamp := time.Now().UTC().Format("20060102T150405Z")
	dst := fmt.Sprintf("%s.backup-%s", d.path, stamp)
	return copyFile(d.path, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // path is an internal database file, not user input
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) //nolint:gosec // path is an internal backup file, not user input
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
