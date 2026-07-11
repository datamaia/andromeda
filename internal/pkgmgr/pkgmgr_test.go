package pkgmgr

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

type dirSource struct {
	version, path, checksum string
	err                     error
}

func (d dirSource) Resolve(context.Context, ports.PackageRequest) (string, string, string, error) {
	return d.version, d.path, d.checksum, d.err
}

func artifact(t *testing.T, content string) (path, checksum string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "pkg.bin")
	os.WriteFile(p, []byte(content), 0o644)
	cs, err := sha256File(p)
	if err != nil {
		t.Fatal(err)
	}
	return p, cs
}

func drainStates(t *testing.T, st ports.Stream[ports.InstallEvent]) []string {
	t.Helper()
	var states []string
	for {
		ev, err := st.Next(context.Background())
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		states = append(states, ev.State)
	}
	return states
}

func TestResolveInstallVerifyRemove(t *testing.T) {
	ctx := context.Background()
	path, cs := artifact(t, "tool bytes")
	installDir := t.TempDir()
	m := New(installDir, dirSource{version: "1.0.0", path: path, checksum: cs})

	plan, err := m.Resolve(ctx, ports.PackageRequest{Name: "cool-tool", Constraint: "^1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Packages) != 1 || plan.Packages[0].Version != "1.0.0" {
		t.Fatalf("plan = %+v", plan)
	}

	st, err := m.Install(ctx, plan)
	if err != nil {
		t.Fatal(err)
	}
	states := drainStates(t, st)
	if states[len(states)-1] != "installed" {
		t.Fatalf("install states = %v", states)
	}

	pkg := plan.Packages[0]
	if rep, _ := m.Verify(ctx, pkg); !rep.OK {
		t.Fatal("installed package should verify")
	}
	if rr, err := m.Remove(ctx, pkg, ports.RemoveOptions{}); err != nil || !rr.Removed {
		t.Fatalf("remove = %+v err=%v", rr, err)
	}
	if rep, _ := m.Verify(ctx, pkg); rep.OK {
		t.Fatal("removed package should not verify")
	}
}

func TestInstallFailsOnChecksumMismatch(t *testing.T) {
	ctx := context.Background()
	path, _ := artifact(t, "bytes")
	m := New(t.TempDir(), dirSource{version: "1", path: path, checksum: "wrong"})
	plan, _ := m.Resolve(ctx, ports.PackageRequest{Name: "bad"})
	states := drainStates(t, mustInstall(t, m, plan))
	if states[len(states)-1] != "failed" {
		t.Fatalf("expected failed terminal state, got %v", states)
	}
}

func mustInstall(t *testing.T, m *Manager, plan ports.ResolutionPlan) ports.Stream[ports.InstallEvent] {
	t.Helper()
	st, err := m.Install(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func TestResolveWithoutSourceErrors(t *testing.T) {
	if _, err := New(t.TempDir(), nil).Resolve(context.Background(), ports.PackageRequest{Name: "x"}); err == nil {
		t.Fatal("expected error without a source")
	}
}
