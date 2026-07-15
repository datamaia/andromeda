// Package skill is layer L3: the Skill System (Volume 6, FR-SKILL-001). A skill is a packaged,
// versioned unit of procedural knowledge — a manifest plus a prompt — that declares the tools,
// capabilities, and providers it needs. The loader parses and validates a skill directory; the
// engine resolves a skill against an available tool set and capability set, composing its
// system prompt and reporting unmet requirements precisely (no silent degradation).
package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/datamaia/andromeda/internal/core"
)

// ManifestFile is the fixed TOML manifest filename within a skill directory (the spec format,
// FR-SKILL-001). Markdown skills (SKILL.md) are also accepted — see LoadDir.
const ManifestFile = "skill.toml"

// markdownManifests are the accepted Markdown skill manifests, in priority order. A Markdown skill
// carries its metadata in YAML-style frontmatter (name, description, optional version) and its
// instructions in the body — the SKILL.md convention shared across the agent-tooling ecosystem
// (Claude Code, Codex, …). This lets a skill authored for those tools be recognized as-is.
var markdownManifests = []string{"SKILL.md", "skill.md"}

// SkillDirs are the conventional workspace base directories that may hold a skills/ subtree, in
// discovery-precedence order. Andromeda's own .agents is authoritative; .claude/.codex/.agent are
// recognized so skills written for compatible tools are found without being moved.
var SkillDirs = []string{".agents", ".claude", ".codex", ".agent"}

// Manifest is a skill's declared metadata.
type Manifest struct {
	Name                 string   `toml:"name"`
	Version              string   `toml:"version"`
	Description          string   `toml:"description"`
	Prompt               string   `toml:"prompt"` // relative path to the prompt file
	RequiredTools        []string `toml:"required_tools"`
	RequiredCapabilities []string `toml:"required_capabilities"`
	CompatibleProviders  []string `toml:"compatible_providers"`
}

// Skill is a loaded skill: its manifest plus the prompt content.
type Skill struct {
	Manifest Manifest
	Prompt   string
	Dir      string
}

// Load reads and validates a skill from a directory.
func Load(dir string) (*Skill, error) {
	data, err := os.ReadFile(filepath.Join(dir, ManifestFile)) //nolint:gosec // dir is an operator-provided skill path
	if err != nil {
		return nil, skErr("E-SKILL-001", "cannot read skill manifest: "+err.Error())
	}
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, skErr("E-SKILL-002", "invalid skill manifest: "+err.Error())
	}
	if m.Name == "" || m.Version == "" {
		return nil, skErr("E-SKILL-003", "skill manifest requires name and version")
	}
	s := &Skill{Manifest: m, Dir: dir}
	if m.Prompt != "" {
		pb, err := os.ReadFile(filepath.Join(dir, m.Prompt)) //nolint:gosec // prompt path is within the skill dir
		if err != nil {
			return nil, skErr("E-SKILL-004", "cannot read skill prompt: "+err.Error())
		}
		s.Prompt = string(pb)
	}
	return s, nil
}

// LoadDir loads a skill from a directory, accepting either a Markdown manifest (SKILL.md — the
// ecosystem-standard format, tried first) or the TOML manifest (skill.toml). This is the loader the
// TUI and discovery use; Load remains the strict spec (skill.toml only) entry point.
func LoadDir(dir string) (*Skill, error) {
	for _, name := range markdownManifests {
		data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // name is a fixed manifest filename under an operator-provided dir
		if err == nil {
			return loadMarkdown(dir, data)
		}
	}
	return Load(dir)
}

// ManifestPath returns the manifest file backing a skill directory (the Markdown manifest if one
// exists, else skill.toml), or the directory itself when none is found. Used to show the user where
// a discovered skill lives.
func ManifestPath(dir string) string {
	for _, name := range append(append([]string{}, markdownManifests...), ManifestFile) {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return filepath.Join(dir, name)
		}
	}
	return dir
}

var mdFrontmatter = regexp.MustCompile(`(?s)^---\r?\n(.*?)\r?\n---\r?\n?(.*)$`)

// loadMarkdown parses a SKILL.md-style skill: optional YAML-ish frontmatter (name, description,
// version, and the same list fields as the TOML manifest) followed by the instruction body, which
// becomes the prompt (unless the frontmatter names a separate prompt file). It is deliberately
// lenient — a Markdown skill with no name falls back to its directory name — so a reasonably-formed
// skill is always recognized rather than silently ignored.
func loadMarkdown(dir string, data []byte) (*Skill, error) {
	var m Manifest
	body := string(data)
	if fm := mdFrontmatter.FindSubmatch(data); fm != nil {
		for _, line := range strings.Split(string(fm[1]), "\n") {
			key, val, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			val = strings.Trim(strings.TrimSpace(val), `"'`)
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "name":
				m.Name = val
			case "description":
				m.Description = val
			case "version":
				m.Version = val
			case "prompt":
				m.Prompt = val
			case "required_tools", "tools":
				m.RequiredTools = splitList(val)
			case "required_capabilities", "capabilities":
				m.RequiredCapabilities = splitList(val)
			case "compatible_providers", "providers":
				m.CompatibleProviders = splitList(val)
			}
		}
		body = string(fm[2])
	}
	if m.Name == "" {
		m.Name = filepath.Base(dir)
	}
	s := &Skill{Manifest: m, Dir: dir}
	if m.Prompt != "" {
		pb, err := os.ReadFile(filepath.Join(dir, m.Prompt)) //nolint:gosec // prompt path is within the skill dir
		if err != nil {
			return nil, skErr("E-SKILL-004", "cannot read skill prompt: "+err.Error())
		}
		s.Prompt = string(pb)
	} else {
		s.Prompt = strings.TrimSpace(body)
	}
	if s.Manifest.Description == "" {
		s.Manifest.Description = firstLine(s.Prompt)
	}
	return s, nil
}

// splitList parses a frontmatter list value written either as a comma-separated string or a YAML
// inline array ("[a, b]").
func splitList(v string) []string {
	v = strings.Trim(v, "[]")
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.Trim(strings.TrimSpace(p), `"'`); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// firstLine returns the first non-empty, de-marked line of text — used to synthesize a description
// from a Markdown skill's body when its frontmatter omits one.
func firstLine(s string) string {
	for _, ln := range strings.Split(s, "\n") {
		if ln = strings.TrimSpace(strings.TrimLeft(ln, "#> ")); ln != "" {
			return ln
		}
	}
	return ""
}

// Discovered is a skill located during workspace discovery, tagged with the base directory (source)
// it was found under and the manifest file that defined it.
type Discovered struct {
	*Skill
	Source string // the base dir it was found under, e.g. ".claude"
	Path   string // the manifest file backing it
}

// Discover scans each conventional base directory (SkillDirs) under root for a skills/ subtree and
// loads every skill directory it finds, accepting both SKILL.md and skill.toml. Skills are
// de-duplicated by name — the first source in SkillDirs order wins — so a workspace can layer
// tool-specific skill folders without duplicates. Unreadable or invalid skill dirs are skipped.
func Discover(root string) []Discovered {
	var out []Discovered
	seen := map[string]bool{}
	for _, base := range SkillDirs {
		skillsDir := filepath.Join(root, base, "skills")
		ents, err := os.ReadDir(skillsDir)
		if err != nil {
			continue
		}
		for _, e := range ents {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(skillsDir, e.Name())
			sk, err := LoadDir(dir)
			if err != nil || seen[sk.Manifest.Name] {
				continue
			}
			seen[sk.Manifest.Name] = true
			out = append(out, Discovered{Skill: sk, Source: base, Path: ManifestPath(dir)})
		}
	}
	return out
}

// Resolution is the outcome of resolving a skill against the environment.
type Resolution struct {
	OK           bool
	SystemPrompt string
	MissingTools []string
	MissingCaps  []string
}

// Resolve checks the skill's requirements against the available tools and capabilities and, when
// satisfied, returns the composed system prompt. Missing requirements are reported precisely.
func Resolve(s *Skill, availableTools []string, availableCaps core.Capabilities) Resolution {
	toolSet := toSet(availableTools)
	var res Resolution
	for _, t := range s.Manifest.RequiredTools {
		if !toolSet[t] {
			res.MissingTools = append(res.MissingTools, t)
		}
	}
	for _, c := range s.Manifest.RequiredCapabilities {
		if !availableCaps.Has(core.Capability(c)) {
			res.MissingCaps = append(res.MissingCaps, c)
		}
	}
	res.OK = len(res.MissingTools) == 0 && len(res.MissingCaps) == 0
	if res.OK {
		res.SystemPrompt = s.Prompt
	}
	return res
}

func toSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func skErr(code, msg string) error {
	return fmt.Errorf("%s: %s", code, msg)
}
