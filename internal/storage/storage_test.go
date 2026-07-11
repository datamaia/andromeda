package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

func TestBackupOnMigration(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "t.db")
	m1 := []Migration{{Version: 1, Name: "a", SQL: "CREATE TABLE t (x INTEGER)"}}
	db, err := open(ctx, path, "test", m1)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	m2 := append(m1, Migration{Version: 2, Name: "b", SQL: "CREATE TABLE u (y INTEGER)"})
	db2, err := open(ctx, path, "test", m2)
	if err != nil {
		t.Fatalf("migrate v2: %v", err)
	}
	defer db2.Close()
	if v, _ := db2.SchemaVersion(ctx); v != 2 {
		t.Fatalf("version = %d, want 2", v)
	}
	backups, _ := filepath.Glob(path + ".backup-*")
	if len(backups) == 0 {
		t.Error("expected a pre-migration backup file")
	}
}

func TestListSessionsAndRunsWithFilters(t *testing.T) {
	ctx := context.Background()
	db, err := openMemory(ctx, workspaceMigrations)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewSessionStore(db)
	s1, s2 := core.NewULID(), core.NewULID()
	_ = s.SaveSession(ctx, ports.SessionSnapshot{ID: s1, State: "active"})
	_ = s.SaveSession(ctx, ports.SessionSnapshot{ID: s2, State: "ended"})

	active, err := s.ListSessions(ctx, ports.SessionFilter{State: "active", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].ID != s1 {
		t.Fatalf("active sessions = %v", active)
	}
	all, _ := s.ListSessions(ctx, ports.SessionFilter{})
	if len(all) != 2 {
		t.Fatalf("all sessions = %d, want 2", len(all))
	}

	r1 := core.NewULID()
	_ = s.CreateRun(ctx, r1, s1, "running")
	runs, _ := s.ListRuns(ctx, ports.RunFilter{SessionID: s1, State: "running", Limit: 5})
	if len(runs) != 1 || runs[0].ID != r1 {
		t.Fatalf("runs = %v", runs)
	}
}

func TestLoadNotFound(t *testing.T) {
	ctx := context.Background()
	db, _ := openMemory(ctx, workspaceMigrations)
	defer db.Close()
	s := NewSessionStore(db)
	if _, err := s.LoadSession(ctx, core.NewULID()); err != ErrNotFound {
		t.Errorf("LoadSession: want ErrNotFound, got %v", err)
	}
	if _, err := s.LoadRun(ctx, core.NewULID()); err != ErrNotFound {
		t.Errorf("LoadRun: want ErrNotFound, got %v", err)
	}
}

func TestOpenWorkspaceAndGlobalDBs(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	ws, err := OpenWorkspaceDB(ctx, root)
	if err != nil {
		t.Fatalf("open workspace: %v", err)
	}
	defer ws.Close()
	if _, err := ws.SQL().Exec(`INSERT INTO meta(key,value) VALUES('schema','ws')`); err != nil {
		t.Fatalf("workspace write: %v", err)
	}
	v, err := ws.SchemaVersion(ctx)
	if err != nil || v != 1 {
		t.Fatalf("workspace schema version = %d, %v; want 1", v, err)
	}
	if _, statErr := ws.SQL().Exec(`SELECT 1`); statErr != nil {
		t.Fatalf("workspace db not usable: %v", statErr)
	}

	data := t.TempDir()
	g, err := OpenGlobalDB(ctx, data)
	if err != nil {
		t.Fatalf("open global: %v", err)
	}
	defer g.Close()
	if g.Path() != filepath.Join(data, GlobalDBName) {
		t.Errorf("global path = %q", g.Path())
	}
}

func TestMigrationIsIdempotentAcrossReopen(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	ws1, err := OpenWorkspaceDB(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	sid := core.NewULID()
	store := NewSessionStore(ws1)
	if err := store.SaveSession(ctx, ports.SessionSnapshot{ID: sid, State: "active"}); err != nil {
		t.Fatal(err)
	}
	ws1.Close()

	// Reopen: migrations already applied, data survives.
	ws2, err := OpenWorkspaceDB(ctx, root)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer ws2.Close()
	got, err := NewSessionStore(ws2).LoadSession(ctx, sid)
	if err != nil {
		t.Fatalf("load after reopen: %v", err)
	}
	if got.State != "active" {
		t.Errorf("state = %q, want active", got.State)
	}
}

func TestFutureSchemaRefused(t *testing.T) {
	ctx := context.Background()
	db, err := openMemory(ctx, workspaceMigrations)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a database written by a newer build.
	if _, err := db.SQL().Exec("PRAGMA user_version = 999"); err != nil {
		t.Fatal(err)
	}
	if err := db.migrate(ctx, workspaceMigrations); err == nil {
		t.Fatal("expected ErrFutureSchema")
	} else if !isFutureSchema(err) {
		t.Fatalf("want ErrFutureSchema, got %v", err)
	}
	db.Close()
}

func isFutureSchema(err error) bool {
	for err != nil {
		if err == ErrFutureSchema {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

func TestSessionRevisionConflict(t *testing.T) {
	ctx := context.Background()
	db, err := openMemory(ctx, workspaceMigrations)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewSessionStore(db)
	id := core.NewULID()
	if err := store.SaveSession(ctx, ports.SessionSnapshot{ID: id, State: "active", Revision: 0}); err != nil {
		t.Fatal(err)
	}
	// Update at the correct revision (0) succeeds and bumps to 1.
	if err := store.SaveSession(ctx, ports.SessionSnapshot{ID: id, State: "suspended", Revision: 0}); err != nil {
		t.Fatalf("expected update to succeed: %v", err)
	}
	// A stale writer at revision 0 now conflicts.
	if err := store.SaveSession(ctx, ports.SessionSnapshot{ID: id, State: "ended", Revision: 0}); err != ErrRevisionConflict {
		t.Fatalf("want ErrRevisionConflict, got %v", err)
	}
}

func TestRunRecordsAndMarkInterrupted(t *testing.T) {
	ctx := context.Background()
	db, err := openMemory(ctx, workspaceMigrations)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewSessionStore(db)
	sid, rid := core.NewULID(), core.NewULID()
	if err := store.SaveSession(ctx, ports.SessionSnapshot{ID: sid, State: "active"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateRun(ctx, rid, sid, "running"); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendRunRecords(ctx, rid, []ports.RunRecord{{Kind: "turn"}, {Kind: "message"}}); err != nil {
		t.Fatal(err)
	}
	run, err := store.LoadRun(ctx, rid)
	if err != nil {
		t.Fatal(err)
	}
	if len(run.Records) != 2 {
		t.Fatalf("records = %d, want 2", len(run.Records))
	}
	ids, err := store.MarkInterrupted(ctx, ports.InterruptScope{SessionID: sid})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != rid {
		t.Fatalf("interrupted = %v, want [%s]", ids, rid)
	}
	// Running again marks nothing (now interrupted, still non-terminal → it re-marks).
	run2, _ := store.LoadRun(ctx, rid)
	if run2.State != "interrupted" {
		t.Errorf("run state = %q, want interrupted", run2.State)
	}
}
