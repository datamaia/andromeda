package app

import (
	"context"
	"testing"
)

func TestDoctorPassesInCleanEnvironment(t *testing.T) {
	// Hermetic: isolate HOME and XDG so the diagnostic touches only temp directories.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	ws := t.TempDir()
	rep, err := Doctor(context.Background(), ws)
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}
	if !rep.OK() {
		for _, c := range rep.Checks {
			t.Logf("[%v] %s: %s", c.OK, c.Name, c.Detail)
		}
		t.Fatal("expected all checks to pass")
	}
	names := map[string]bool{}
	for _, c := range rep.Checks {
		names[c.Name] = c.OK
	}
	for _, want := range []string{"config", "global-db", "workspace-db", "events"} {
		if !names[want] {
			t.Errorf("missing or failed check %q", want)
		}
	}
}

func TestDoctorIsRepeatable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	ws := t.TempDir()
	ctx := context.Background()
	// Running twice against the same workspace must remain healthy (idempotent migrations,
	// reopened databases, additional persisted event).
	if rep, _ := Doctor(ctx, ws); !rep.OK() {
		t.Fatal("first run failed")
	}
	rep, _ := Doctor(ctx, ws)
	if !rep.OK() {
		t.Fatal("second run failed")
	}
}
