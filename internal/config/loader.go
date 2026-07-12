package config

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/datamaia/andromeda/internal/pal"
)

// ConfigFileName is the fixed configuration filename (Volume 0 product identity).
const ConfigFileName = "andromeda.toml"

// Load assembles a Manager from the standard layers, lowest to highest precedence:
// built-in defaults, the global config file (<ConfigHome>/andromeda.toml), the workspace file
// (<root>/.andromeda/andromeda.toml), the project file (<root>/andromeda.toml), and the
// process environment. Invocation overrides are added by the caller via SetOverrides. Missing
// files are skipped; a present-but-malformed file returns its E-CFG error.
func Load(ctx context.Context, dirs pal.ConfigDirs, workspaceRoot string) (*Manager, error) {
	m := New()
	m.SetDefaults(Defaults())

	configHome, err := dirs.ConfigHome()
	if err == nil {
		if err := loadFileInto(m, SourceGlobal, filepath.Join(configHome, ConfigFileName)); err != nil {
			return nil, err
		}
	}
	if workspaceRoot != "" {
		if err := loadFileInto(m, SourceWorkspace, filepath.Join(workspaceRoot, ".andromeda", ConfigFileName)); err != nil {
			return nil, err
		}
		if err := loadFileInto(m, SourceProject, filepath.Join(workspaceRoot, ConfigFileName)); err != nil {
			return nil, err
		}
	}
	m.SetEnv(os.Environ())
	return m, nil
}

func loadFileInto(m *Manager, source, path string) error {
	data, err := os.ReadFile(path) //nolint:gosec // path is derived from resolved config dirs
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	m.TrackFile(source, path)
	return m.LoadTOML(source, data)
}

// Defaults returns the built-in default configuration. It is intentionally small at this stage
// and grows as owning epics register their table defaults; owners contribute keys under their
// Volume 0 chapter 03 table ownership.
func Defaults() map[string]any {
	return map[string]any{
		"agent": map[string]any{
			"loop": map[string]any{"max_iterations": int64(50)},
		},
		"tui": map[string]any{
			"theme": map[string]any{"mode": "dark"},
		},
		"logging": map[string]any{
			"level":  "info",
			"format": "json",
		},
	}
}
