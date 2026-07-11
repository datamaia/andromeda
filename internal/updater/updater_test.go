package updater

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// dirSource serves a single version from a local file.
type dirSource struct {
	version  string
	path     string
	checksum string
}

func (d dirSource) Latest(context.Context, string) (string, string, error) {
	return d.version, d.checksum, nil
}
func (d dirSource) Fetch(context.Context, string) (string, string, error) {
	return d.path, d.checksum, nil
}

func makeArtifact(t *testing.T, content string) (path, checksum string) {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "artifact")
	if err := os.WriteFile(p, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	cs, err := sha256File(p)
	if err != nil {
		t.Fatal(err)
	}
	return p, cs
}

func TestCheckUpToDateAndAvailable(t *testing.T) {
	ctx := context.Background()
	art, cs := makeArtifact(t, "v2 binary")
	u := New("1.0.0", "stable", filepath.Join(t.TempDir(), "andromeda"), dirSource{version: "1.1.0", path: art, checksum: cs})
	res, err := u.Check(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "update_available" || res.Latest != "1.1.0" {
		t.Fatalf("check = %+v", res)
	}
}

func TestVerifyApplyRollback(t *testing.T) {
	ctx := context.Background()
	art, cs := makeArtifact(t, "new binary")
	target := filepath.Join(t.TempDir(), "andromeda")
	os.WriteFile(target, []byte("old binary"), 0o755)

	u := New("1.0.0", "stable", target, dirSource{version: "1.1.0", path: art, checksum: cs})
	rel := ports.ReleaseRef{Version: "1.1.0"}

	// Apply before verify is refused.
	if _, err := u.Apply(ctx, rel); err == nil {
		t.Fatal("apply before verify must be refused")
	}

	rep, err := u.Verify(ctx, rel)
	if err != nil || !rep.OK {
		t.Fatalf("verify = %+v err=%v", rep, err)
	}
	ar, err := u.Apply(ctx, rel)
	if err != nil || !ar.Applied || ar.ToVersion != "1.1.0" {
		t.Fatalf("apply = %+v err=%v", ar, err)
	}
	if data, _ := os.ReadFile(target); string(data) != "new binary" {
		t.Fatalf("target after apply = %q", data)
	}
	if _, err := u.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(target); string(data) != "old binary" {
		t.Fatalf("target after rollback = %q", data)
	}
}

func TestVerifyDetectsChecksumMismatch(t *testing.T) {
	ctx := context.Background()
	art, _ := makeArtifact(t, "content")
	u := New("1.0.0", "stable", filepath.Join(t.TempDir(), "a"), dirSource{version: "2", path: art, checksum: "deadbeef"})
	rep, _ := u.Verify(ctx, ports.ReleaseRef{Version: "2"})
	if rep.OK {
		t.Fatal("checksum mismatch should not verify")
	}
}

func TestCheckNoSourceIsClean(t *testing.T) {
	res, err := New("1.0.0", "stable", "x", nil).Check(context.Background())
	if err != nil || res.Status != "up_to_date" {
		t.Fatalf("no-source check = %+v err=%v", res, err)
	}
}
