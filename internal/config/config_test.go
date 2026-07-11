package config

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

func TestPrecedenceAndSourceAttribution(t *testing.T) {
	ctx := context.Background()
	m := New()
	m.SetDefaults(map[string]any{
		"tui":   map[string]any{"theme": map[string]any{"mode": "dark"}},
		"agent": map[string]any{"max_iterations": int64(50)},
	})
	if err := m.LoadTOML(SourceGlobal, []byte("[agent]\nmax_iterations = 20\n")); err != nil {
		t.Fatal(err)
	}
	if err := m.LoadTOML(SourceWorkspace, []byte("[tui.theme]\nmode = \"light\"\n")); err != nil {
		t.Fatal(err)
	}
	m.SetEnv([]string{"ANDROMEDA_AGENT__MAX_ITERATIONS=99", "PATH=/usr/bin"})

	res, err := m.Resolve(ctx, ports.ConfigQuery{Scope: "workspace"})
	if err != nil {
		t.Fatal(err)
	}
	// env (highest here) wins for agent.max_iterations
	if res.Values["agent.max_iterations"] != "99" {
		t.Errorf("agent.max_iterations = %v (source %s), want 99 from env",
			res.Values["agent.max_iterations"], res.Sources["agent.max_iterations"])
	}
	if res.Sources["agent.max_iterations"] != SourceEnv {
		t.Errorf("source = %s, want env", res.Sources["agent.max_iterations"])
	}
	// workspace overrides defaults for tui.theme.mode
	if res.Values["tui.theme.mode"] != "light" {
		t.Errorf("tui.theme.mode = %v, want light", res.Values["tui.theme.mode"])
	}
	if res.Sources["tui.theme.mode"] != SourceWorkspace {
		t.Errorf("source = %s, want workspace", res.Sources["tui.theme.mode"])
	}
}

func TestEnvMappingMirrorsExample(t *testing.T) {
	m := New()
	m.SetEnv([]string{"ANDROMEDA_TUI_THEME_MODE=light"})
	res, _ := m.Resolve(context.Background(), ports.ConfigQuery{})
	if res.Values["tui.theme.mode"] != "light" {
		t.Errorf("env mapping wrong: %v", res.Values)
	}
}

func TestLoadTOMLInvalidReportsE_CFG_001(t *testing.T) {
	m := New()
	err := m.LoadTOML(SourceGlobal, []byte("this is = = not toml"))
	if err == nil {
		t.Fatal("expected an error for malformed TOML")
	}
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != "E-CFG-001" {
		t.Fatalf("want E-CFG-001 PortError, got %v", err)
	}
}

func TestValidate(t *testing.T) {
	m := New()
	ctx := context.Background()
	ok, _ := m.Validate(ctx, ports.ConfigDocument{Format: "toml", Raw: []byte("[a]\nb = 1\n")})
	if !ok.Valid {
		t.Error("expected valid")
	}
	bad, _ := m.Validate(ctx, ports.ConfigDocument{Format: "toml", Raw: []byte("x = = 1")})
	if bad.Valid || len(bad.Findings) == 0 {
		t.Error("expected invalid with findings")
	}
	unsup, _ := m.Validate(ctx, ports.ConfigDocument{Format: "yaml"})
	if unsup.Valid || unsup.Findings[0].Code != "E-CFG-002" {
		t.Error("expected E-CFG-002 for unsupported format")
	}
}

func TestFlattenUnflattenRoundTrip(t *testing.T) {
	nested := map[string]any{
		"a": map[string]any{"b": map[string]any{"c": "v"}},
		"x": int64(1),
	}
	flat := flatten(nested)
	if flat["a.b.c"] != "v" || flat["x"] != int64(1) {
		t.Fatalf("flatten wrong: %v", flat)
	}
	back := unflatten(flat)
	ab, _ := back["a"].(map[string]any)
	bc, _ := ab["b"].(map[string]any)
	if bc["c"] != "v" {
		t.Fatalf("unflatten wrong: %v", back)
	}
	if s, ok := asString(int64(5)); !ok || s != "5" {
		t.Errorf("asString(int64) = %q,%v", s, ok)
	}
}

func TestWatchReturnsUsableStream(t *testing.T) {
	m := New()
	st, err := m.Watch(context.Background(), ports.ConfigSelector{})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.Next(context.Background()); err != ports.ErrEndOfStream {
		t.Errorf("want ErrEndOfStream, got %v", err)
	}
}
