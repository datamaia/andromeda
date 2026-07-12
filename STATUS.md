# Andromeda — Implementation Status

Living tracker of the build. The **specification** (a private companion document, v1.0.0) is complete; this
file tracks the **implementation** against Volume 15's epics and milestones. Updated and
pushed on every advance.

**Last updated:** 2026-07-12 · **MVP phase functionally complete** · **Public repo · v0.1.0 released (signed)** ·
**Ports:** 18/18 ✅ · **CLI:** 20 command groups · **TUI:** ✅ · **Built-in tools:** 20/20 ·
**CI:** matrix (Linux amd64/arm64, macOS arm64), golangci-lint gate green · ~20k LOC Go, 60 test files, `make ci` green

## MVP-minimum coverage (Volume 1 chapter 05, 27 items)

Every item of the change-controlled MVP minimum is implemented:

| Item | Status | Item | Status |
|---|---|---|---|
| Functional CLI | ✅ (20 groups) | Streaming | ✅ (SSE adapters) |
| Functional TUI | ✅ (Bubble Tea v2) | Configuration | ✅ (FR-CFG-001) |
| Agent runtime | ✅ (FR-AGT-001) | Logging | ✅ (slog+redaction) |
| Basic planner | ✅ (in-loop) | Session persistence | ✅ (SQLite) |
| Execution engine | ✅ | macOS | ✅ (dev + CI) |
| Context manager | ~ (assembly in loop) | Linux | ✅ (CI matrix) |
| Tool runtime | ✅ (FR-TOOL-001) | Installation | ✅ (shell installer) |
| Permission manager | ✅ (FR-SEC-100) | Basic update | ✅ (UpdaterPort) |
| Workspace engine | ✅ | Unit tests | ✅ (42 files) |
| Terminal | ✅ (TerminalPort) | Integration tests | ✅ |
| Filesystem tools | ✅ (full MVP catalog 8/8) | Main E2E | ✅ (doctor + run) |
| Basic Git | ✅ (GitPort + git_exec) | GitHub Actions | ✅ (10 workflows; SHA-pinned) |
| Provider abstraction | ✅ (FR-PROV-001) | Signed releases | ✅ (v0.1.0, cosign) |
| ≥1 cloud provider | ✅ (Anthropic, OpenAI-compat) | ≥1 local provider | ✅ (Ollama) |

Beyond the minimum, also implemented: **SDD Workflow Engine**, **MCP client + tool bridging**,
**HTTP MCP transport with OAuth bearer**, **plugin subprocess runtime (ARP)**, **WASM plugin
runtime (wazero)**, **skill system**, **scheduler**, **package manager**, **auth layer + OAuth
device grant**, **secret store (keychain+age)**, **sandbox (process + macOS Seatbelt / Linux
bubblewrap)**, **JSON Schema tool-payload validation**, and the **complete built-in tool catalog (20/20 —
filesystem, git, terminal, http, sqlite, process, docker, kubernetes, browser, and the six
service integrations as transport surfaces)**.

## Remaining work — all spec-designated later phases (Beta/v1/v2) or refinements

Per the specification's own phasing and PENDING VALIDATION items — not part of the MVP:

- **OS-level sandbox**: macOS **Seatbelt** (`sandbox-exec`) now implemented and verified
  (enforces the write policy; effective containment reported as `os`). Linux Landlock/bubblewrap
  remains a follow-up (ADR-021 PENDING VALIDATION per platform).
- ✅ **Windows PAL backends** implemented (build-tagged): Credential Manager (via keyring),
  known-folder config/data dirs, paths, temp files, exclusive-create file locking, plus Windows
  sandbox exec and terminal signal (taskkill). The whole tree **cross-compiles for windows/
  amd64 and windows/arm64**; runtime validation on a Windows host remains (v2, no host here).
- ✅ **OAuth device authorization grant** implemented (`auth`, ADR-063, RFC 8628):
  start-device-flow + poll-token (honoring authorization_pending, slow_down, and deadlines) +
  store, verified with an httptest OAuth server.
- ✅ **HTTP MCP transport with OAuth bearer** implemented (`mcp`, ADR-010): JSON-response mode of
  the Streamable HTTP transport (POST-per-request, also parses `text/event-stream` responses),
  adapting the newline-delimited JSON-RPC framing to HTTP. `BearerFromSecretStore` binds it to
  the device grant — the token the OAuth flow stored is read from the Secret Store as the bearer
  on each request. Verified with an httptest MCP server (bearer attached, anonymous, 401 fails a
  single call, SSE response, and the full Secret-Store→bearer path).
- ✅ **WASM plugin runtime** implemented (`plugin`, ADR-009 v2, wazero): in-process WebAssembly
  plugins with no host capabilities, bridged to permission-mediated ToolPorts; verified with an
  embedded wasm module. (Was previously a v2 candidate; wazero is pure-Go so it is testable here.)
- ✅ **macOS notarization wiring** added to goreleaser, credential-gated (activates when
  MACOS_SIGN_P12 is set); it needs an Apple Developer identity to run (OQ-003) but is ready.
- ✅ **Semantic embeddings retrieval** now implemented (`indexer.SemanticEngine`): embeds files
  and answers queries by in-process cosine similarity (ADR-020), backed by any `ProviderPort`
  via `ProviderEmbedder`; verified with a deterministic embedder.
- ✅ **Live config watch** now implemented (`ConfigPort.Watch`): polls tracked config files'
  mtimes and emits `ConfigChange` events for affected keys matching the selector; verified by
  editing a file and observing the change.
- ✅ **PTY terminal mode** now implemented (`TerminalPort` PTY path via creack/pty): a real
  pseudoterminal with merged output, stdin write, and window `Resize`; verified end-to-end.
- ✅ **OpenTelemetry SDK export** now implemented (`telemetry.OTel`): a TelemetryPort backed by
  the official OTel Go SDK (ADR-011) — metrics via counters, nested spans with attributes and
  error/status, and flush; export is consent-gated at the caller (Volume 10). Verified.
- ✅ **Bubble Tea v2 migration done** (ADR-006): the TUI now uses the Charm v2 stack via the
  `charm.land/*/v2` vanity paths (v2 `KeyPressMsg`, `View() tea.View`, declarative AltScreen); a
  `make lint-charm` gate bans the v1 `github.com/charmbracelet/*` paths. The start screen renders
  the **brand splash** — the ASCII cat mascot with sparkles, the wordmark, and the tagline
  (ADR-026, from `docs/brand/banner-sketch.png`).
- ✅ **Per-stage SDD agent wiring** now implemented (`app.RunSDD`): each of the 14 SDD stages
  runs an agent goal scoped to the stage and objective; `andromeda workflow run sdd --goal ...`
  drives it, gates halt without `--auto-approve`. Verified with a scripted provider.
- ✅ **All mandated CLI commands** now present (added `context` and `trace`): the full Volume 8
  command surface — run, plan/exec (via run), init (via workspace open), config, auth, provider,
  model, tool, plugin (via runtime), skill, workflow, mcp (via client), memory, context, index,
  git, doctor, update, version, completion, logs, trace, export, tui.

The repository is **public**, and **branch protection** on `main` is applied (force-push and
deletion blocked, administrators included). The full traceability automation (Volume 11 ch 07
GitHub-side checks) remains platform configuration applied on GitHub, not in-repo code.


## How work is organized

Implementation follows the Volume 15 epic sequence. Milestone **MS-1 "Foundations"** =
EP-01 → EP-04, whose exit is: *a binary that starts on macOS and Linux, resolves
configuration with attribution, opens both databases with migrations and backups, and emits
enveloped events to persisted storage.*

The authoritative quality gate is **`make ci`** (runs locally, no CI-minute dependency).

## Legend

✅ done · 🔄 in progress · ⬜ pending

## Milestones

| Milestone | Epics | Status |
|---|---|---|
| MS-1 Foundations | EP-01 ✅, EP-02 ✅, EP-03 ✅, EP-04 ✅ | ✅ |
| MS-2 Runtime core | EP-05 ✅, EP-06 ✅, EP-07 ✅ | ✅ |
| MS-3 Memory/index/tools/agent | memory, indexer, tool, agent ✅ | ✅ |
| MS-4 Usable CLI + TUI | CLI (20 cmds) + interactive TUI ✅ | ✅ |
| MS-4+ | TUI, auth, MCP/skills/plugins, dist | ⬜ |

## Epics

### EP-01 — Repository, CI, and process foundations · 🔄 (near complete)

Realizes FR-GH-002 (repository structure) and FR-GH-003 (branching rules).

- ✅ Go module (`github.com/datamaia/andromeda`, `go 1.24`; toolchain 1.26 installed)
- ✅ Walking-skeleton binary: `andromeda version` (cobra, ADR-005) + `internal/buildinfo`
- ✅ `Makefile` — the authoritative local gate (`make ci`): fmt-check, lint, build, race
  tests, coverage gate, spec lint, structure check, tidy check
- ✅ `.golangci.yml` (ADR-018) with a depguard seed for the layer rules (ADR-033)
- ✅ Lean CI mirror `.github/workflows/ci.yml` (single Linux job — see deviations)
- ✅ Issue forms (15 types + `config.yml`), PR template, CODEOWNERS, `dependabot.yml`,
  `labels.yml`
- ✅ Community/governance files: LICENSE (Apache-2.0, ADR-002), CONTRIBUTING, CODE_OF_CONDUCT,
  SECURITY, GOVERNANCE, MAINTAINERS, CHANGELOG
- ✅ `scripts/structure_check.sh` (FR-GH-002 structure check)
- ✅ Commit-message hook already active (`.githooks/commit-msg`, ADR-015)
- ⬜ Full traceability validators (`scripts/traceability/`, branch-grammar/commit/linkage
  checks) — chapter 07; deferred to a dedicated EP-01 follow-up issue
- ⬜ Branch protection settings on `main` (configured on GitHub, not in-repo)

**Gate status:** `make ci` passes — build OK, tests green with `-race`, coverage 80.8%
(≥ 70% SM-14 floor), spec lint 0/0, structure-check OK.

### EP-02 — Architecture skeleton and PAL · ✅

Realizes FR-ARCH-003 (port freeze), FR-ARCH-001/004 (layering, context propagation),
FR-PORT-001 (platform encapsulation).

- ✅ L0 core (`internal/core`): ULID, `Phase`, and the closed capability (15), permission
  (13), scope (10), and decision (7) enumerations with wire-value guard tests
- ✅ L1 ports (`internal/ports`): **all 18 frozen port interfaces** with faithful signatures
  (`Provider`, `Auth`, `Tool`, `Terminal`, `MemoryStore`, `Indexer`, `EventBus`, `Permission`,
  `SecretStore`, `Sandbox`, `Config`, `SessionStore`, `Git`, `Workspace`, `Scheduler`,
  `Updater`, `Package`, `Telemetry`), shared `Stream[T]`/`PortError`/error-family primitives,
  and minimal contract types (grow additively per owning volume)
- ✅ Platform Abstraction Layer (`internal/pal`): the 19 surface interfaces; Unix reference
  implementations for Paths, ConfigDirs (XDG-honoring), TempFiles, and FileLocking (flock)
  with tests
- ✅ Dependency-rule enforcement (`internal/arch`): layer manifest + an import-graph test
  (ADR-033) that runs in `make test` with no external tooling; depguard seed in
  `.golangci.yml`
- ✅ `sdk/` mirror module (second Go module per ADR-031; builds independently, does not import
  `internal/`)
- ⬜ Concrete Unix implementations of the remaining PAL surfaces (Processes, Signals, PTY,
  Shell, CredentialStore, …) — delivered by their owning epics
- ⬜ SDK mirror content (port contract mirror) — Extension SDK epics

**Gate status:** `make ci` passes — coverage 85.5%; 18/18 ports; import-graph rule enforced.

### EP-03 — Persistence and configuration · ✅

Realizes FR-CFG-001 (configuration precedence) and the ADR-007/028/029 persistence decisions.

- ✅ ULID generator (`internal/core`, ADR-027): monotonic, Crockford base32, with uniqueness
  and sort-order tests
- ✅ SQLite persistence (`internal/storage`, ADR-007): pure-Go modernc driver, WAL mode,
  `libc` pinned; workspace `state.db` + machine `global.db` split (ADR-028)
- ✅ Forward-only migrations (ADR-029): `user_version` tracking, pre-migration file backup,
  `integrity_check` + `foreign_key_check`, and clean refusal of future schemas
- ✅ `SessionStore` implementing `SessionStorePort`: sessions with optimistic-concurrency
  revisions, run records with per-run sequencing, and crash-recovery `MarkInterrupted`
- ✅ Reusable `Stream[T]` implementations (`internal/streams`): slice- and channel-backed
- ✅ Configuration Manager (`internal/config`) implementing `ConfigPort`: layered precedence
  (defaults → global → profile → workspace → project → runtime → env → cli), `ANDROMEDA_*`
  env mapping, TOML parsing (go-toml/v2), validation with E-CFG findings, per-value source
  attribution, and a file loader assembling layers from disk via the PAL
- ⬜ Live config file watching (`Watch` currently returns an empty stream) — later epic
- ⬜ Typed schema validation (ADR-024) beyond TOML syntax — later epic

**Gate status:** `make ci` passes — coverage 82.0%.

**Decision recorded (to reconcile with Volume 10 FR-CFG-004):** the `ANDROMEDA_*` env-var
mapping uses `__` to separate config-table levels and treats a single `_` as literal within a
key segment (so `ANDROMEDA_AGENT__LOOP__MAX_ITERATIONS` → `agent.loop.max_iterations`), with a
single-underscore fallback when no `__` is present (so the spec's `ANDROMEDA_TUI_THEME_MODE →
tui.theme.mode` example still works). This resolves the underscore ambiguity the spec itself
flags; Volume 10's text should adopt the same rule.

**Config key alignment:** the loop-iteration default is keyed `agent.loop.max_iterations` (the
Volume 10 catalog's canonical name) and is now actually consumed: `RunAgent` reads it from the
resolved configuration when no `--max-iterations` flag/override is set, so `andromeda.toml` can
cap the agent loop (flag > config > engine default).

### EP-04 — Observability foundation · ✅

Realizes FR-OBS-001 (event envelope) and the ADR-011/012 observability decisions. **Completes
milestone MS-1.**

- ✅ Event bus (`internal/eventbus`) implementing `EventBusPort` (ADR-012): in-process typed
  pub/sub, exact-name and prefix topic selectors, bounded per-subscriber buffers with a
  drop-oldest overflow policy that never blocks publishers, context-cancel auto-close
- ✅ Event envelope (FR-OBS-001): `NewEvent` builder (version, UTC timestamp, correlation ID)
  and event-name grammar validation (`<area>[.<noun>].<verb-past>`)
- ✅ Structured logging (`internal/logging`, ADR-011): `slog` JSON handler with level control
  and secret redaction at the handler (Volume 9 redaction)
- ✅ Telemetry (`internal/telemetry`) implementing `TelemetryPort`: local-first metric registry
  and span tree; never fails the observed operation
- ✅ Event persistence (`internal/storage`): workspace-DB `events` table (migration v2) and an
  `EventStore` writing/reading enveloped Event records
- ✅ **`andromeda doctor`** composition (`internal/app`) exercising the MS-1 exit end to end:
  resolves config with attribution, opens both databases with migrations, emits and persists
  an enveloped event — verified live and by hermetic tests
- ⬜ OpenTelemetry Go SDK + OTLP export wiring (consent-gated) — local sinks now; SDK later
- ⬜ Full metric/trace/cost catalogs — owning volumes' later epics

**Gate status:** `make ci` passes — coverage 77.0%. `andromeda doctor` exits 0 with all checks
green.

## Milestone MS-1 — Foundations · ✅ COMPLETE

Exit criterion met: the `andromeda` binary starts on macOS (verified) and Linux (CI), resolves
configuration with source attribution, opens both the workspace and global databases with
migrations and backups, and emits enveloped events to persisted storage — all demonstrated by
`andromeda doctor` and covered by tests.

### EP-05 — Security kernel · ✅

- ✅ Permission Manager (`internal/permission`) implementing `PermissionPort` (**FR-SEC-100**):
  the closed 13/10/7 enums, the deny > ask > allow > else-ask evaluation algorithm (ADR-121),
  scope enclosure (session/workspace) and selector matching (exact, `/prefix/**`, `/prefix/*`),
  standing grants with expiry and revocation, policy rules, interactive approvals minting
  scoped grants, decision persistence and per-decision audit (fail-closed on audit failure,
  E-SEC-014), and unknown-permission → E-SEC-002 deny. Backed by workspace-DB migration v3.
- ✅ Secret Store (`internal/secret`) implementing `SecretStorePort` (**FR-SEC-102**, ADR-014):
  OS keychain backend via zalando/go-keyring (PAL `CredentialStore`, macOS/Linux) with an
  opt-in age-encrypted file fallback (passphrase/scrypt, 0600, verified encrypted-at-rest and
  wrong-passphrase-rejected); a per-namespace reference index so `List` works on keychains
  that cannot enumerate; secrets never surface in errors (E-SEC PortErrors carry no material)
- ✅ Sandbox Engine (`internal/sandbox`) implementing `SandboxPort` (**FR-SEC-101**, ADR-021):
  process-level MVP controls — deny-by-default env filtering (with sensitive-name stripping),
  command allow/deny lists, working-directory path policy, wall-clock time limit, and
  process-group teardown that kills the whole tree; effective containment level (`process`) is
  observable and never silently weakened. OS-level isolation (Seatbelt/Landlock) is the
  Beta/v1 layer (PENDING VALIDATION).
- ⬜ Approval state machine wiring to a real driver (CLI/TUI) — with EP-13

**Gate status:** `make ci` passes — coverage ~74%.

### EP-06 — Workspace and Git engines · ✅

- ✅ Git Engine (`internal/git`) implementing `GitPort` (**FR-GIT-001**, ADR-025): shells out
  to system git (≥ 2.40, gated at `Version`); `Status` (porcelain), `Diff`/`Log` (streamed),
  `Show`, `Stage`/`Unstage`, `Commit`, `ListBranches`/`CreateBranch`/`SwitchBranch`,
  `ApplyPatch` (check-then-apply, atomic), and worktree add/list/remove; failures map to E-GIT
  with git's stderr as safe detail. Verified against real temporary repositories.
- ✅ Workspace Engine (`internal/workspace`) implementing `WorkspacePort`: upward `Discover`
  (`.andromeda/` marker or `.git`), `Open` (creates the marker, opens the workspace database,
  registers in the global registry with a stable ID across reopens), `Snapshot` (root,
  projects, VCS summary via the Git Engine, timestamp), and clean `Close`.

**Gate status:** `make ci` passes. Ports implemented: 10 / 18.

### EP-07 — Providers, models, and routing · ✅

Realizes the provider contract (**FR-PROV-001**) and the MVP provider seed.

- ✅ Provider Layer base (`internal/provider`): the E-PROV error family with HTTP→code mapping
  and retryability, a shared JSON/SSE HTTP client, and the **Router** that itself implements
  `ProviderPort` — failing over primary→fallbacks only on retryable errors (connectivity, rate
  limit, 5xx) and never on auth/bad-request, so a fallback can't mask a misconfiguration or run
  a costly retry on a correctly-rejected request; emits change notices (Transparent AI).
- ✅ OpenAI-compatible adapter (`provider/openaicompat`, FR-PROV-081): Chat, streaming Chat
  (SSE), Embed, model discovery, capability declaration; the generic adapter covering many
  services.
- ✅ Anthropic adapter (`provider/anthropic`): Messages API Chat and streaming (content-block
  deltas), system-message extraction, capability declaration.
- ✅ Ollama adapter (`provider/ollama`): local `/api/chat`, `/api/embed`, `/api/tags`
  discovery — a thin hand-rolled client (ADR-019).
- All adapters use documented public APIs only and are verified with `httptest` mock servers
  (no real network). Capabilities are declared honestly; token counting returns the
  unavailable error so the Context Manager estimates.

**Milestone MS-2 (Runtime core) complete.** Ports implemented: **14 / 18** (adds Provider;
Auth/Tool/Terminal/Memory/Indexer/Updater/Package remain).

### EP — Memory and Indexing (Volume 7) · 🔄

- ✅ Memory Store (`internal/memory`) implementing `MemoryStorePort` (**FR-MEM-001**):
  transactional ingest with provenance, layer- and text-filtered retrieval, term-overlap
  ranking, retention expiry, hard delete, and streamed export — over workspace-DB migration v4.
- ✅ Indexing Engine (`internal/indexer`) implementing `IndexerPort` (**FR-IDX-001**): an
  in-memory lexical inverted index over workspace files with the frozen Index lifecycle
  (created→building→ready→updating→stale), incremental `Update`, `Query` with generation-tagged
  hits, `Invalidate`, and `Status`; excludes `.git`/`.andromeda`/configured paths and skips
  binary and oversized files. Indexes are rebuildable caches (INV-IDX-02).
- ⬜ Context Manager (`CTX`), semantic embeddings retrieval (ADR-020), and memory
  encryption/redaction hooks — later increments.

**Gate status:** `make ci` passes. Ports implemented: **16 / 18** (adds MemoryStore, Indexer;
Auth, Tool/Terminal, Updater, Package remain).

### EP — Tool Runtime and built-in tools (Volume 6) · 🔄

- ✅ Tool Runtime (`internal/tool`) implementing the mediation of `ToolPort` (**FR-TOOL-001**):
  registry, per-invocation input validation, permission evaluation via `PermissionPort` with
  **denial-as-data** (a refused invocation returns a terminal error event, not a transport
  failure), and path-level permission derivation for tools that declare their resources.
- ✅ **JSON Schema payload validation (FR-TOOL-002, ADR-024)**: the Runtime compiles each tool's
  declared `InputSchema` at registration (santhosh-tekuri/jsonschema v6) — a malformed schema is
  rejected there — and validates every invocation's input against it before the tool's own
  semantic `Validate`, so structural conformance is a Runtime guarantee independent of the tool.
- ✅ Built-in tools (`internal/tool/builtin`) — the **full MVP catalog (8/8, FR-TOOL-007)**:
  `fs_read`, `fs_write`, `fs_search`, `fs_replace` (exact/regex, unique-unless-replace_all),
  `fs_diff` and `fs_patch` (a self-contained offline unified-diff engine — compute and atomic
  all-or-none apply that round-trips), `git_exec` (structured operations over the Git Engine;
  reads need `read`, mutations additionally request `git_mutation` per operation), and
  `terminal_run`. Each is a `ToolPort` with input/output schemas, declared permissions, and
  path/repository-scoped resource queries so grants apply per file/path/repo. Wired into the
  `andromeda run` composition (write-gated tools appear only with `--allow-write`).
- ✅ **Full built-in tool catalog implemented (20/20, FR-TOOL-007)** — all phases:
  - MVP (8): `fs_read/write/search/replace/diff/patch`, `git_exec`, `terminal_run`.
  - Beta/v1 network/data/platform (3): `http_request` (per-host `network`, `credential_ref`
    resolved from the Secret Store and never echoed, redirect/body caps), `sqlite_query` (refuses
    state DBs, classifies read vs. mutation, DB-level `query_only` so a CTE-hidden write can't
    slip past), `process_control` (list/inspect/signal/terminate over the Terminal Engine).
  - Container runtimes (2): `docker_control` (docker CLI), `kubernetes_control` (kubectl;
    `exec` additionally gated `execute`) — structured operation→argv, never a raw shell string; a
    missing runtime binary surfaces as a tool error.
  - Service integrations (6): `github/gitlab/jira/slack/notion/linear_request` over a shared
    transport-and-schema surface. Per ADR-074 the per-service endpoints/auth/rate-limits are
    PENDING VALIDATION, so **none are hardcoded** — the base URL and credential ref come from
    `[services.<name>]` config and the Secret Store; the transport injects auth, executes, and
    surfaces rate-limit headers. Secret material never appears in the record.
  - Browser (1): `browser_control` over the **W3C WebDriver** standard (not a proprietary
    mechanism) against a config-supplied driver endpoint.
  Every tool is a `ToolPort` with input/output schemas, declared permissions, and scoped resource
  queries; verified with httptest servers, mock WebDriver, and stub-binary CLI runtimes. The
  network/data/platform tools are wired into `andromeda run`; the container/service/browser tools
  need their runtime/endpoint/credentials configured before use (not available on this host).

**Gate status:** `make ci` passes. Ports implemented: **17 / 18** (Tool; Terminal, Auth,
Updater, Package remain).

### EP — Agent Engine (Volume 4) · 🔄

- ✅ Agent Engine (`internal/agent`) implementing the plan–act–observe loop (**FR-AGT-001**):
  one mode-invariant loop that sends the conversation plus tool declarations to the provider,
  executes returned tool calls through the mediated Tool Runtime (permissions, denial-as-data),
  feeds results back, and iterates to a tool-free answer or the iteration budget (E-AGT-001 on
  exhaustion); accumulates token usage, honors context cancellation (→ `cancelled`), and
  persists run records through `SessionStorePort`. Verified with a scripted provider and fake
  tools (tool round-trip, immediate finish, budget exhaustion, cancellation).
- ⬜ Planner as a separate component, sub-agent delegation, prompt engine, full Run/Task state
  machines — later increments.
- ⬜ `andromeda run` CLI wiring (compose agent + real provider from config + fs tools) — next.

**Gate status:** `make ci` passes. The keystone agent loop is implemented and tested.

### EP — Usable CLI agent (`andromeda run`) · ✅ (MVP slice)

- ✅ End-to-end composition (`internal/app`): `RunAgent` opens the workspace, sets up the
  permission manager with **safe-by-default grants** (read within the workspace subtree; write
  only with `--allow-write`), registers the built-in filesystem tools in the Tool Runtime,
  persists a session, and drives the Agent Engine loop.
- ✅ `BuildProvider` selects and configures an adapter (Ollama local, OpenAI-compatible,
  Anthropic) — unknown names error rather than silently defaulting; cloud adapters require a
  key.
- ✅ **`andromeda run <goal>`** command with `--provider/--base-url/--api-key-env/--model/
  --system/--allow-write/--max-iterations` flags. Verified: reads a real file end-to-end via a
  scripted provider; write denied without `--allow-write`; an unreachable provider fails
  cleanly with `E-PROV-005` and a run ID.

**This is the MVP payoff: a real agent that plans, calls permission-mediated tools against the
workspace, and answers — driven from the command line.** Point `--provider ollama` at a running
Ollama, or `--provider anthropic --api-key-env ANDROMEDA_ANTHROPIC_KEY`, for a live run.

**Gate status:** `make ci` passes. Ports implemented: **17 / 18** (Terminal, Auth, Updater,
Package remain).

### EP — Terminal Engine and command execution · ✅

- ✅ Terminal Engine (`internal/terminal`) implementing `TerminalPort`: pipe-based streaming
  execution with tagged stdout/stderr chunks, bounded capture with explicit truncation, stdin
  `Write`, portable `Signal` (interrupt/terminate/kill → Unix signals), and `Wait` returning the
  command outcome. Verified: stdout capture, non-zero exit, stdin piping, and kill-stops-sleep.
- ✅ `terminal_run` built-in tool routing through the Tool Runtime (requires `execute`
  permission; command-scoped resource query), wired into `andromeda run` behind `--allow-exec`.
  Verified end-to-end: the agent runs a real command and observes its output.
- ⬜ PTY mode, sandbox-policy integration for terminal executions — later increments.

**Port recount (corrected):** implemented **15 / 18** — Provider, Auth⬜, Tool, **Terminal✅**,
MemoryStore, Indexer, EventBus, Permission, SecretStore, Sandbox, Config, SessionStore, Git,
Workspace, Scheduler⬜, Updater⬜, Package⬜, Telemetry. Remaining: **Auth, Scheduler, Updater,
Package**.

### EP — Remaining ports: Scheduler, Auth, Updater, Package · ✅  → **ALL 18 PORTS DONE**

- ✅ Task Scheduler (`internal/scheduler`) implementing `SchedulerPort` (ADR-023): named bounded
  pools (interactive/tools/background/io), supervised submit with panic capture, cancellation,
  structured groups with first-error propagation (errgroup semantics), and pool stats. Verified:
  bounded concurrency, panic capture, group error propagation, cancel-before-start, shutdown.
- ✅ Authentication Layer (`internal/auth`) implementing `AuthPort` (Volume 5): API-key intake
  stored only behind Secret Store references (ADR-014), `Authenticate`/`Refresh`/`Revoke`/
  `Rotate`/`ListProfiles`, `none` mechanism for local providers, and a clean "unsupported
  mechanism" error for OAuth flows deferred to a later phase. Secrets never returned or logged.
- ✅ Updater (`internal/updater`) implementing `UpdaterPort` (Volume 14): channel check,
  download, SHA-256 verify, atomic binary swap with retained backup, and offline rollback;
  Apply refuses to run unless Verify passed. Verified end-to-end with local artifacts.
- ✅ Package Manager (`internal/pkgmgr`) implementing `PackagePort` (Volume 6): resolve →
  install through the frozen installation states with checksum verification, verify, and remove;
  a failed checksum leaves nothing active. Verified end-to-end.

## 🎯 All 18 port interfaces now have real, tested implementations.

Provider, Auth, Tool, Terminal, MemoryStore, Indexer, EventBus, Permission, SecretStore,
Sandbox, Config, SessionStore, Git, Workspace, Scheduler, Updater, Package, Telemetry — every
frozen Volume 3 port is implemented and exercised by tests, and the layer dependency rule is
enforced on every commit.

### EP — CLI command surface (Volume 8) · 🔄

- ✅ Real, working commands wired to the engines, each tested by driving the command:
  `run` (agent), `doctor`, `version`, `config show [--json]` (resolved config with source
  attribution), `git status|log` (Git Engine), `memory add|list|search` (Memory Store over
  SQLite), `tool list`, `index query` (lexical Indexer), `auth add|list|remove` (credentials
  via the keychain/age Secret Store — keys read from an env var, never the command line).
- ✅ Verified live: `andromeda git status` shows the real repository state; `memory add`/
  `search` round-trips through the workspace database; `index query` searches real files.
- ⬜ Remaining commands (provider, model, plugin, skill, workflow, mcp, context, logs, trace,
  export, update, completion) and the full flag grammar / exit-code table — later.
- ⬜ TUI (Volume 8), MCP/skills/plugins (Volume 6), Workflow Engine/SDD (Volume 4), and
  distribution/release (Volume 14) — remaining milestones.

**Gate status:** `make ci` passes. **All 18 ports implemented;** a usable multi-command CLI with
a working agent.

### EP — Workflow Engine (SDD) and more commands · 🔄

- ✅ Workflow Engine (`internal/workflow`) implementing specification-driven development
  (**FR-WF-001**): a stage executor driving the frozen Workflow Run states (pending → running →
  awaiting_approval → completed/failed/cancelled/interrupted), human-approval **gate** stages,
  stage/gate/run events, and **resume from a stage boundary**. The 14-stage SDD pipeline
  (intake → release-preparation) is a built-in definition with the correct gate stages.
- ✅ Verified: full-pipeline completion, stage-failure → failed, gate-halt without approver,
  gate approval, and resume; the CLI runs the 14-stage pipeline live.
- ✅ More CLI commands: `workflow list|run sdd [--auto-approve]`, `provider list`,
  `model list` (live model discovery via the provider adapters).
- ⬜ Per-stage agent wiring for SDD, workflow persistence/resume across processes,
  gate approvals through the TUI — later.

**Gate status:** `make ci` passes. CLI groups: **13**. All 18 ports done; agent + SDD workflow
shell both runnable from the command line.

### EP — MCP client, distribution, and remaining CLI · 🔄

- ✅ JSON-RPC 2.0 stdio transport (`internal/jsonrpc`): newline-delimited framing (MCP-faithful),
  a background read loop demultiplexing responses to concurrent callers, context cancellation,
  and error propagation. Shared by MCP (ADR-010) and the Andromeda Runtime Protocol (ADR-009).
- ✅ MCP client (`internal/mcp`, **FR-MCP-001**): `initialize`, `tools/list`, `tools/call`, and a
  **bridge** exposing MCP tools as `ToolPort`s (origin `mcp`, untrusted) so the Tool Runtime
  mediates them like built-ins. Verified against an in-memory fake MCP server.
- ✅ Distribution (Volume 14, ADR-013): `.goreleaser.yaml` (darwin/linux × amd64/arm64,
  checksums, SBOMs, keyless cosign signing, Homebrew tap), the `release.yml` tag-triggered
  workflow (least-privilege, OIDC), and a checksum-verifying `packaging/installer/install.sh`.
- ✅ More CLI: `update` (channel check) and `completion` (bash/zsh/fish).
- ⬜ Plugin subprocess runtime over ARP, skill format/loader, MCP server subprocess spawning,
  the TUI, and `logs/trace/export/context` commands — remaining.

**Gate status:** `make ci` passes. CLI groups: **15**. Distribution pipeline configured.

### EP — Plugin Runtime and Skill System (Volume 6) · ✅

- ✅ Plugin Runtime (`internal/plugin`, **FR-PLUG-001**): plugins are subprocesses speaking the
  Andromeda Runtime Protocol (JSON-RPC 2.0 over stdio, ADR-009) — the same surface as MCP, so
  the MCP client drives both. `Connect` (injectable transport) and `Spawn` (real subprocess)
  do the handshake, `Tools` bridges the plugin's tools to permission-mediated ToolPorts, and
  `Stop` runs the frozen Plugin lifecycle. Verified with an in-memory server AND a real
  subprocess (a POSIX-sh ARP server responding to initialize).
- ✅ Skill System (`internal/skill`, **FR-SKILL-001**): a loader that parses and validates a
  skill directory (`skill.toml` manifest + prompt), and `Resolve` that checks required tools
  and capabilities against the environment, composing the system prompt only when satisfied and
  reporting missing requirements precisely (no silent degradation).
- ✅ WASM plugin runtime (`internal/plugin/wasm.go`, **ADR-009 v2**): in-process WebAssembly
  plugins on wazero (pure-Go, no CGO — builds and tests on every Tier-1 platform). Guests run
  with no host capabilities (no filesystem, clock, or network). `Runtime.LoadWASM` instantiates
  a module and tracks it for shutdown; `Runtime.Close` releases the runtime and all modules. A
  module's `run(ptr,len)->i64` ABI bridges to a permission-mediated ToolPort (untrusted origin);
  compute-only guests use `CallI32`. Verified end to end with an embedded minimal wasm module.

**Extensibility axis complete:** MCP client + tool bridging, plugin subprocess runtime over ARP,
the in-process WASM plugin runtime (wazero), and the skill format/loader — all permission-mediated
through the Tool Runtime.

**Gate status:** `make ci` passes.

### EP — TUI and observability commands · ✅

- ✅ Terminal UI (`internal/tui`, Volume 8): a **Bubble Tea v2** session model with a scrollable
  transcript, a prompt input, and a brand-styled status bar (design tokens from ADR-026 —
  violet primary, off-white, taupe, the fixed danger red). The start screen renders the ASCII cat
  mascot splash (wordmark + tagline). The **responder drives the real agent** for each submitted
  goal. Update/render logic is unit-tested headlessly (submit, backspace, empty-submit no-op,
  esc-quit, ctrl+c-quit, splash/status render, splash-hidden-after-exchange).
- ✅ `andromeda tui` command wiring the model to a provider + agent (`--provider/--model/
  --allow-write/--allow-exec`).
- ✅ `logs` (recent persisted events from the workspace event store) and `export` (sessions as
  JSON) commands, tested by driving them.

**Gate status:** `make ci` passes. **20 CLI command entries**; an interactive TUI that runs the
agent.

**ADR-006 compliance:** the TUI uses the Charm **v2** stack via the `charm.land/*/v2` vanity
paths; `make lint-charm` fails the build if any v1 `github.com/charmbracelet/*` import reappears.

## Deliberate deviations from the specification (free-tier accommodations)

Recorded honestly so they can be reverted when the constraint lifts.

1. **CI platform matrix — RESOLVED (repo is public).** The Tier-1 matrix (Linux amd64/arm64,
   macOS arm64) runs on every push/PR, and the mandated `policy.yml`, `traceability.yml`,
   `security.yml`, `e2e.yml`, `labels.yml`, `audit.yml`, `upgrade.yml`, and `docs.yml` workflows
   are all active (10 workflows total). macOS Intel (macos-13) is omitted — those hosted runners
   are being retired. Still pending: `benchmarks.yml` (no benchmark suite exists yet) and
   `project.yml` (needs a GitHub Project board).
2. **Action pinning — RESOLVED.** Every workflow action is pinned to a full commit SHA (ADR-149)
   with its version in a trailing comment; Dependabot's `github-actions` updater bumps them, and
   `policy.yml` fails any unpinned action.
3. **CODEOWNERS is a single maintainer** (`@datamaia`) rather than maintainer teams; teams
   activate once the GitHub org exists (namespace PENDING VALIDATION, OQ-001).
4. **Structure check enforces the EP-01 subset** of the mandated tree; it extends as later
   epics add `sdk/`, `schemas/`, `packaging/`, `.goreleaser.yaml`, etc.

## Environment notes

- Go was not installed at the start; installed via Homebrew (`go1.26`, satisfies `go 1.24`).
- `golangci-lint` is optional locally (`make lint` degrades to gofmt+vet); CI installs it
  pinned. Install locally to run the full linter set.
