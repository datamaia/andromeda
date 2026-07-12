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

	toml "github.com/pelletier/go-toml/v2"

	"github.com/datamaia/andromeda/internal/core"
)

// ManifestFile is the fixed manifest filename within a skill directory.
const ManifestFile = "skill.toml"

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
