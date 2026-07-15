// Package permstore persists the workspace command allow/deny policy under the .andromeda marker
// directory (permissions.toml). The interactive /permission menu edits it and the agent runtime
// merges it with andromeda.toml's [permission] section, so vetted commands the agent may run without
// a prompt — and commands it must always refuse — are visible and version-controllable in the
// workspace. It is layer L3 (a small file-backed engine, like memnote).
package permstore

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/datamaia/andromeda/internal/storage"
)

// FileName is the permissions file within the .andromeda marker directory.
const FileName = "permissions.toml"

// Lists to mutate.
const (
	Allow = "allow"
	Deny  = "deny"
)

// Rules is the workspace-managed command policy: argv-prefix entries the agent may run without a
// prompt (Allow) and entries it must always refuse (Deny). It mirrors andromeda.toml's [permission].
type Rules struct {
	Allow []string `toml:"allow"`
	Deny  []string `toml:"deny"`
}

// file wraps Rules under a [permission] table so the on-disk layout reads identically to the
// corresponding andromeda.toml section.
type file struct {
	Permission Rules `toml:"permission"`
}

// Path returns the permissions file path under the workspace marker directory.
func Path(root string) string { return filepath.Join(root, storage.MarkerDir, FileName) }

// Load reads the workspace permission rules, returning empty rules when the file is absent.
func Load(root string) (Rules, error) {
	data, err := os.ReadFile(Path(root)) //nolint:gosec // fixed path under the workspace marker dir
	if err != nil {
		if os.IsNotExist(err) {
			return Rules{}, nil
		}
		return Rules{}, err
	}
	var f file
	if err := toml.Unmarshal(data, &f); err != nil {
		return Rules{}, err
	}
	f.Permission.Allow = normalize(f.Permission.Allow)
	f.Permission.Deny = normalize(f.Permission.Deny)
	return f.Permission, nil
}

// Save writes the rules atomically (tmp + rename), creating the marker directory as needed. Both
// lists are always emitted so the file reflects the full allow/deny policy at a glance.
func Save(root string, r Rules) error {
	r.Allow = normalize(r.Allow)
	r.Deny = normalize(r.Deny)
	if r.Allow == nil {
		r.Allow = []string{}
	}
	if r.Deny == nil {
		r.Deny = []string{}
	}
	dir := filepath.Join(root, storage.MarkerDir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	body, err := toml.Marshal(file{Permission: r})
	if err != nil {
		return err
	}
	var b bytes.Buffer
	b.WriteString("# Managed by `/permission`. Commands the agent may run without an approval prompt\n")
	b.WriteString("# (allow) or must always refuse (deny), matched by argv prefix. Deny wins over allow.\n")
	b.WriteString("# Merged with andromeda.toml's [permission] section at runtime.\n\n")
	b.Write(body)
	tmp := Path(root) + ".tmp"
	if err := os.WriteFile(tmp, b.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, Path(root))
}

// Add inserts cmd into the named list (Allow or Deny), de-duplicated, and persists. It returns the
// updated rules. A blank command or unknown list is a no-op (rules are still returned).
func Add(root, list, cmd string) (Rules, error) {
	r, err := Load(root)
	if err != nil {
		return r, err
	}
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return r, nil
	}
	switch list {
	case Allow:
		r.Allow = addUnique(r.Allow, cmd)
	case Deny:
		r.Deny = addUnique(r.Deny, cmd)
	default:
		return r, nil
	}
	return r, Save(root, r)
}

// Remove deletes cmd from the named list (matched exactly, after trimming) and persists.
func Remove(root, list, cmd string) (Rules, error) {
	r, err := Load(root)
	if err != nil {
		return r, err
	}
	cmd = strings.TrimSpace(cmd)
	switch list {
	case Allow:
		r.Allow = without(r.Allow, cmd)
	case Deny:
		r.Deny = without(r.Deny, cmd)
	default:
		return r, nil
	}
	return r, Save(root, r)
}

// normalize trims entries, drops blanks, de-duplicates, and sorts for a stable file. Order does not
// affect matching semantics (deny is always evaluated before allow at the allowlist layer).
func normalize(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if x = strings.TrimSpace(x); x != "" && !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}

func addUnique(xs []string, x string) []string {
	for _, e := range xs {
		if e == x {
			return xs
		}
	}
	return append(xs, x)
}

func without(xs []string, x string) []string {
	out := xs[:0:0]
	for _, e := range xs {
		if e != x {
			out = append(out, e)
		}
	}
	return out
}
