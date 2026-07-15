package app

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AgentsFileName is the project agent-instructions file Andromeda reads on every run and folds into
// the agent's system context (the AGENTS.md convention shared with compatible tools). It is created
// by `/init`; see scaffoldProject in the CLI driver.
const AgentsFileName = "AGENTS.md"

// maxInstructionsBytes caps how much of AGENTS.md is injected so an oversized file cannot crowd out
// the conversation. The head of the file is kept.
const maxInstructionsBytes = 32 * 1024

// projectInstructions reads <root>/AGENTS.md and returns its content (size-capped and trimmed), or
// "" when the file is absent or unreadable. Reading it here — in the composition layer — keeps the
// agent engine pure (it receives the assembled system prompt, not a filesystem dependency).
func projectInstructions(root string) string {
	f, err := os.Open(filepath.Join(root, AgentsFileName)) //nolint:gosec // reads the workspace's own AGENTS.md
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, maxInstructionsBytes))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// composeSystem folds project AGENTS.md instructions into the base system prompt: the base identity
// stays first, and the project guidance follows under a labeled section so the model knows where the
// instructions came from. An empty base or empty instructions degrades gracefully.
func composeSystem(base, instructions string) string {
	if instructions == "" {
		return base
	}
	block := "Project instructions from " + AgentsFileName + " (apply throughout this session):\n\n" + instructions
	if strings.TrimSpace(base) == "" {
		return block
	}
	return base + "\n\n" + block
}

// projectMemory reads the workspace memory index (<root>/.andromeda/memory/MEMORY.md), size-capped
// and trimmed, or "" when absent. This is the human/agent-readable index of the file-based memory
// notes (see internal/memnote), folded into the system prompt so the model can recall durable facts.
func projectMemory(root string) string {
	f, err := os.Open(filepath.Join(root, ".andromeda", "memory", "MEMORY.md")) //nolint:gosec // reads the workspace's own memory index
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, maxInstructionsBytes))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// withMemory appends the workspace memory index to the system prompt under a labeled section, so the
// agent knows what has been durably remembered (and can open the referenced notes). Empty degrades.
func withMemory(base, memory string) string {
	if memory == "" {
		return base
	}
	block := "Workspace memory index from .andromeda/memory/MEMORY.md " +
		"(durable facts remembered for this project; open a referenced note when it is relevant):\n\n" + memory
	if strings.TrimSpace(base) == "" {
		return block
	}
	return base + "\n\n" + block
}

// projectMaps reports which workspace maps exist under .andromeda/ — the deterministic ontology
// (project.ttl) and the visual graph (index.md + graph.json), built by `/ontology` and `/graph`.
// Surfacing their presence lets the agent orient via a precomputed map instead of exploring blindly.
func projectMaps(root string) string {
	var have []string
	if fileExists(filepath.Join(root, ".andromeda", "ontology", "project.ttl")) {
		have = append(have, "- .andromeda/ontology/project.ttl — a deterministic Turtle map of how files and directories relate")
	}
	if fileExists(filepath.Join(root, ".andromeda", "graph", "index.md")) {
		have = append(have, "- .andromeda/graph/ (index.md, graph.json) — a node/edge map of the workspace with human-readable notes")
	}
	return strings.Join(have, "\n")
}

// withMaps tells the agent that precomputed workspace maps exist and to consult them first, so it can
// understand and navigate the repository quickly. Empty degrades gracefully.
func withMaps(base, maps string) string {
	if maps == "" {
		return base
	}
	block := "Workspace maps are available under .andromeda/ — read them to understand and navigate the " +
		"project quickly before exploring file by file:\n\n" + maps
	if strings.TrimSpace(base) == "" {
		return block
	}
	return base + "\n\n" + block
}

// fileExists reports whether path is an existing regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
