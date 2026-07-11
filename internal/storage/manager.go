package storage

import (
	"context"
	"os"
	"path/filepath"
)

// WorkspaceDBName and GlobalDBName are the fixed database filenames (ADR-028).
const (
	WorkspaceDBName = "state.db"
	GlobalDBName    = "global.db"
	MarkerDir       = ".andromeda"
)

// globalMigrations is the append-only migration set for the machine-global database.
var globalMigrations = []Migration{
	{
		Version: 1,
		Name:    "init_global",
		SQL: `
CREATE TABLE meta (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL
);
CREATE TABLE workspaces (
  id             TEXT PRIMARY KEY,
  root           TEXT NOT NULL UNIQUE,
  created_at     TEXT NOT NULL,
  last_opened_at TEXT NOT NULL
);`,
	},
}

// workspaceMigrations is the append-only migration set for a workspace database. Session and
// run tables live here (durability behind PRD-010); later epics add their own migrations.
var workspaceMigrations = []Migration{
	{
		Version: 1,
		Name:    "init_workspace",
		SQL: `
CREATE TABLE meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE sessions (
  id         TEXT PRIMARY KEY,
  state      TEXT NOT NULL,
  revision   INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  data       BLOB
);
CREATE TABLE runs (
  id         TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  state      TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX idx_runs_session ON runs(session_id);
CREATE TABLE run_records (
  id         TEXT PRIMARY KEY,
  run_id     TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  seq        INTEGER NOT NULL,
  kind       TEXT NOT NULL,
  payload    BLOB,
  created_at TEXT NOT NULL
);
CREATE INDEX idx_run_records_run ON run_records(run_id, seq);`,
	},
	{
		Version: 2,
		Name:    "add_events",
		SQL: `
CREATE TABLE events (
  id             TEXT PRIMARY KEY,
  name           TEXT NOT NULL,
  version        INTEGER NOT NULL,
  producer       TEXT NOT NULL,
  correlation_id TEXT,
  session_id     TEXT,
  run_id         TEXT,
  ts             TEXT NOT NULL,
  payload        BLOB
);
CREATE INDEX idx_events_name ON events(name);
CREATE INDEX idx_events_correlation ON events(correlation_id);
CREATE INDEX idx_events_run ON events(run_id);`,
	},
	{
		Version: 3,
		Name:    "add_permissions",
		SQL: `
CREATE TABLE grants (
  id                TEXT PRIMARY KEY,
  permission        TEXT NOT NULL,
  scope             TEXT NOT NULL,
  selector          TEXT NOT NULL,
  effect            TEXT NOT NULL,
  subject_session   TEXT,
  subject_workspace TEXT,
  valid_until       TEXT,
  revoked           INTEGER NOT NULL DEFAULT 0,
  created_at        TEXT NOT NULL
);
CREATE INDEX idx_grants_permission ON grants(permission);
CREATE TABLE permission_audit (
  id         TEXT PRIMARY KEY,
  permission TEXT NOT NULL,
  scope      TEXT NOT NULL,
  subject    TEXT NOT NULL,
  outcome    TEXT NOT NULL,
  deciding   TEXT,
  actor      TEXT,
  ts         TEXT NOT NULL
);
CREATE INDEX idx_permission_audit_ts ON permission_audit(ts);`,
	},
	{
		Version: 4,
		Name:    "add_memory",
		SQL: `
CREATE TABLE memory_records (
  id         TEXT PRIMARY KEY,
  layer      TEXT NOT NULL,
  content    TEXT NOT NULL,
  provenance TEXT,
  source     TEXT,
  status     TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL
);
CREATE INDEX idx_memory_layer ON memory_records(layer);
CREATE INDEX idx_memory_status ON memory_records(status);`,
	},
}

// OpenWorkspaceDB opens (creating on first use) the workspace database under
// <root>/.andromeda/state.db, applying the workspace migration set.
func OpenWorkspaceDB(ctx context.Context, root string) (*DB, error) {
	dir := filepath.Join(root, MarkerDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return open(ctx, filepath.Join(dir, WorkspaceDBName), "workspace", workspaceMigrations)
}

// OpenGlobalDB opens (creating on first use) the machine-global database under
// <dataDir>/global.db, applying the global migration set.
func OpenGlobalDB(ctx context.Context, dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, err
	}
	return open(ctx, filepath.Join(dataDir, GlobalDBName), "global", globalMigrations)
}

// openMemory opens an in-memory database with the given migrations, for tests.
func openMemory(ctx context.Context, migs []Migration) (*DB, error) {
	return open(ctx, ":memory:", "memory", migs)
}
