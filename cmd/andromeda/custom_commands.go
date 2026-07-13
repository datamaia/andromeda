package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/datamaia/andromeda/internal/tui"
)

// customCommandDirs are the workspace-relative directories scanned for user-authored slash commands,
// in precedence order (first occurrence of a name wins). Each command is a Markdown file whose name
// (without .md) is the command name; an optional YAML front matter supplies a description. Commands
// are agent capabilities, so `.agents/commands` is the home; `.claude/commands` is kept for compat.
var customCommandDirs = []string{
	".agents/commands",
	".claude/commands",
}

// discoverCustomCommands loads user-defined slash commands from the workspace. Templates support
// $ARGUMENTS and $1..$9 substitution (expanded in the TUI when the command runs). Discovery is
// best-effort: unreadable files and directories are skipped silently.
func discoverCustomCommands(wd string) []tui.CustomCommand {
	seen := map[string]bool{}
	var out []tui.CustomCommand
	for _, d := range customCommandDirs {
		ents, err := os.ReadDir(filepath.Join(wd, d))
		if err != nil {
			continue
		}
		for _, e := range ents {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".md")
			if name == "" || seen[name] {
				continue
			}
			data, err := os.ReadFile(filepath.Join(wd, d, e.Name())) //nolint:gosec // reads a discovered command file under the workspace
			if err != nil {
				continue
			}
			desc, body := parseFrontMatter(string(data))
			if strings.TrimSpace(desc) == "" {
				desc = "custom command"
			}
			seen[name] = true
			out = append(out, tui.CustomCommand{Name: name, Desc: desc, Template: strings.TrimSpace(body)})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// parseFrontMatter splits an optional leading "---\n…\n---" YAML front matter from a Markdown body,
// returning the value of a `description:` key (if present) and the remaining body.
func parseFrontMatter(s string) (desc, body string) {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return "", s
	}
	rest := strings.TrimPrefix(s, "---\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", s
	}
	header := rest[:end]
	body = strings.TrimLeft(rest[end+len("\n---"):], "\n")
	for _, line := range strings.Split(header, "\n") {
		if k, v, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(k) == "description" {
			desc = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return desc, body
}
