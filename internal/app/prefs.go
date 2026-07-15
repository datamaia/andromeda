package app

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/datamaia/andromeda/internal/pal"
)

// Prefs are lightweight, global user preferences remembered across sessions — currently the last
// provider and model the user actually used, so a returning user is dropped straight back into their
// setup instead of re-onboarding every launch. Stored as a small JSON file in the app data dir.
type Prefs struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

func prefsPath() (string, error) {
	base, err := pal.NewConfigDirs().DataHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "prefs.json"), nil
}

// LoadPrefs reads the remembered preferences. A missing file yields a zero Prefs and no error, so
// first-run callers see empty values.
func LoadPrefs() (Prefs, error) {
	path, err := prefsPath()
	if err != nil {
		return Prefs{}, err
	}
	data, err := os.ReadFile(path) //nolint:gosec // fixed filename in the app data dir
	if err != nil {
		if os.IsNotExist(err) {
			return Prefs{}, nil
		}
		return Prefs{}, err
	}
	var p Prefs
	if err := json.Unmarshal(data, &p); err != nil {
		return Prefs{}, err
	}
	return p, nil
}

// SavePrefs persists the remembered preferences atomically (temp file + rename, owner-only perms).
// Callers treat it as best-effort.
func SavePrefs(p Prefs) error {
	path, err := prefsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
