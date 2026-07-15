# Changelog

All notable changes to Andromeda are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows
[Semantic Versioning](https://semver.org/spec/v2.0.0.html) per
ADR-015. Release entries are derived from Conventional
Commit history by the release automation (ADR-013) and committed at release time.

## [Unreleased]

Deep session and context commands.

### Added

- **Session branching.** `/branch` bookmarks the current conversation as a new saved session while
  you keep working where you are; `/clone` freezes the current line and continues on a fresh copy so
  the original is preserved. Each fork records its origin, and `/tree` shows the resulting lineage.
- **Session switching.** `/sessions` lists every saved session (turns, date, title, current marked);
  `/sessions resume <id>` swaps the live conversation in place and re-seeds the transcript;
  `/sessions rm <id>` deletes one. The session you are in is never removed out from under you.
- **`/btw` notes.** `/btw <note>` queues an out-of-band "by the way" that is folded into your next
  message to the agent — context without triggering a reply now. Multiple notes stack and clear once
  sent.
- **Real conversation compaction.** `/compact` now actually summarizes: it asks the provider to
  condense the conversation and replaces the agent's cross-turn context with the summary (encoded as
  a valid user→assistant pair, so strict-alternation providers keep working), trimming token cost
  while preserving decisions, changes, and open tasks. It runs off the UI thread.
- **`/autocompact`.** Toggle automatic compaction (persisted per workspace in
  `.andromeda/settings.toml`): once a conversation grows past a turn threshold, it is summarized
  before the next turn, with a visible notice so the trim is never silent.

## [0.1.11] - 2026-07-15

### Fixed

- **`/clear` and `/new` now truly reset context.** They previously cleared only the on-screen
  transcript while the agent kept earlier turns in its cross-turn memory, so it still "remembered"
  them. They now also drop the session's conversation history and start a fresh session (the previous
  conversation stays saved on disk under its own id) — clearing means the agent genuinely forgets.

## [0.1.10] - 2026-07-15

Workspace maps the agent (and you) can navigate faster.

### Added

- **The agent knows about workspace maps.** When the deterministic ontology
  (`.andromeda/ontology/project.ttl`) or the visual graph (`.andromeda/graph/`) exist, a pointer is
  folded into the agent's system prompt — alongside `AGENTS.md` and the memory index — so it orients
  via the precomputed map instead of exploring the tree blindly.
- **Obsidian-style graph previews.** Hovering a node in the `/graph` viewer now shows a short summary
  of the file it maps — a frontmatter `description` when present, otherwise the first meaningful line
  — so you can read the graph without opening each file. Summaries are computed at build time and
  HTML-escaped.

## [0.1.9] - 2026-07-15

Provider reach and command-parity groundwork.

### Added

- **AWS Kiro provider.** A new `kiro` provider reaches AWS Kiro's models through a local
  OpenAI-compatible gateway (kiro-gateway) using your Kiro / AWS Builder ID sign-in, with its model
  list (Claude Sonnet 4.5/4, Claude Haiku 4.5, GLM-5, DeepSeek-V3.2, MiniMax M2.5/M2.1,
  Qwen3-Coder-Next) selectable even before the gateway is running. Kiro exposes no official public
  API; this path is unofficial, and Amazon Bedrock is the sanctioned alternative for the same models.
- **Working-directory commands.** `/add-dir <path>` adds an extra directory whose files join
  `@`-mention completion; `/cd <path>` moves the session's working directory, updating the header and
  refreshing file/skill completion.
- **Command aliases.** `/new` (→ clear), `/models` (→ model), `/summarize` (→ compact).

### Changed

- **`/help` and `/commands` are now distinct.** `/help` is an orientation guide (modes, the `/ @ $`
  sigils, keybindings); `/commands` is the exhaustive reference with each command's aliases.
- **A "thinking" animation.** The header shows a calmer, distinct spinner while the model reasons
  between steps — separate from the busy "working" spin (a tool executing) and "responding"
  (streaming) — so a multi-step turn reads clearly.

## [0.1.8] - 2026-07-15

Skills, workflows, and command permissions.

### Added

- **Skills that just work.** Andromeda now recognizes `SKILL.md` skills (YAML frontmatter plus an
  instruction body — the convention shared with Claude Code and Codex) in addition to `skill.toml`,
  and discovers them across `.agents`, `.claude`, `.codex`, and `.agent` (de-duplicated by name, with
  the source shown). `/skills` lists everything it finds, and the create action scaffolds a
  recognizable `SKILL.md` so a skill you make is a skill Andromeda sees.
- **Inline skill invocation.** Type `$` in the composer to open a ranked menu of your skills;
  selecting one inserts `$name`. On send, each referenced skill's instructions are folded into the
  agent run and a line records the activation — the `$`-invocation convention Codex popularized.
- **Runnable workflows.** `/workflows` now lists step-by-step Markdown recipes discovered under
  `.agents/workflows`, `.andromeda/workflows`, `.windsurf/workflows`, and `.cursor/workflows`, and
  can **Run** one (send the recipe to the agent as a goal), create, or edit it — the workflow
  convention used by Cursor, Devin, and Windsurf.
- **Command permissions from the TUI.** A new `/permission` menu manages the command allow/deny
  lists — pre-approve vetted commands, or always refuse dangerous ones, matched by argv prefix.
  Rules are stored in `.andromeda/permissions.toml` (both lists visible in the workspace) and merged
  with `andromeda.toml`'s `[permission]` section at runtime, so changes take effect immediately.
  Text form: `/permission allow <cmd>`, `deny <cmd>`, `rm allow|deny <cmd>`, `list`.
- **Wider custom-command discovery.** User-authored slash commands are now also read from
  `.andromeda/commands` and `.codex/commands`, alongside `.agents/commands` and `.claude/commands`.

## [0.1.7] - 2026-07-15

Interactive TUI overhaul.

### Added

- **Navigable command menus.** `/skills`, `/mcp`, `/workflows`, and a new `/plugins` open an
  interactive menu that lists what exists (each drilling into a detail view with its file path and
  an "edit via chat" action), shows a friendly empty state when there is nothing, and offers to
  create a new one. `/ontology`, `/graph`, and `/memory` use the same drill-in/back framework
  (breadcrumb, per-row descriptions, ↑/↓ · enter/→ · esc/←).
- **File-based workspace memory.** `/memory` is now a CRUD menu over Markdown notes under
  `.andromeda/memory/` — each with frontmatter (a consecutive id, title, tags, date) and a generated
  `MEMORY.md` index — with tag/text search. The index is folded into the agent's system prompt,
  alongside `AGENTS.md`, so the model can recall durable facts.
- **Header badge and version.** The top banner shows a colored active-mode badge (agent / plan /
  shell) and the running version.
- **Remembered setup.** A bare launch defaults to the provider and model you last used and skips
  onboarding when it builds cleanly.

### Fixed

- **Command output on the start screen.** Slash commands run before the first conversation turn no
  longer vanish behind the brand splash — the splash yields to the transcript as soon as a command
  produces output.
- **`update` is descriptive.** `andromeda update` (and `/update`) now report the running version,
  channel, and build, check the real release feed, explain a development build, and show the exact
  upgrade command instead of a bare "up to date". The TUI runs the check off the UI thread.

## [0.1.6] - 2026-07-15

### Added

- **Ontology and graph workspace maps.** Two context-engineering surfaces that give an agent
  (or a person) a fast, deterministic map of a repository before touching it.
  `andromeda ontology build` scans the workspace (honoring `.gitignore`) and writes a Turtle
  (`.ttl`) description of how files and directories relate — byte-for-byte reproducible, with no
  timestamps. `andromeda graph serve` derives a node/edge graph, writes `graph.json` plus
  human-readable Markdown notes, and opens a self-contained, dependency-free force-directed viewer
  on localhost. Both are exposed in the TUI as the `/ontology` and `/graph` slash commands, each
  with a build / show / adjust-via-chat / delete menu. Artifacts live under the gitignored
  `.andromeda/` surface (ADR-028).
- **Roadmap board automation (FR-GH-008).** `project.yml` now drives the full status lifecycle of
  the Andromeda Roadmap board, not just intake: a linked PR opening moves its issues to
  `In Review`, merging moves them to `Validation`, a published release moves every `Validation`
  item to `Released` and stamps the Target release field, and an issue closed as *not planned* is
  archived off the board. Linked issues are resolved via the PR's `closingIssuesReferences`, and
  transitions are forward-only (a merged PR never regresses a shipped item). The logic lives in
  `scripts/project_sync.py`; a maintainer runbook (`docs/maintainers/roadmap-board.md`) documents
  the board's statuses, fields, automations, and the five UI-only views.

## [0.1.3] - 2026-07-12

Hardening and distribution release. Binaries are rebuilt with a patched standard library and
Homebrew installation is now available. (Supersedes 0.1.1 and 0.1.2, whose binaries published
but whose cask-publishing steps failed on goreleaser template issues; use 0.1.3 or later for
`brew`.)

### Added

- **Homebrew tap.** `brew install datamaia/tap/andromeda` (the release publishes a cask to the
  `datamaia/homebrew-tap` repository, with a Gatekeeper-quarantine post-install step until
  notarization lands).
- **CI expansion.** The full Tier-1 platform matrix (Linux and macOS on amd64/arm64), an
  enforced `golangci-lint` gate, and `e2e`, `security` (CodeQL + govulncheck), and
  `traceability` workflows.

### Fixed

- **Terminal signal delivery.** `Signal` now targets the whole process group, so `Wait` can no
  longer hang when a shell orphans a child holding the output pipes.
- **Standard-library advisories.** The build toolchain is pinned to go1.25.12, clearing the
  crypto/tls and crypto/x509 advisories that govulncheck flagged in the go1.25.0 stdlib.
- Resolved all 374 `golangci-lint` findings across the codebase.

## [0.1.0] - 2026-07-12

First tagged release. Cross-platform signed binaries (Linux and macOS, amd64 and arm64)
with SBOMs and a cosign signature are published on the GitHub release; install via
`scripts/install.sh` or download an archive directly.

### Added

- **EP-01 — Repository, CI, and process foundations.** Go module and walking-skeleton
  binary (`andromeda version`); the `make ci` local quality gate (format, lint, build,
  race tests, coverage gate, spec lint, structure check); a lean Linux CI mirror; issue
  and pull-request templates; CODEOWNERS; Dependabot; label taxonomy; and the community,
  security, and governance files. Realizes the Volume 11 repository structure (FR-GH-002)
  and branching rules (FR-GH-003).
- **EP-02 — Architecture skeleton and PAL.** The L0 core domain (ULID, phases, closed
  capability/permission enumerations), the L1 `ports` package with all 18 frozen port
  interfaces (FR-ARCH-003), the Platform Abstraction Layer surfaces with Unix reference
  implementations (FR-PORT-001), the layer manifest with an import-graph dependency test
  (ADR-033), and the `sdk/` mirror module (ADR-031).
- **EP-03 — Persistence and configuration.** The monotonic ULID generator (ADR-027); the
  SQLite persistence layer with WAL, the workspace/global database split (ADR-028), and
  forward-only migrations with backups (ADR-029); a `SessionStore`; reusable stream
  implementations; and the Configuration Manager with layered precedence, env mapping, and
  source attribution (FR-CFG-001).
- **EP-04 — Observability foundation.** The in-process event bus (ADR-012), the event
  envelope and name grammar (FR-OBS-001), `slog` JSON logging with redaction (ADR-011),
  a local telemetry recorder, event persistence, and the `andromeda doctor` diagnostic.
  **Completes milestone MS-1 (Foundations).**
