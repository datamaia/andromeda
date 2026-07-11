package tool_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/permission"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/datamaia/andromeda/internal/tool"
	"github.com/datamaia/andromeda/internal/tool/builtin"
)

func newRuntime(t *testing.T, grantAll bool) (*tool.Runtime, *permission.Manager) {
	t.Helper()
	db, err := storage.OpenWorkspaceDB(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	pm := permission.NewManager(permission.NewStore(db))
	if grantAll {
		ctx := context.Background()
		pm.GrantPermission(ctx, permission.Grant{Permission: core.PermRead, Scope: core.ScopePath, Selector: "*", Effect: permission.EffectAllow})
		pm.GrantPermission(ctx, permission.Grant{Permission: core.PermWrite, Scope: core.ScopePath, Selector: "*", Effect: permission.EffectAllow})
	}
	rt := tool.NewRuntime(pm)
	ctx := context.Background()
	for _, tl := range []ports.ToolPort{builtin.FSRead{}, builtin.FSWrite{}, builtin.FSSearch{}} {
		if err := rt.Register(ctx, tl); err != nil {
			t.Fatal(err)
		}
	}
	return rt, pm
}

func drain(t *testing.T, st ports.Stream[ports.ToolEvent]) ports.ToolEvent {
	t.Helper()
	var terminal ports.ToolEvent
	for {
		ev, err := st.Next(context.Background())
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if ev.Terminal {
			terminal = ev
		}
	}
	return terminal
}

func TestWriteThenReadWithPermission(t *testing.T) {
	ctx := context.Background()
	rt, _ := newRuntime(t, true)
	dir := t.TempDir()
	file := filepath.Join(dir, "note.txt")

	winput, _ := json.Marshal(map[string]string{"path": file, "content": "hello tools"})
	st, err := rt.Invoke(ctx, "fs_write", ports.PermissionQuery{}, winput)
	if err != nil {
		t.Fatal(err)
	}
	if term := drain(t, st); term.Outcome != "success" {
		t.Fatalf("write outcome = %q (%s)", term.Outcome, term.Text)
	}
	if data, _ := os.ReadFile(file); string(data) != "hello tools" {
		t.Fatalf("file content = %q", data)
	}

	rinput, _ := json.Marshal(map[string]string{"path": file})
	st2, _ := rt.Invoke(ctx, "fs_read", ports.PermissionQuery{}, rinput)
	term := drain(t, st2)
	if term.Outcome != "success" {
		t.Fatalf("read outcome = %q", term.Outcome)
	}
	var out struct{ Content string }
	json.Unmarshal([]byte(term.Text), &out)
	if out.Content != "hello tools" {
		t.Errorf("read content = %q", out.Content)
	}
}

func TestDenialIsDeliveredAsData(t *testing.T) {
	ctx := context.Background()
	rt, _ := newRuntime(t, false) // no grants → default ask → non-interactive deny
	rinput, _ := json.Marshal(map[string]string{"path": "/etc/hosts"})
	st, err := rt.Invoke(ctx, "fs_read", ports.PermissionQuery{}, rinput)
	if err != nil {
		t.Fatalf("invoke should not error on denial: %v", err)
	}
	term := drain(t, st)
	if term.Outcome != "error" || term.Text == "" {
		t.Fatalf("expected a denial terminal event, got %+v", term)
	}
}

func TestValidationFailureIsAnError(t *testing.T) {
	ctx := context.Background()
	rt, _ := newRuntime(t, true)
	_, err := rt.Invoke(ctx, "fs_read", ports.PermissionQuery{}, []byte(`{}`))
	if err == nil {
		t.Fatal("expected a validation error for missing path")
	}
}

func TestUnknownTool(t *testing.T) {
	ctx := context.Background()
	rt, _ := newRuntime(t, true)
	if _, err := rt.Invoke(ctx, "nope", ports.PermissionQuery{}, []byte(`{}`)); err == nil {
		t.Fatal("expected unknown-tool error")
	}
}

func TestSearchFindsMatches(t *testing.T) {
	ctx := context.Background()
	rt, _ := newRuntime(t, true)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "x.txt"), []byte("alpha\nNEEDLE here\n"), 0o600)
	input, _ := json.Marshal(map[string]string{"query": "needle", "root": dir})
	st, _ := rt.Invoke(ctx, "fs_search", ports.PermissionQuery{}, input)
	term := drain(t, st)
	var out struct {
		Count int `json:"count"`
	}
	json.Unmarshal([]byte(term.Text), &out)
	if out.Count != 1 {
		t.Errorf("search count = %d, want 1", out.Count)
	}
}
