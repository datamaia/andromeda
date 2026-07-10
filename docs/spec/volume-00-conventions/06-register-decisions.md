# 06 — Register: Decisions

Index of all architecture decision records. Full bodies live in `annexes/adr/ADR-NNN.md`,
following the template in [chapter 02](02-normative-language.md). Numbers are block-allocated
per [chapter 03](03-id-taxonomy-and-ownership.md); gaps are permanent. All dates 2026-07-11
unless noted.

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-001](../annexes/adr/ADR-001.md) | Implementation Language: Go | Accepted | Go, after scored evaluation of Rust/Go/TypeScript/Python; user preference recorded as a provided constraint |
| [ADR-002](../annexes/adr/ADR-002.md) | Project License: Apache-2.0 | Accepted | Apache-2.0 for patent grant and enterprise adoption; MPL-2.0 test-only dep (rapid) flagged in license policy |
| [ADR-003](../annexes/adr/ADR-003.md) | Repository Strategy: Monorepo | Accepted | Single `andromeda` monorepo: core, CLI/TUI, SDKs, docs, packaging |
| [ADR-004](../annexes/adr/ADR-004.md) | Branching Model: Trunk-Based Development | Accepted | Short-lived branches, PRs required, release branches for stabilization only; Git Flow evaluated and rejected |
| [ADR-005](../annexes/adr/ADR-005.md) | CLI Framework: spf13/cobra with pflag | Accepted | cobra + pflag; viper explicitly not used (Configuration Manager is custom) |
| [ADR-006](../annexes/adr/ADR-006.md) | TUI Framework: Charm Bubble Tea v2 Stack | Accepted | charm.land/bubbletea/v2 + lipgloss/v2 + bubbles/v2, pinned |
| [ADR-007](../annexes/adr/ADR-007.md) | Persistence: SQLite via modernc.org/sqlite | Accepted | Pure-Go SQLite, WAL mode, libc pinned; benchmark-gated review vs CGO driver |
| [ADR-008](../annexes/adr/ADR-008.md) | TOML Parsing: pelletier/go-toml/v2 | Accepted | go-toml/v2; BurntSushi/toml documented fallback |
| [ADR-009](../annexes/adr/ADR-009.md) | Plugin Mechanism: Subprocess with JSON-RPC 2.0 over stdio | Accepted | Language-agnostic Andromeda Runtime Protocol; go-plugin and Go native plugins rejected; WASM = v2 candidate |
| [ADR-010](../annexes/adr/ADR-010.md) | MCP Support: Official modelcontextprotocol/go-sdk | Accepted | Official Go SDK (stable v1); protocol revision pin PENDING VALIDATION |
| [ADR-011](../annexes/adr/ADR-011.md) | Observability: OpenTelemetry + log/slog | Accepted | OTel traces/metrics with local-first sinks + optional OTLP; stdlib slog JSON logging |
| [ADR-012](../annexes/adr/ADR-012.md) | Event Bus and External IPC | Accepted | In-process typed channel bus; external IPC via Unix domain socket + JSON-RPC 2.0; no broker |
| [ADR-013](../annexes/adr/ADR-013.md) | Release Tooling: goreleaser, cosign, syft, SLSA | Accepted | goreleaser + GitHub Actions; cosign v3 signing, syft SBOM, SLSA provenance, Homebrew tap; notarization PENDING VALIDATION |
| [ADR-014](../annexes/adr/ADR-014.md) | Credential Storage: OS keychains + age fallback | Accepted | zalando/go-keyring (Keychain/Secret Service) + opt-in age-encrypted file fallback; 99designs/keyring rejected (dormant) |
| [ADR-015](../annexes/adr/ADR-015.md) | SemVer + Conventional Commits | Accepted | SemVer for product and public contracts; Conventional Commits with 24 fixed scopes |
| [ADR-016](../annexes/adr/ADR-016.md) | Error Codes, Envelope, Exit Codes | Accepted | E-<AREA>-NNN codes, 14-field error envelope, closed exit codes 0–9 |
| [ADR-017](../annexes/adr/ADR-017.md) | Testing Stack | Accepted | stdlib testing + go-cmp + testify; rapid (property), native fuzzing, teatest/v2 (TUI) |
| [ADR-018](../annexes/adr/ADR-018.md) | Formatting and Linting | Accepted | gofmt + golangci-lint, curated set, pinned in CI |
| [ADR-019](../annexes/adr/ADR-019.md) | Provider HTTP Approach | Accepted | stdlib net/http against documented APIs; official SDKs adoptable per adapter (PENDING VALIDATION each); Ollama via thin HTTP client |
| [ADR-020](../annexes/adr/ADR-020.md) | Embeddings in SQLite, Exact Search for MVP | Accepted | Vectors in SQLite, in-process cosine similarity (≤100k chunks); ANN deferred (CGO conflict) |
| [ADR-021](../annexes/adr/ADR-021.md) | Layered Sandboxing | Accepted | MVP process-level controls; Beta/v1 OS-level isolation (Seatbelt/Landlock/bubblewrap) PENDING VALIDATION |
| [ADR-022](../annexes/adr/ADR-022.md) | Directory Layout: XDG via adrg/xdg | Accepted | XDG-style dirs, Apple-native mapping on macOS honoring XDG_* overrides, `.andromeda/` project-local |
| [ADR-023](../annexes/adr/ADR-023.md) | Concurrency Model | Accepted | Goroutines + context cancellation + errgroup; supervised tasks, bounded pools, backpressure |
| [ADR-024](../annexes/adr/ADR-024.md) | Schema Validation | Accepted | santhosh-tekuri/jsonschema/v6 for JSON Schema; typed layer over TOML config |
| [ADR-025](../annexes/adr/ADR-025.md) | Git Engine: System git Behind Adapter | Accepted | Shell out to system git (min 2.40) in PAL boundary; go-git read-only ops PENDING VALIDATION |
| [ADR-026](../annexes/adr/ADR-026.md) | Brand Identity and TUI Theming | Accepted | Closed design tokens (#7C5CFF/#8C7B6E/#F5F2ED/#121417, Geist/JetBrains Mono, cat mascot, tagline); dark-first theme, truecolor→monochrome tiers |
| [ADR-027](../annexes/adr/ADR-027.md) | ULID Primary Keys | Accepted | 26-char Crockford base32 ULIDs, monotonic per process; explicit sequence_no as ordering authority |
| [ADR-028](../annexes/adr/ADR-028.md) | Database Split: Workspace + Global | Accepted | `.andromeda/state.db` per workspace + global.db per machine (credentials global-only); index.db is a rebuildable cache |
| [ADR-029](../annexes/adr/ADR-029.md) | Forward-Only Migrations | Accepted | Pre-migration backup + integrity checks; clean refusal of future schemas; exit code 9; recovery via backup restore |
