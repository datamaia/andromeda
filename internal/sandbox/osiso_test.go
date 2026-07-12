package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

func TestContainmentLevelReportsHonestly(t *testing.T) {
	ctx := context.Background()
	e := New()
	sb, _ := e.Prepare(ctx, ports.SandboxSpec{WorkingDir: t.TempDir()})

	// Requesting OS isolation yields "os" only where the platform supports it; else "process".
	_ = e.ApplyPolicy(ctx, sb, ports.SandboxPolicy{Isolation: "os"})
	level := e.ContainmentLevel(sb)
	if osIsolationSupported() {
		if level != "os" {
			t.Errorf("expected os containment where supported, got %q", level)
		}
	} else if level != "process" {
		t.Errorf("expected process containment where OS isolation is unsupported, got %q", level)
	}

	// A bare policy is always process-level.
	_ = e.ApplyPolicy(ctx, sb, ports.SandboxPolicy{})
	if e.ContainmentLevel(sb) != "process" {
		t.Errorf("bare policy should be process-level, got %q", e.ContainmentLevel(sb))
	}
}

func TestOSIsolationEnforcesWritePolicy(t *testing.T) {
	if !osIsolationSupported() {
		t.Skip("OS-level isolation not available on this platform")
	}
	ctx := context.Background()
	e := New()
	work := t.TempDir()
	allowed := filepath.Join(work, "allowed")
	denied := filepath.Join(work, "denied")
	os.MkdirAll(allowed, 0o755)
	os.MkdirAll(denied, 0o755)

	sb, _ := e.Prepare(ctx, ports.SandboxSpec{WorkingDir: work})
	_ = e.ApplyPolicy(ctx, sb, ports.SandboxPolicy{
		Isolation:     "os",
		WritePaths:    []ports.Path{allowed},
		NetworkPolicy: "deny",
		CommandAllow:  []string{"sh"},
	})
	if e.ContainmentLevel(sb) != "os" {
		t.Fatalf("expected os containment, got %q", e.ContainmentLevel(sb))
	}

	// A write inside the allowed subpath succeeds.
	idOK, err := e.ExecuteIn(ctx, sb, ports.CommandSpec{Program: "sh", Args: []string{"-c", "echo ok > " + filepath.Join(allowed, "f")}})
	if err != nil {
		t.Fatalf("execute allowed: %v", err)
	}
	if out, _ := e.Wait(ctx, sb, idOK); out.Status != "succeeded" {
		t.Fatalf("allowed write outcome = %+v", out)
	}
	if _, err := os.Stat(filepath.Join(allowed, "f")); err != nil {
		t.Errorf("allowed file was not written: %v", err)
	}

	// A write outside the allowed subpath is denied by the sandbox profile (non-zero exit).
	idBad, err := e.ExecuteIn(ctx, sb, ports.CommandSpec{Program: "sh", Args: []string{"-c", "echo no > " + filepath.Join(denied, "f")}})
	if err != nil {
		t.Fatalf("execute denied: %v", err)
	}
	out, _ := e.Wait(ctx, sb, idBad)
	if out.Status == "succeeded" {
		t.Errorf("write outside the sandbox write policy should have failed")
	}
	if _, err := os.Stat(filepath.Join(denied, "f")); err == nil {
		t.Errorf("file outside the sandbox policy was written")
	}
}
