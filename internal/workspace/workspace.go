package workspace

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/git"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
)

// Engine implements ports.WorkspacePort.
type Engine struct {
	global *storage.DB
	git    *git.Engine

	mu   sync.Mutex
	open map[core.ULID]*openWorkspace
}

type openWorkspace struct {
	root string
	db   *storage.DB
}

// New returns a Workspace Engine over the machine-global database and a Git Engine (for VCS
// summaries in snapshots; may be nil).
func New(global *storage.DB, g *git.Engine) *Engine {
	return &Engine{global: global, git: g, open: map[core.ULID]*openWorkspace{}}
}

var _ ports.WorkspacePort = (*Engine)(nil)

// Discover walks upward from start to locate a workspace root, reporting what it found.
func (e *Engine) Discover(ctx context.Context, start ports.Path) (ports.WorkspaceCandidate, error) {
	if err := ctx.Err(); err != nil {
		return ports.WorkspaceCandidate{}, err
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return ports.WorkspaceCandidate{}, err
	}
	for {
		if isDir(filepath.Join(dir, storage.MarkerDir)) {
			return ports.WorkspaceCandidate{Root: dir, Found: true, HasMarker: true, IsRepo: isDir(filepath.Join(dir, ".git"))}, nil
		}
		if isDir(filepath.Join(dir, ".git")) {
			return ports.WorkspaceCandidate{Root: dir, Found: true, HasMarker: false, IsRepo: true}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ports.WorkspaceCandidate{Found: false}, nil
		}
		dir = parent
	}
}

// Open opens (creating .andromeda/ when options allow) a workspace and registers it globally.
func (e *Engine) Open(ctx context.Context, root ports.Path, opts ports.OpenOptions) (ports.WorkspaceHandle, error) {
	if err := ctx.Err(); err != nil {
		return ports.WorkspaceHandle{}, err
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return ports.WorkspaceHandle{}, err
	}
	marker := filepath.Join(abs, storage.MarkerDir)
	if !isDir(marker) && !opts.CreateIfMissing {
		return ports.WorkspaceHandle{}, &ports.PortError{
			Code: "E-AGT-010", Category: "agent", Severity: "error",
			Message: "not an Andromeda workspace", Detail: abs,
		}
	}
	db, err := storage.OpenWorkspaceDB(ctx, abs)
	if err != nil {
		return ports.WorkspaceHandle{}, &ports.PortError{
			Code: "E-AGT-011", Category: "agent", Severity: "error",
			Message: "failed to open workspace database", Detail: err.Error(), Cause: err,
		}
	}
	id, err := e.register(ctx, abs)
	if err != nil {
		_ = db.Close()
		return ports.WorkspaceHandle{}, err
	}
	e.mu.Lock()
	e.open[id] = &openWorkspace{root: abs, db: db}
	e.mu.Unlock()
	return ports.WorkspaceHandle{ID: id, Root: abs}, nil
}

// register inserts or updates the global workspaces registry and returns the workspace ULID.
func (e *Engine) register(ctx context.Context, root string) (core.ULID, error) {
	if e.global == nil {
		return core.NewULID(), nil // registry optional (e.g. tests without a global DB)
	}
	var id core.ULID
	err := e.global.SQL().QueryRowContext(ctx, `SELECT id FROM workspaces WHERE root = ?`, root).Scan(&id)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if errors.Is(err, sql.ErrNoRows) {
		id = core.NewULID()
		_, err = e.global.SQL().ExecContext(ctx,
			`INSERT INTO workspaces (id, root, created_at, last_opened_at) VALUES (?, ?, ?, ?)`,
			id, root, now, now)
		return id, err
	}
	if err != nil {
		return "", err
	}
	_, err = e.global.SQL().ExecContext(ctx, `UPDATE workspaces SET last_opened_at = ? WHERE id = ?`, now, id)
	return id, err
}

// Snapshot returns a consistent read-only description of workspace state.
func (e *Engine) Snapshot(ctx context.Context, ws ports.WorkspaceHandle) (ports.WorkspaceSnapshot, error) {
	e.mu.Lock()
	ow, ok := e.open[ws.ID]
	e.mu.Unlock()
	if !ok {
		return ports.WorkspaceSnapshot{}, &ports.PortError{Code: "E-AGT-012", Category: "agent", Message: "workspace not open"}
	}
	snap := ports.WorkspaceSnapshot{
		WorkspaceID: ws.ID,
		Root:        ow.root,
		Projects:    []ports.Path{ow.root},
		TakenAt:     time.Now().UTC().Format(time.RFC3339Nano),
	}
	if e.git != nil {
		if st, err := e.git.Status(ctx, ports.RepoRef{Root: ow.root}); err == nil {
			summary := "branch " + st.Branch
			if !st.Clean {
				summary += " (dirty)"
			}
			snap.VCSSummary = summary
		}
	}
	return snap, nil
}

// Close detaches a workspace: closes its database and forgets the handle.
func (e *Engine) Close(ctx context.Context, ws ports.WorkspaceHandle) error {
	e.mu.Lock()
	ow, ok := e.open[ws.ID]
	if ok {
		delete(e.open, ws.ID)
	}
	e.mu.Unlock()
	if !ok {
		return nil
	}
	return ow.db.Close()
}

// WorkspaceDB returns the open workspace's database (for engines that persist into it).
func (e *Engine) WorkspaceDB(ws ports.WorkspaceHandle) (*storage.DB, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	ow, ok := e.open[ws.ID]
	if !ok {
		return nil, false
	}
	return ow.db, true
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}
