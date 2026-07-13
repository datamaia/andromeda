package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/datamaia/andromeda/internal/app"
)

// scaffoldProject creates or completes Andromeda's project layout in wd and returns a human summary.
// It is idempotent and never overwrites existing files: it creates whatever is missing and augments
// an existing andromeda.toml with any config sections it lacks (allowlist / MCP / plugins).
//
// Layout:
//
//	AGENTS.md       — agent instructions, read into the system context on every run (app.AgentsFileName)
//	andromeda.toml  — project config: provider, command allowlist, MCP servers, plugins
//	.agents/        — skills, custom commands, and MCP definitions (committed, project-scoped)
//	.andromeda/     — persistent memory and generated-data surface (per-workspace customization)
func scaffoldProject(wd, provider, model string) string {
	var lines []string
	note := func(name, status string) { lines = append(lines, fmt.Sprintf("  %-16s %s", name, status)) }

	note(app.AgentsFileName, ensureFile(filepath.Join(wd, app.AgentsFileName), agentsTemplate(wd)))
	note("andromeda.toml", ensureTOML(filepath.Join(wd, "andromeda.toml"), provider, model))

	// .agents/ — the home for agent capabilities.
	agentsStatus := ensureDir(filepath.Join(wd, ".agents"))
	for _, sub := range []string{"skills", "commands", "mcp"} {
		_ = ensureDir(filepath.Join(wd, ".agents", sub))
	}
	_ = ensureFile(filepath.Join(wd, ".agents", "README.md"), agentsDirReadme)
	note(".agents/", agentsStatus+" (skills · commands · mcp)")

	// .andromeda/ — the per-workspace memory and data surface.
	andStatus := ensureDir(filepath.Join(wd, ".andromeda"))
	_ = ensureDir(filepath.Join(wd, ".andromeda", "memory"))
	_ = ensureFile(filepath.Join(wd, ".andromeda", "README.md"), andromedaDirReadme)
	note(".andromeda/", andStatus+" (memory)")

	return "init · project scaffold\n" + strings.Join(lines, "\n") +
		"\nedit " + app.AgentsFileName + " and andromeda.toml — Andromeda folds " + app.AgentsFileName +
		" into its context on every run."
}

// ensureFile creates path with content if it does not exist, and reports what happened. Existing
// files are left untouched. Files are written 0600 (git tracks content, not local perms).
func ensureFile(path, content string) string {
	if _, err := os.Stat(path); err == nil {
		return "kept"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return "error: " + err.Error()
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "error: " + err.Error()
	}
	return "created"
}

// ensureDir creates a directory (and parents) if absent, reporting what happened.
func ensureDir(path string) string {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return "kept"
	}
	if err := os.MkdirAll(path, 0o750); err != nil {
		return "error: " + err.Error()
	}
	return "created"
}

// ensureTOML writes a fresh andromeda.toml (full scaffold) when absent, or augments an existing one
// with any of the config sections it is missing (allowlist / MCP / plugins), never clobbering values
// the user already wrote.
func ensureTOML(path, provider, model string) string {
	data, err := os.ReadFile(path) //nolint:gosec // reads the workspace's own andromeda.toml to augment it
	if os.IsNotExist(err) {
		full := tomlHeader + tomlProvider(provider, model) + tomlAgent + tomlPermission + tomlMCP + tomlPlugins + tomlLogging + tomlTUI
		if werr := os.WriteFile(path, []byte(full), 0o600); werr != nil {
			return "error: " + werr.Error()
		}
		return fmt.Sprintf("created (provider %s · model %s)", provider, model)
	}
	if err != nil {
		return "error: " + err.Error()
	}
	existing := string(data)
	var added []string
	var b strings.Builder
	b.WriteString(existing)
	if !strings.HasSuffix(existing, "\n") {
		b.WriteString("\n")
	}
	for _, sec := range []struct{ header, block string }{
		{"[permission]", tomlPermission},
		{"[mcp]", tomlMCP},
		{"[plugins]", tomlPlugins},
	} {
		if !hasTOMLSection(existing, sec.header) {
			b.WriteString("\n" + sec.block)
			added = append(added, strings.Trim(sec.header, "[]"))
		}
	}
	if len(added) == 0 {
		return "kept (already complete)"
	}
	if werr := os.WriteFile(path, []byte(b.String()), 0o600); werr != nil {
		return "error: " + werr.Error()
	}
	return "updated (+" + strings.Join(added, ", ") + ")"
}

// hasTOMLSection reports whether content already declares the given top-level table header.
func hasTOMLSection(content, header string) bool {
	for _, ln := range strings.Split(content, "\n") {
		if strings.TrimSpace(ln) == header {
			return true
		}
	}
	return false
}

// agentsTemplate renders a starter AGENTS.md, pre-filling the project name and (when detectable) the
// build/test/lint commands for the workspace's stack.
func agentsTemplate(wd string) string {
	name := filepath.Base(wd)
	build, test, lint := detectCommands(wd)
	return fmt.Sprintf(`# AGENTS.md

Guidance for AI coding agents (Andromeda and compatible AGENTS.md tools) working in this repository.
Andromeda reads this file on every run and folds it into the agent's system context — keep it short,
current, and specific; it steers every turn.

## Project

%s — <one or two sentences: what this project is and its primary stack.>

## Setup & commands

- Build: %s
- Test:  %s
- Lint:  %s
- Run:   <how to run it locally>

## Conventions

<!-- Coding style, module layout, naming, commit-message rules, anything non-obvious. -->

## Guardrails

<!-- What the agent must NOT do: destructive commands, files never to touch, secrets, etc. -->
`, name, build, test, lint)
}

// detectCommands guesses build/test/lint commands from marker files, falling back to placeholders.
func detectCommands(wd string) (build, test, lint string) {
	switch {
	case fileExists(wd, "go.mod"):
		return "`go build ./...`", "`go test ./...`", "`gofmt -l . && go vet ./...`"
	case fileExists(wd, "package.json"):
		return "`npm run build`", "`npm test`", "`npm run lint`"
	case fileExists(wd, "Cargo.toml"):
		return "`cargo build`", "`cargo test`", "`cargo clippy`"
	case fileExists(wd, "pyproject.toml"), fileExists(wd, "setup.py"):
		return "`python -m build`", "`pytest`", "`ruff check .`"
	default:
		return "<build command>", "<test command>", "<lint command>"
	}
}

func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

// --- templates ---

const tomlHeader = `# andromeda.toml — project configuration for Andromeda.
# Committed to the repo; values here apply to everyone working in this project.
# Precedence (low→high): defaults < global < .andromeda/andromeda.toml < this file < env < flags.

`

func tomlProvider(provider, model string) string {
	return fmt.Sprintf(`[provider]
# The provider and model this project prefers. Override with --provider/--model or /provider,/model.
default = %q
model   = %q

`, provider, model)
}

const tomlAgent = `[agent]
# Safety rail: the maximum number of tool-using iterations a single run may take.
max_iterations = 25

`

const tomlPermission = `[permission]
# Command allowlist for terminal_run: argv prefixes the agent may run WITHOUT prompting. Anything
# not matched here asks for approval; entries in ` + "`deny`" + ` are always refused.
allow = [
  # "git status",
  # "git diff",
  # "go build ./...",
  # "go test ./...",
]
deny = [
  # "git push --force",
  # "rm -rf",
]

`

const tomlMCP = `[mcp]
# Model Context Protocol servers connected at agent start; each exposes tools the agent can call.
# Definitions may also live as files under .agents/mcp/. Uncomment and edit to enable.
# [mcp.servers.filesystem]
# command = "npx"
# args    = ["-y", "@modelcontextprotocol/server-filesystem", "."]
#
# [mcp.servers.github]
# transport = "http"
# url       = "https://api.githubcopilot.com/mcp/"

`

const tomlPlugins = `[plugins]
# WASM/extension plugins loaded at start. Point at a built module under .agents/plugins/.
# enabled = []
# [plugins.example]
# path = ".agents/plugins/example.wasm"

`

const tomlLogging = `[logging]
# One of: debug, info, warn, error.
level = "info"

`

const tomlTUI = `[tui]
# Start-screen theme: "dark" (default), "light", or "auto".
theme.mode = "dark"
`

const agentsDirReadme = `# .agents/

Home for this project's agent capabilities, shared by Andromeda and compatible AGENTS.md tools.

- ` + "`skills/`" + `   — reusable skills; one per directory, described by a ` + "`skill.toml`" + ` manifest.
- ` + "`commands/`" + ` — custom slash commands as Markdown templates ($ARGUMENTS, $1..$9).
- ` + "`mcp/`" + `      — Model Context Protocol server definitions (also configurable in andromeda.toml [mcp]).

These are project-scoped and committed with the repo. Machine-local memory and generated data live
under .andromeda/ instead.
`

const andromedaDirReadme = `# .andromeda/

Andromeda's per-workspace surface: persistent memory and generated, customization-specific data that
should not live in the source tree proper — a durable "SSD" scratch surface.

- ` + "`memory/`" + `        — persistent conversational and project memory.
- ` + "`andromeda.toml`" + ` — optional workspace-level config overrides (lower precedence than the root andromeda.toml).

Reserved for upcoming features: dynamic knowledge graphs, ontologies, and a data→Markdown processor.
If any contents are machine-local, add them to .gitignore.
`
