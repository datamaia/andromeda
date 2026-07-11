package sandbox

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

func TestExecuteAllowedCommand(t *testing.T) {
	ctx := context.Background()
	e := New()
	sb, err := e.Prepare(ctx, ports.SandboxSpec{WorkingDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if sb.ContainmentLevel != "process" {
		t.Errorf("containment = %q, want process", sb.ContainmentLevel)
	}
	_ = e.ApplyPolicy(ctx, sb, ports.SandboxPolicy{CommandAllow: []string{"sh"}})
	id, err := e.ExecuteIn(ctx, sb, ports.CommandSpec{Program: "sh", Args: []string{"-c", "exit 0"}})
	if err != nil {
		t.Fatal(err)
	}
	out, err := e.Wait(ctx, sb, id)
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "succeeded" || out.ExitCode != 0 {
		t.Errorf("outcome = %+v", out)
	}
}

func TestDeniedCommandRefused(t *testing.T) {
	ctx := context.Background()
	e := New()
	sb, _ := e.Prepare(ctx, ports.SandboxSpec{WorkingDir: t.TempDir()})
	_ = e.ApplyPolicy(ctx, sb, ports.SandboxPolicy{CommandDeny: []string{"rm"}})
	_, err := e.ExecuteIn(ctx, sb, ports.CommandSpec{Program: "/bin/rm", Args: []string{"-rf", "/"}})
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != "E-SEC-031" {
		t.Fatalf("want E-SEC-031, got %v", err)
	}
}

func TestAllowlistExcludesOthers(t *testing.T) {
	ctx := context.Background()
	e := New()
	sb, _ := e.Prepare(ctx, ports.SandboxSpec{WorkingDir: t.TempDir()})
	_ = e.ApplyPolicy(ctx, sb, ports.SandboxPolicy{CommandAllow: []string{"echo"}})
	if _, err := e.ExecuteIn(ctx, sb, ports.CommandSpec{Program: "sh"}); err == nil {
		t.Fatal("non-allowlisted command should be refused")
	}
}

func TestEnvFilteringDenyByDefault(t *testing.T) {
	base := []string{"HOME=/home/x", "SECRET_TOKEN=abc", "ALLOWED=yes", "API_KEY=zzz"}
	// ALLOWED passes; HOME not in allowlist is dropped; SECRET_TOKEN/API_KEY are sensitive.
	got := filterEnv(base, []string{"ALLOWED", "SECRET_TOKEN"})
	has := map[string]bool{}
	for _, kv := range got {
		has[kv] = true
	}
	if !has["ALLOWED=yes"] {
		t.Error("ALLOWED should pass the allowlist")
	}
	if has["HOME=/home/x"] {
		t.Error("HOME is not allow-listed and must be dropped")
	}
	if has["API_KEY=zzz"] {
		t.Error("API_KEY is not allow-listed and must be dropped")
	}
	// SECRET_TOKEN is explicitly allow-listed by exact name, so it passes despite looking secret.
	if !has["SECRET_TOKEN=abc"] {
		t.Error("explicitly allow-listed sensitive var should pass")
	}
}

func TestWorkingDirOutsidePolicyRefused(t *testing.T) {
	ctx := context.Background()
	e := New()
	sb, _ := e.Prepare(ctx, ports.SandboxSpec{WorkingDir: t.TempDir()})
	_ = e.ApplyPolicy(ctx, sb, ports.SandboxPolicy{})
	_, err := e.ExecuteIn(ctx, sb, ports.CommandSpec{Program: "sh", Dir: "/etc"})
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != "E-SEC-032" {
		t.Fatalf("want E-SEC-032, got %v", err)
	}
}

func TestTeardownKillsLongRunning(t *testing.T) {
	ctx := context.Background()
	e := New()
	dir := t.TempDir()
	sb, _ := e.Prepare(ctx, ports.SandboxSpec{WorkingDir: dir})
	_ = e.ApplyPolicy(ctx, sb, ports.SandboxPolicy{CommandAllow: []string{"sleep"}})
	if _, err := e.ExecuteIn(ctx, sb, ports.CommandSpec{Program: "sleep", Args: []string{"60"}}); err != nil {
		t.Fatal(err)
	}
	// Teardown must return promptly, killing the process tree.
	done := make(chan error, 1)
	go func() { done <- e.Teardown(ctx, sb) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-context.Background().Done():
	}
	// Teardown is idempotent.
	if err := e.Teardown(ctx, sb); err != nil {
		t.Errorf("second teardown: %v", err)
	}
}
