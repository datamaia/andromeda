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
