# Compatibility Matrix

Consolidated compatibility reference for the Andromeda specification corpus, assembled at
consolidation (Phase C) from the volume registers and the chapters they index. This annex
**aggregates and cites; it decides nothing**: it mints no requirements, ADRs, error codes,
events, or configuration keys, and it renames nothing. Every table names its normative home;
if this annex and an owning chapter ever disagree, the owning chapter prevails and this annex
is corrected through the Volume 0 change procedure.

Version facts quoted from ADRs (library versions, dates) are the web-verified facts recorded
at each decision's date; the currency of a pin is governed by that ADR's review conditions,
not by this annex.

## 1. Platform tiers and minimums

Normative home: [Volume 3, chapter 07](../volume-03-architecture/07-platform-abstraction-layer.md)
(platform support matrix, FR-PORT-004, NFR-PORT-001); tier definitions in Volume 1,
chapter 05. Tier 1 platforms build, test, and gate every release on the full acceptance
suite; Tier 2 platforms build and smoke-test, and defects do not gate releases.

| Platform | Architecture | Minimum version | Tier | Phase |
|---|---|---|---|---|
| macOS on Apple Silicon | arm64 | macOS 13 | Tier 1 | MVP |
| macOS on Intel | x86_64 | macOS 13 | Tier 2 — PENDING VALIDATION (build/test capacity; Volume 3 register V3-OQ-1) | MVP when viable |
| Ubuntu | x86_64, arm64 | 22.04 LTS | Tier 1 | MVP |
| Debian | x86_64, arm64 | 12 | Tier 1 | MVP |
| Fedora | x86_64, arm64 | 39 | Tier 1 | MVP |
| Other Linux distributions | x86_64, arm64 | Kernel floor per Volume 3 notes | Best effort (no gate) | MVP |
| Windows 11 native | x86_64; arm64 subject to viability | Windows 11 | — | v2 candidate |
| Other Unix or Unix-like systems | — | — | — | Future |

Load-bearing notes carried with the matrix (Volume 3, chapter 07):

- **Linux kernel floor.** MVP assumes no kernel features beyond the Go runtime's requirements
  as shipped by the reference distributions (Ubuntu 22.04 ships 5.15). Landlock requires
  kernel ≥ 5.13 and is PENDING VALIDATION per ADR-021 before any Beta/v1 isolation claim;
  namespaces and bubblewrap availability are likewise PENDING VALIDATION per distribution.
  Capability absence degrades observably, never silently (E-PORT-002).
- **Linux binaries are statically linked** (ADR-001, ADR-007 cgo-free posture; NFR-PORT-003:
  0 dynamic library dependencies on Linux).
- **WSL is Linux.** Andromeda under WSL is supported as Linux and is never documented or
  marketed as Windows support.
- **Unsupported hosts refuse cleanly**: startup on a platform below the matrix floors exits
  with E-PORT-001 (exit code 3) before touching any state.
- **Runtime prerequisites are bounded** (NFR-PORT-003): system git ≥ 2.40 for Git features
  (ADR-025) plus optional platform services discovered by capability probes (FR-PORT-002).

### Reference shells

| Shell | Role |
|---|---|
| bash ≥ 5 | Reference POSIX-family shell; CI and script baseline |
| zsh | Reference interactive shell (macOS default); completion target |
| fish | Supported interactive shell; completion target |
| POSIX `sh` | Floor for generated scripts (install/uninstall per Volume 14) |

### Constrained environments

| Environment | Contract (Volume 3, chapter 07) |
|---|---|
| SSH sessions | Full CLI/TUI over a remote TTY; clipboard via OSC 52 fallback, notifications via terminal bell, per surface probes |
| Headless (no TTY) | CLI non-interactive mode and the IPC surface operate fully; the TUI refuses with a usage error; permission resolution is policy-only (exit code 5 on unresolved) |
| Containers | Supported for CLI/headless use; missing HOME/XDG roots engage the ADR-022 fallback with the `pal.fallback.engaged` diagnostic |
| CI (GitHub Actions reference) | Non-interactive contract as headless; deterministic exit codes; offline suite runs here with network disabled |

## 2. Terminal capability tiers

Normative home: [Volume 8, chapter 08](../volume-08-cli-and-tui/08-theming-and-design-tokens.md)
(color tiers, token-to-ANSI mapping, FR-UX-040, NFR-UX-040) and
[chapter 12](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) (glyph tiers,
no-color operation, terminal matrix, FR-TUI-066/067/068, NFR-TUI-070). Minimum supported
geometry is 80×24 with UTF-8 (Volume 3, chapter 07; Volume 8 layout rules govern narrower
terminals).

### Color tiers

Four tiers per the ADR-026 ladder, resolved per FR-TUI-008 (ADR-106 selection; ADR-103
aligns CLI styling signals):

| Tier | Rendering | Contrast posture |
|---|---|---|
| `truecolor` | Exact brand-token and derived-palette values (dark and light themes) | Text roles ≥ 4.5:1, non-text state-bearing elements ≥ 3:1 (NFR-UX-040) |
| `ansi256` | Nearest xterm-256 palette entries per role | Text roles keep ≥ 4.5:1 against the tier's background entry |
| `ansi16` | Named ANSI indices; terminal palette defines actual colors; no painted backgrounds, selection as reverse video | Thresholds met on the VGA reference palette; variance is RISK-UX-040 |
| `none` | No color attributes; bold, reverse video, and mandatory textual markers carry meaning | Complete UI, not degraded color (FR-TUI-066) |

Tier resolution ladder: explicit `tui.theme.tier` configuration → the ADR-103 styling
decision (`NO_COLOR`, `--color=never`, non-TTY force `none`) → `COLORTERM`
truecolor/24bit → `TERM` containing `256color` → terminfo color capability → `none`.
Danger and all semantic states are never color-only at any tier (FR-UX-040).

### Glyph tiers

Two closed glyph sets per ADR-112 with a normative parity table (FR-TUI-067): `unicode`
(box-drawing, braille spinner, `✓`/`✗`/`•`/`…`/`✦`) and `ascii` (`+ - |`, `- \ | /` spinner,
`ok`/`x`/`*`/`...`). Resolution: explicit `tui.glyphs` value, else `auto` = UTF-8 locale and
`TERM` not `linux`/`dumb` → `unicode`; every other combination → `ascii`. One set per
process; chrome draws only from the resolved set.

### Terminal compatibility matrix

Expected capability profiles for test planning (Volume 8, chapter 12). Every cell is
PENDING VALIDATION until the Volume 13 compatibility suite verifies it per release; a
probe-versus-matrix mismatch on a Tier A terminal fails the release gate (NFR-TUI-070).

| Terminal | Tier | Color (expected) | Glyphs | Mouse | Bracketed paste | OSC 52 |
|---|---|---|---|---|---|---|
| iTerm2 (macOS) | A | truecolor | unicode | yes | yes | yes |
| Terminal.app (macOS) | A | 256-color | unicode | yes | yes | no (expected) |
| GNOME Terminal | A | truecolor | unicode | yes | yes | to verify |
| Konsole | A | truecolor | unicode | yes | yes | to verify |
| Alacritty | A | truecolor | unicode | yes | yes | yes |
| kitty | A | truecolor | unicode | yes | yes | yes |
| WezTerm | A | truecolor | unicode | yes | yes | yes |
| tmux (over any above) | A | inherits ≤ outer | unicode | yes | yes | with passthrough setting |
| GNU screen (over any above) | A | ≤ 256-color | unicode | limited | to verify | with passthrough setting |
| Linux console | B | 16-color | ascii | no | no | no |
| Unknown `TERM` (resolvable) | B | 16-color baseline | per locale | no | no | no |
| `TERM=dumb` / non-TTY | — | CLI path only (E-CLI-005 for TUI) | — | — | — | — |

Tier A rows gate releases on both Tier 1 operating systems where the terminal exists;
multiplexer rows are tested over at least two distinct outer terminals; tmux and GNU screen
are first-class targets with OSC 52 clipboard fallback over SSH (FR-TUI-068, ADR-113,
ADR-114).

## 3. Provider adapters: phases, authentication, capabilities

Normative home: [Volume 5, chapter 09](../volume-05-providers-and-auth/09-provider-adapters-catalog.md)
(catalog, FR-PROV-080..083) with the capability model of chapter 02 (FR-PROV-010/011,
ADR-056). Governing rules: **generic-adapter-first** (ADR-065 — any OpenAI-compatible
service is reachable from MVP through `openai_compatible` with a configured base URL and
key) and **no invented facts** (unverified capability, endpoint, pricing, and rate-limit
details are PENDING VALIDATION, register entry V5B-OQ-1). Catalog presence does not imply
MVP support; it implies a committed phase and a stable slug.

| # | Adapter slug | Service | Kind | Dedicated-adapter phase | Auth family | Transport |
|---|---|---|---|---|---|---|
| 1 | `openai_compatible` | Any OpenAI-compatible endpoint | Cloud or local | MVP | `none` or `api_key` | HTTP + SSE streaming |
| 2 | `anthropic` | Anthropic | Cloud | MVP | `api_key` | HTTP + SSE streaming |
| 3 | `ollama` | Ollama local server | Local | MVP | `none` | HTTP (localhost) |
| 4 | `openai` | OpenAI | Cloud | Beta | `api_key` | HTTP + SSE streaming |
| 5 | `gemini` | Google Gemini API | Cloud | Beta | `api_key`; Vertex surface PENDING VALIDATION | HTTP + streaming |
| 6 | `openrouter` | OpenRouter | Cloud aggregator | Beta | `api_key` | HTTP + SSE streaming |
| 7 | `mistral` | Mistral La Plateforme | Cloud | Beta | `api_key` | HTTP + SSE streaming |
| 8 | `azure_openai` | Azure OpenAI | Cloud | v1 | `api_key`; Entra ID / managed identity PENDING VALIDATION | HTTP + SSE streaming |
| 9 | `groq` | Groq | Cloud | v1 | `api_key` | HTTP + SSE streaming |
| 10 | `together` | Together AI | Cloud | v1 | `api_key` | HTTP + SSE streaming |
| 11 | `deepseek` | DeepSeek | Cloud | v1 | `api_key` | HTTP + SSE streaming |
| 12 | `xai` | xAI | Cloud | v1 | `api_key` | HTTP + SSE streaming |
| 13 | `vllm` | vLLM server | Local or self-hosted | v1 | `none` or `api_key` (server-configured) | HTTP + SSE streaming |
| 14 | `lm_studio` | LM Studio local server | Local | v1 | `none`; details PENDING VALIDATION | HTTP (localhost) |
| 15 | `llama_cpp` | llama.cpp server | Local | v1 | `none` or `api_key`; details PENDING VALIDATION | HTTP (localhost) |
| 16 | `litellm` | LiteLLM proxy | Self-hosted aggregator | v2 | `api_key` (proxy-issued); details PENDING VALIDATION | HTTP + SSE streaming |
| 17 | `localai` | LocalAI | Local | v2 | `none` or `api_key`; details PENDING VALIDATION | HTTP (localhost) |
| 18 | `fastchat` | FastChat server | Local or self-hosted | Future | details PENDING VALIDATION | HTTP |
| 19 | `text_generation_webui` | Text Generation WebUI API | Local | Future | details PENDING VALIDATION | HTTP |

### Capability vocabulary and how capability facts resolve

The capability enum (frozen seed formalized by Volume 5, chapter 02; extended once by
ADR-056) has fifteen members: `chat`, `streaming`, `tool_calling`, `parallel_tool_calling`,
`structured_outputs`, `reasoning`, `vision`, `audio_input`, `audio_output`, `embeddings`,
`token_usage_reporting`, `cost_reporting`, `model_discovery`, `cancellation`,
`token_counting`.

This annex deliberately publishes **no adapter × capability truth table**: per-model
capability values are resolved at runtime from provenance (`declared`, `discovered`,
`configured`, `verified`, masked by `refuted` — ADR-056), and simulating an absent
capability silently is prohibited (FR-PROV-011 degradation strategies: `refuse`,
`report_unavailable`, opt-in `substitute`, `reroute`). What the catalog commits per adapter:

| Adapter | Committed capability posture (Volume 5, chapter 09) |
|---|---|
| `openai_compatible` | Declared surface: chat completions + SSE streaming; models listing (`model_discovery`) and embeddings endpoints optional and detected, never assumed; every other capability established per model by detection plus configuration override (FR-PROV-081) |
| `anthropic` | Documented Messages API: chat, streaming, tool calling, per-model context/output metadata, `token_usage_reporting`, model discovery; reasoning exposed only as officially provided summaries (FR-PROV-082) |
| `ollama` | Documented local REST API: chat, streaming, embeddings, discovery of locally installed models; per-model capabilities by detection (local models vary in tool-calling and structured-output fidelity); anchors the offline guarantee (FR-PROV-083, FR-PROV-085) |
| Adapters 4–19 | All capability specifics PENDING VALIDATION at each adapter's implementation from official documentation (V5B-OQ-1); OpenAI-compatible services reachable earlier via adapter 1 |

MVP exit requires `openai_compatible`, `anthropic`, and `ollama` to pass their adapter
conformance suites (FR-PROV-080); local-serving conformance is tracked by NFR-PROV-002.

## 4. MCP protocol revisions

Normative home: [ADR-010](adr/ADR-010.md); runtime behavior in Volume 6, chapters 05–06
(FR-MCP-001).

| Item | Commitment |
|---|---|
| Implementation | `github.com/modelcontextprotocol/go-sdk`, pinned within the stable v1 major line (v1.6.1 verified at decision date) |
| Protocol revisions negotiable | 2025-11-25, 2025-06-18, 2025-03-26, 2024-11-05 (the SDK's verified supported set; negotiation delegated to the SDK) |
| Pinned/certified revision set | PENDING VALIDATION — deferred until the specification RC cycle beginning 2026-07-28 settles and the SDK declares support (ADR-010 rule 3; Volume 6 register V6B-OQ-1) |
| Out-of-set servers | A server reporting a revision outside the SDK's supported set fails connection with a defined error; negotiated revision and server implementation version are recorded per connection (Volume 6, chapter 05) |
| OAuth-based server authorization | PENDING VALIDATION — SDK client OAuth is experimental; not a stable MVP feature (ADR-010 rule 4; V6B-OQ-2). Supported MVP path: stdio servers and token/header HTTP auth with ADR-014 credential handling; FR-MCP-004 (Beta) covers the non-OAuth authorization surface |
| Type isolation | SDK types never leak past the MCP Runtime port wrapper (ADR-010 rule 1) |

## 5. Go toolchain and pinned dependencies

Normative homes: the foundation ADRs cited per row (bodies in [`adr/`](adr/)). Policy: every
production dependency is pinned; majors are adopted deliberately (a new major line requires
the owning ADR's review, never a routine bump). Version numbers are the verified facts at
each ADR's decision date (2026-07-11).

| Concern | Dependency | Pinned line / floor | ADR | Notes |
|---|---|---|---|---|
| Language toolchain | Go | ≥ 1.24 minimum at adoption | ADR-001 | Floor driven by the verified anthropic-sdk-go requirement; single primary language |
| CLI framework | `github.com/spf13/cobra` + `github.com/spf13/pflag` | Pinned; current major | ADR-005 | viper explicitly prohibited (custom Configuration Manager) |
| TUI stack | `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `charm.land/bubbles/v2` | v2 majors, exact-version pins (v2.0.8 / v2.0.5 / v2.1.1 verified) | ADR-006 | `charm.land` vanity paths mandatory; v1 `github.com/charmbracelet/*` paths banned by CI lint |
| TUI test harness | `github.com/charmbracelet/x/exp/teatest/v2` | Exact pseudo-version, test-only | ADR-006, ADR-017 | Experimental upstream; contained by pinning and test-only scope |
| Persistence | `modernc.org/sqlite` | v1 line (v1.53.0 verified; embeds SQLite 3.53.2) | ADR-007 | WAL via DSN pragma; `modernc.org/libc` pinned to the exact version in the driver's `go.mod`, CI-enforced |
| TOML | `github.com/pelletier/go-toml/v2` | `/v2` module path, pinned (v2.4.3 verified; TOML v1.1.0) | ADR-008 | Strict decoding; `BurntSushi/toml` is the wrapped, documented fallback |
| Plugins (ARP) | None — stdlib JSON-RPC 2.0 over stdio | — | ADR-009 | Subprocess protocol; no plugin library dependency |
| MCP | `github.com/modelcontextprotocol/go-sdk` | Stable v1 major (v1.6.1 verified) | ADR-010 | See section 4 |
| Observability | OpenTelemetry Go SDK + stdlib `log/slog` | Pinned at implementation; OTLP exporters `http/protobuf` and `grpc` | ADR-011 | Semantic-conventions pin PENDING VALIDATION (Volume 10 register V10B-OQ-1); third-party logging frameworks prohibited |
| Event bus / IPC | None — typed channels; Unix domain socket + JSON-RPC 2.0 | — | ADR-012 | No broker dependency |
| Release tooling | goreleaser / cosign / syft | goreleaser v2 (v2.17.0), cosign v3 major pinned in CI (v3.1.1), syft v1 (v1.46.0) — all verified | ADR-013 | Keyless signing; SLSA provenance; Homebrew tap |
| Credential storage | `github.com/zalando/go-keyring` + `filippo.io/age` | Pinned (v0.2.8 / v1.3.1 verified) | ADR-014 | OS keychain first; age-encrypted file fallback strictly opt-in |
| Testing stack | stdlib `testing`, `github.com/google/go-cmp`, testify, `pgregory.net/rapid`, native fuzzing | Pinned (rapid v1.3.0 verified, MPL-2.0, test-only) | ADR-017 | rapid's MPL-2.0 flagged in the ADR-002 license policy |
| Format / lint | gofmt + golangci-lint | Linter version pinned in CI and contributor tooling | ADR-018 | Curated linter set |
| Provider HTTP | stdlib `net/http` | — | ADR-019 | Official provider SDKs adoptable per adapter, each PENDING VALIDATION (openai-go v3.41.1 and anthropic-sdk-go v1.57.0 verified as candidates); Ollama via thin hand-rolled client |
| Directories | `github.com/adrg/xdg` | Pinned (v0.5.3 verified) | ADR-022 | XDG on Linux; Apple-native mapping honoring `XDG_*` overrides |
| Concurrency | stdlib `context` + `golang.org/x/sync` errgroup | Pinned | ADR-023 | Supervised tasks, bounded pools |
| JSON Schema | `github.com/santhosh-tekuri/jsonschema/v6` | `/v6` major, pinned (v6.0.2 verified) | ADR-024 | Drafts 2020-12, 2019-09, 7, 6, 4; earlier majors not adopted |
| PAL syscalls | `golang.org/x/sys` | Pinned; `internal/pal` only | ADR-001, FR-PORT-001 | Prohibited outside the PAL |

### Runtime prerequisites (not linked libraries)

| Prerequisite | Floor | Home |
|---|---|---|
| System git | ≥ 2.40, detected at runtime with defined refusal below the floor | ADR-025; Volume 11 |
| Optional platform services (keychain, notifications, clipboard, isolation mechanisms) | Probed per PAL surface; declared degradation on absence | Volume 3, chapter 07 (FR-PORT-002) |

## 6. Configuration and schema version compatibility

Normative homes: [Volume 10, chapter 01](../volume-10-config-storage-observability/01-configuration-model.md)
(FR-CFG-008, ADR-133), [ADR-029](adr/ADR-029.md) (database schema),
[Volume 14](../volume-14-distribution/00-index.md) (pairing, upgrade paths, contract
stability).

### Configuration schema (`andromeda.toml`)

| Scenario | Behavior | Codes / events |
|---|---|---|
| File declares the current `config_version` | Loads normally | — |
| File declares an older `config_version` | Forward-only transforms upgrade it **in-memory** on every load; the file on disk is untouched | `config.migration.applied` on the explicit surface only |
| Persistent rewrite requested | Only via the explicit migration command: verified backup first, then rewrite, re-validate, report | E-CFG-011 on failure; `config.migration.failed` |
| File declares a **newer** `config_version` than the build supports | Refused — never partially interpreted | E-CFG-010 (exit code 3) |
| Deprecated key in use | Still resolves via its migration alias; diagnostic reported with replacement and removal release | E-CFG-013; `config.deprecation.detected` |

Schema-tightening rule: a validation-rule addition that rejects previously accepted
documents ships only with a schema version bump and a migration or deprecation path
(FR-CFG-008); contract-diff discipline per SM-20 applies.

### Database schema (workspace and global databases)

Forward-only per ADR-029, operationalized by FR-CFG-009: numbered immutable embedded
migrations, `schema_migrations` ledger cross-checked against `PRAGMA user_version`,
verified backup plus integrity checks before migrating, strictly ordered transactional
application.

| Scenario | Behavior | Codes |
|---|---|---|
| Older database opened by a newer binary | Migrates forward after backup and pre-checks | E-CFG-016 on failure, E-CFG-018 if the backup cannot be created, E-CFG-017 on integrity failure (all exit code 9) |
| Newer database opened by an older binary | Clean refusal naming both versions; partial reads of future schemas forbidden | E-CFG-015 (exit code 9) |
| Migration failure | Stop; preserve current file and backup; recovery is restore-from-backup — no downgrade scripts exist | E-CFG-016 |
| Cache databases | Never migrate; layout changes drop and rebuild them | — |

### Binary–database pairing, upgrade paths, and public contracts

- **Product downgrade** reverts the *pair*: Updater binary rollback (ADR-192, FR-REL-008)
  with database restore per Volume 10 — never the schema alone (ADR-029 rule 6). Rollback
  across a schema boundary requires the explicit pairing dialogue (ADR-192).
- **Supported upgrade paths** are validated per release (FR-REL-014, ADR-193); an
  unsupported path refuses with E-REL-009. Update channels form the closed enum `stable`,
  `rc`, `beta`, `nightly` with maturity-floor offer semantics (ADR-191); majors are never
  auto-applied.
- **Product and public contracts follow SemVer** (ADR-015, FR-REL-012): from v1, breaking
  changes to public contracts (CLI grammar and `--json` schemas, configuration schema, event
  envelope, SDK and ARP surfaces, port contracts) require a major; contract-diff tooling per
  release backs NFR-REL-003 (SM-20) and NFR-ARCH-002.
- **Event envelope and payload compatibility** is governed by Volume 10's envelope
  versioning (FR-OBS-001); the ARP handshake carries its own protocol version negotiation
  (ADR-009, FR-PLUG-001); MCP revision negotiation per section 4.
