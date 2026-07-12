package config

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

// Source labels, lowest to highest precedence (FR-CFG-001).
const (
	SourceDefaults  = "defaults"
	SourceGlobal    = "global"
	SourceProfile   = "profile"
	SourceWorkspace = "workspace"
	SourceProject   = "project"
	SourceRuntime   = "runtime-override"
	SourceEnv       = "env"
	SourceCLI       = "cli-flag"
)

// precedence orders sources from lowest to highest.
var precedence = []string{
	SourceDefaults, SourceGlobal, SourceProfile, SourceWorkspace,
	SourceProject, SourceRuntime, SourceEnv, SourceCLI,
}

// EnvPrefix is the environment-variable prefix that maps into configuration.
const EnvPrefix = "ANDROMEDA_"

// layer is one contributing configuration layer: a flat dotted-key map with a source label.
type layer struct {
	source string
	values map[string]any
}

// Manager implements ports.ConfigPort by merging registered layers with precedence.
type Manager struct {
	layers  []layer
	byLabel map[string]int
	files   map[string]string // source label -> file path (for watching)
}

var _ ports.ConfigPort = (*Manager)(nil)

// New returns an empty Manager. Layers are added with SetDefaults, LoadTOML, SetEnv, etc.
func New() *Manager {
	return &Manager{byLabel: map[string]int{}, files: map[string]string{}}
}

// TrackFile records the on-disk path backing a source label so Watch can observe it.
func (m *Manager) TrackFile(source, path string) { m.files[source] = path }

// SetLayer replaces (or adds) a layer's flat values under a source label.
func (m *Manager) SetLayer(source string, values map[string]any) {
	if i, ok := m.byLabel[source]; ok {
		m.layers[i].values = values
		return
	}
	m.byLabel[source] = len(m.layers)
	m.layers = append(m.layers, layer{source: source, values: values})
}

// SetDefaults registers the defaults layer from a nested or flat map.
func (m *Manager) SetDefaults(v map[string]any) { m.SetLayer(SourceDefaults, flatten(v)) }

// LoadTOML parses TOML bytes and registers them under the given source label. Malformed TOML
// is reported as an E-CFG validation error.
func (m *Manager) LoadTOML(source string, data []byte) error {
	var nested map[string]any
	if err := toml.Unmarshal(data, &nested); err != nil {
		return &ports.PortError{
			Code:     "E-CFG-001",
			Category: "configuration",
			Severity: "error",
			Message:  "configuration file is not valid TOML",
			Detail:   err.Error(),
		}
	}
	m.SetLayer(source, flatten(nested))
	return nil
}

// SetEnv registers configuration derived from environment variables (ANDROMEDA_* → dotted
// keys). Mapping rule (FR-CFG-004): after stripping the prefix, a double underscore "__"
// separates configuration-table levels and a single underscore "_" is literal within a key
// segment. When the name contains no "__", single underscores are treated as separators, so
// the simple case still works: ANDROMEDA_TUI_THEME_MODE → tui.theme.mode, while
// ANDROMEDA_AGENT__MAX_ITERATIONS → agent.max_iterations.
func (m *Manager) SetEnv(environ []string) {
	values := map[string]any{}
	for _, kv := range environ {
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			continue
		}
		name, val := kv[:eq], kv[eq+1:]
		if !strings.HasPrefix(name, EnvPrefix) {
			continue
		}
		rest := strings.TrimPrefix(name, EnvPrefix)
		var key string
		if strings.Contains(rest, "__") {
			segs := strings.Split(rest, "__")
			for i := range segs {
				segs[i] = strings.ToLower(segs[i])
			}
			key = strings.Join(segs, ".")
		} else {
			key = strings.ToLower(strings.ReplaceAll(rest, "_", "."))
		}
		values[key] = val
	}
	m.SetLayer(SourceEnv, values)
}

// SetOverrides registers invocation-level overrides (flag or runtime) under the given source.
func (m *Manager) SetOverrides(source string, values map[string]any) {
	m.SetLayer(source, values)
}

// Resolve merges all registered layers by precedence and returns the effective values with
// per-key source attribution.
func (m *Manager) Resolve(ctx context.Context, _ ports.ConfigQuery) (ports.ResolvedConfig, error) {
	if err := ctx.Err(); err != nil {
		return ports.ResolvedConfig{}, err
	}
	out := ports.ResolvedConfig{
		Values:  map[string]any{},
		Sources: map[string]string{},
	}
	for _, src := range precedence {
		i, ok := m.byLabel[src]
		if !ok {
			continue
		}
		for k, v := range m.layers[i].values {
			out.Values[k] = v
			out.Sources[k] = src
		}
	}
	return out, nil
}

// Validate parses and checks a configuration document without applying it, returning all
// findings. At this stage it verifies TOML syntax; typed schema validation (ADR-024) is added
// by later epics.
func (m *Manager) Validate(ctx context.Context, doc ports.ConfigDocument) (ports.ValidationReport, error) {
	if err := ctx.Err(); err != nil {
		return ports.ValidationReport{}, err
	}
	if doc.Format != "" && doc.Format != "toml" {
		return ports.ValidationReport{
			Valid: false,
			Findings: []ports.ConfigFinding{{
				Message:  fmt.Sprintf("unsupported configuration format %q", doc.Format),
				Code:     "E-CFG-002",
				Severity: "error",
			}},
		}, nil
	}
	var nested map[string]any
	if err := toml.Unmarshal(doc.Raw, &nested); err != nil {
		return ports.ValidationReport{
			Valid: false,
			Findings: []ports.ConfigFinding{{
				Message:  "invalid TOML syntax",
				Detail:   err.Error(),
				Code:     "E-CFG-001",
				Severity: "error",
			}},
		}, nil
	}
	return ports.ValidationReport{Valid: true}, nil
}

// WatchInterval is how often the file watcher polls for changes.
var WatchInterval = 500 * time.Millisecond

// Watch returns a change stream for a selector. It polls the tracked configuration files'
// modification times; when one changes it reloads that layer, re-resolves, and emits a
// ConfigChange for each affected key that matches the selector. The stream ends when closed or
// when the context is cancelled.
func (m *Manager) Watch(ctx context.Context, sel ports.ConfigSelector) (ports.Stream[ports.ConfigChange], error) {
	if len(m.files) == 0 {
		return streams.Slice([]ports.ConfigChange{}), nil
	}
	st, send, closeFn := streams.Chan[ports.ConfigChange](64)
	prev, _ := m.Resolve(ctx, ports.ConfigQuery{})
	mtimes := map[string]time.Time{}
	for _, path := range m.files {
		if fi, err := os.Stat(path); err == nil {
			mtimes[path] = fi.ModTime()
		}
	}
	go func() {
		defer closeFn()
		ticker := time.NewTicker(WatchInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			changed := false
			for source, path := range m.files {
				fi, err := os.Stat(path)
				if err != nil {
					continue
				}
				if !fi.ModTime().After(mtimes[path]) {
					continue
				}
				mtimes[path] = fi.ModTime()
				if data, rerr := os.ReadFile(path); rerr == nil { //nolint:gosec // tracked config path
					_ = m.LoadTOML(source, data)
					changed = true
				}
			}
			if !changed {
				continue
			}
			next, _ := m.Resolve(ctx, ports.ConfigQuery{})
			for k, nv := range next.Values {
				if !selectorMatches(sel, k) {
					continue
				}
				if ov, ok := prev.Values[k]; !ok || ov != nv {
					if !send(ports.ConfigChange{Key: k, OldValue: prev.Values[k], NewValue: nv, Source: next.Sources[k]}) {
						return
					}
				}
			}
			prev = next
		}
	}()
	return st, nil
}

func selectorMatches(sel ports.ConfigSelector, key string) bool {
	if len(sel.Keys) == 0 && len(sel.Prefixes) == 0 {
		return true
	}
	for _, k := range sel.Keys {
		if k == key {
			return true
		}
	}
	for _, p := range sel.Prefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// Keys returns the sorted set of keys present across all layers (diagnostic helper).
func (m *Manager) Keys() []string {
	seen := map[string]struct{}{}
	for _, l := range m.layers {
		for k := range l.values {
			seen[k] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
