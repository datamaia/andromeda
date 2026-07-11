package builtin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

func terminalOf(t *testing.T, st ports.Stream[ports.ToolEvent]) ports.ToolEvent {
	t.Helper()
	var term ports.ToolEvent
	for {
		ev, err := st.Next(context.Background())
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if ev.Terminal {
			term = ev
		}
	}
	return term
}

func TestFSReadDescribeAndExecute(t *testing.T) {
	ctx := context.Background()
	d, _ := FSRead{}.Describe(ctx)
	if d.Name != "fs_read" || len(d.Permissions) != 1 || d.Permissions[0] != core.PermRead {
		t.Fatalf("descriptor = %+v", d)
	}
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	os.WriteFile(f, []byte("data"), 0o600)
	in, _ := json.Marshal(map[string]string{"path": f})
	st, _ := FSRead{}.Execute(ctx, ports.ToolExecuteRequest{Input: in})
	if term := terminalOf(t, st); term.Outcome != "success" {
		t.Fatalf("outcome = %q", term.Outcome)
	}
}

func TestFSReadMissingFile(t *testing.T) {
	in, _ := json.Marshal(map[string]string{"path": "/no/such/file"})
	st, _ := FSRead{}.Execute(context.Background(), ports.ToolExecuteRequest{Input: in})
	if term := terminalOf(t, st); term.Outcome != "error" {
		t.Errorf("missing file should error, got %q", term.Outcome)
	}
}

func TestFSWriteResourcesArePathScoped(t *testing.T) {
	in, _ := json.Marshal(map[string]string{"path": "/tmp/x", "content": "y"})
	qs, err := FSWrite{}.Resources(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 || qs[0].Permission != core.PermWrite || qs[0].Subject != "/tmp/x" {
		t.Fatalf("resources = %+v", qs)
	}
}

func TestValidation(t *testing.T) {
	ctx := context.Background()
	if r, _ := (FSRead{}).Validate(ctx, []byte(`{}`)); r.Valid {
		t.Error("fs_read should require path")
	}
	if r, _ := (FSWrite{}).Validate(ctx, []byte(`{"path":"x"}`)); r.Valid {
		t.Error("fs_write should require content")
	}
	if r, _ := (FSSearch{}).Validate(ctx, []byte(`{"query":"q"}`)); !r.Valid {
		t.Error("fs_search with query should be valid")
	}
}
