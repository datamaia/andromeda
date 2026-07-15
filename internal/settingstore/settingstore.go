// Package settingstore persists small workspace-scoped UI settings under the .andromeda marker
// directory (settings.toml). These are toggles the interactive TUI reads and writes — kept separate
// from andromeda.toml (project configuration) and permissions.toml (command policy) so a UI
// preference never collides with project config. It is layer L3 (a small file-backed engine, like
// permstore and memnote).
package settingstore

import (
	"bytes"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/datamaia/andromeda/internal/storage"
)

// FileName is the settings file within the .andromeda marker directory.
const FileName = "settings.toml"

// Settings is the workspace-scoped TUI settings persisted between sessions. Fields are added as
// commands need them; a missing file (or field) yields the zero value, so defaults are the zero
// values.
type Settings struct {
	// AutoCompact summarizes the conversation automatically once it grows large, before the next
	// agent turn, to keep the context (and token cost) bounded. Toggled by /autocompact.
	AutoCompact bool `toml:"auto_compact"`
}

// file wraps Settings under a [settings] table for a self-describing on-disk layout.
type file struct {
	Settings Settings `toml:"settings"`
}

// Path returns the settings file path under the workspace marker directory.
func Path(root string) string { return filepath.Join(root, storage.MarkerDir, FileName) }

// Load reads the workspace settings, returning the zero value when the file is absent.
func Load(root string) (Settings, error) {
	data, err := os.ReadFile(Path(root)) //nolint:gosec // fixed path under the workspace marker dir
	if err != nil {
		if os.IsNotExist(err) {
			return Settings{}, nil
		}
		return Settings{}, err
	}
	var f file
	if err := toml.Unmarshal(data, &f); err != nil {
		return Settings{}, err
	}
	return f.Settings, nil
}

// Save writes the settings atomically (tmp + rename), creating the marker directory as needed.
func Save(root string, s Settings) error {
	dir := filepath.Join(root, storage.MarkerDir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	body, err := toml.Marshal(file{Settings: s})
	if err != nil {
		return err
	}
	var b bytes.Buffer
	b.WriteString("# Managed by Andromeda's interactive TUI (e.g. /autocompact). Workspace-scoped UI settings.\n\n")
	b.Write(body)
	tmp := Path(root) + ".tmp"
	if err := os.WriteFile(tmp, b.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, Path(root))
}
