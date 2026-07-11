# Spine Pack — Phase B authoring contract

**Working document. Not part of the published specification** (deleted at consolidation;
Volume 0 is the authoritative home of everything here). You are one of 18 parallel agents
writing Volumes 4–14. This file is your binding contract: it tells you what to read, what is
frozen, what you own, and how to finish. Deviating from it creates cross-volume contradictions
that a later audit will bounce back to your volume.

## 1. Mandatory reading (in this order, before writing anything)

1. This file, fully.
2. `docs/spec/volume-00-conventions/` — every chapter. Normative keywords, banned vague
   adjectives, the exact FR/NFR/RISK/ADR templates (lint-enforced headings), identifier
   taxonomy/ownership, glossary (exact entity/component names), single-home matrix, exit
   codes, event naming, config-table ownership.
3. `docs/spec/volume-03-architecture/02-port-interfaces.md` — the 18 frozen ports. You
   elaborate semantics for ports you own; you MUST NOT rename ports or methods.
4. `docs/spec/volume-02-domain-model/09-lifecycles-and-canonical-states.md` — frozen state
   vocabularies. If you own a full state machine, use those exact state names.
5. `docs/spec/volume-01-vision-and-product/05-scope-and-phases.md` (phase definitions, MVP
   minimum) and `06-success-metrics.md` (metrics your volume must formalize as NFRs).
6. `docs/spec/volume-00-conventions/06-register-decisions.md` — all 33 ADRs. Open the bodies
   in `docs/spec/annexes/adr/` for decisions touching your area. You MUST NOT contradict any
   accepted ADR.
7. `/Users/maia/Documents/lyra/andromeda/prompt.md` — the project brief (Spanish; you write
   ENGLISH): the line ranges for your volume given in §8, plus global sections 97–129
   (normative rules), 154–197 (assumptions policy — never invent external facts), 2754–2771
   (acceptance criteria), 2776–2819 (error/event field requirements).

## 2. Writing rules (identical for every agent)

- English only. Volume 0 normative keywords (MUST/MUST NOT/SHOULD/SHOULD NOT/MAY/PENDING
  VALIDATION/OUT OF SCOPE/DEPRECATED).
- Banned vague adjectives (fast, easy, intuitive, robust, scalable, secure, efficient, modern,
  advanced, compatible, powerful) never appear in a normative statement without a metric or
  verifiable criterion.
- No `TBD`, no `TODO`, no placeholders. Make a reasoned decision and document it, or define an
  abstraction and mark the concrete part PENDING VALIDATION with an open-questions entry in
  your register file.
- Never invent: private APIs, undocumented endpoints, provider capabilities, prices, rate
  limits, licenses, library versions, unofficial auth mechanisms. External integrations use
  official, documented mechanisms only. A subscription never implies programmatic access.
- Entity, component, port, and state names EXACTLY as frozen (glossary, Volume 3 ch 02,
  Volume 2 ch 09). One name per concept; never reuse a name for a different concept.
- Every FR uses the FULL Volume 0 template: all 9 metadata bullets and all 19 `####` sections
  with real content. Acceptance criteria in Given/When/Then where it adds precision and
  covering negative, error, permission, and observability cases. Every NFR/RISK uses its full
  bullet template. Phase values only: Core, MVP, Beta, v1, v2, Future, Out of Scope.
- Every Mermaid diagram is accompanied by prose describing components, relations, constraints.
  Fence languages: `pseudo` (typed pseudocode), `toml`, `yaml`, `json`, `text`, `mermaid`.
  Intentionally invalid examples: ` ```toml invalid ` / ` ```json invalid `.
- State machines you own define: initial state, terminal states, transitions, events, guards,
  side effects, persistence, recovery, timeouts, cancellation, retries, errors — using the
  frozen state names.
- Every error you mint (`E-<AREA>-NNN`) declares the full ADR-016 envelope: code, category,
  severity, user message, technical message, cause, safe context data, recoverability, retry
  policy, recommended action, exit-code mapping, HTTP mapping where applicable, telemetry
  event, security implications.
- Events you mint follow `<area>.<noun>.<verb-past>` (e.g., `tool.invocation.denied`); the
  envelope is Volume 10's (reference it; do not redefine).
- Config keys you mint live in your TOML tables per Volume 0 ch 03; Volume 10 owns schema,
  precedence, and validation rules — reference, never restate.
- Dangerous actions bind to the permission model: use permission names from §5 only.

## 3. What you own (and nothing else)

Your agent row in §8 lists: your files (disjoint from every other agent — write ONLY those),
your ID areas and numeric ranges, your ADR block, and your keystone requirements. Keystones
are pre-assigned IDs+topics you MUST define with exactly that ID and topic. You MUST NOT mint
IDs outside your ranges, edit files outside your list, or reference a foreign-area ID unless
it is (a) already defined in Volumes 0–3, (b) a keystone listed in §7, or (c) an ADR ≤ 033.
Reference everything else by name (entity/component/port/chapter).

## 4. Frozen technical facts (cite, do not re-decide)

Go (ADR-001) · Apache-2.0 (ADR-002) · monorepo (ADR-003) · trunk-based (ADR-004) · cobra
(ADR-005) · Bubble Tea v2 `charm.land` (ADR-006) · SQLite modernc, WAL (ADR-007) · go-toml/v2
(ADR-008) · plugins = subprocess JSON-RPC 2.0 "Andromeda Runtime Protocol" (ADR-009) · MCP
official Go SDK stable v1 (ADR-010) · OTel + slog (ADR-011) · channel event bus + Unix-socket
JSON-RPC IPC (ADR-012) · goreleaser/cosign/syft/SLSA (ADR-013) · zalando/go-keyring + age file
fallback (ADR-014) · SemVer + Conventional Commits, commit messages carry change info only —
never AI/vendor attribution (ADR-015) · E-codes + exit codes 0–9 (ADR-016) · testing stack
(ADR-017) · gofmt + golangci-lint (ADR-018) · provider HTTP baseline (ADR-019) · embeddings in
SQLite, exact cosine ≤100k chunks (ADR-020) · layered sandbox: MVP process-level, OS-level
Beta/v1 PENDING VALIDATION (ADR-021) · XDG dirs, `.andromeda/` (ADR-022) · concurrency
conventions (ADR-023) · jsonschema/v6 (ADR-024) · system git ≥2.40 behind adapter (ADR-025) ·
brand design tokens (ADR-026) · ULIDs (ADR-027) · workspace+global DB split (ADR-028) ·
forward-only migrations (ADR-029) · hexagonal layers + dependency matrix (ADR-030) · module
layout (ADR-031) · headless mode Beta (ADR-032) · dependency enforcement in CI (ADR-033).

Design tokens (ADR-026): Primary `#7C5CFF`, Secondary `#8C7B6E`, Tertiary `#F5F2ED`, Neutral
`#121417`, Danger fixed by Volume 8; Geist (docs) / JetBrains Mono (code; recommended-never-
required in terminals); ASCII cat mascot + four-pointed star; tagline "Your terminal companion
for shipping great software."

## 5. Frozen enum seeds

**Permissions** (Volume 9 formalizes; everyone uses these names): `read`, `write`, `execute`,
`network`, `credential_access`, `git_mutation`, `process_spawn`, `container_access`,
`external_service_access`, `clipboard`, `notifications`, `package_installation`,
`system_modification`.
**Permission scopes**: `session`, `workspace`, `command`, `tool`, `provider`, `host`, `path`,
`domain`, `repository`, `organization`.
**Permission decisions**: `allow_once`, `allow_for_session`, `allow_for_workspace`,
`always_allow_policy`, `deny_once`, `always_deny`, `ask_every_time`.
**Capabilities** (Volume 5 formalizes and MAY extend; everyone uses these names): `chat`,
`streaming`, `tool_calling`, `parallel_tool_calling`, `structured_outputs`, `reasoning`,
`vision`, `audio_input`, `audio_output`, `embeddings`, `token_usage_reporting`,
`cost_reporting`, `model_discovery`, `cancellation`.

## 6. Registers and multi-agent mechanics

- Single-agent volumes (7, 11, 12, 13, 14) write `99-volume-register.md` directly.
- Multi-agent volumes (4, 5, 6, 8, 9, 10) — each agent writes its own fragment
  `98-register-a.md` / `98-register-b.md` / `98-register-c.md` with the SAME sections; the
  orchestrator merges fragments into `99-volume-register.md` and deletes them. Do not write
  `99-volume-register.md` yourself in a multi-agent volume.
- Register sections (all required): `## Requirements index` (table: | ID | Title | Phase |
  Verification method |, listing every FR/NFR/RISK you minted), `## ADRs minted`,
  `## Error codes minted`, `## Events minted`, `## Config keys minted`, `## Glossary
  additions`, `## Assumptions`, `## Open questions` (every PENDING VALIDATION you used),
  `## Cross-volume references`.
- `00-index.md` for your volume is written ONLY by the agent marked "(writes 00-index)" in §8,
  listing ALL chapters of the volume per the fixed map, status "Authored (draft)".

## 7. Keystone requirements (pre-assigned IDs — define exactly these)

FR-AGT-001 agent loop · FR-WF-001 SDD workflow · FR-PROV-001 provider contract · FR-AUTH-001
official-mechanisms-only authentication · FR-TOOL-001 tool contract · FR-SDK-001 extension SDK
· FR-MCP-001 MCP client support · FR-SKILL-001 skill format · FR-PLUG-001 plugin runtime (ARP)
· FR-MEM-001 memory model · FR-CTX-001 context assembly · FR-IDX-001 indexing · FR-CLI-001 CLI
grammar · FR-TUI-001 TUI shell · FR-SEC-100 permission model · FR-SEC-101 sandbox · FR-SEC-102
secret storage · FR-CFG-001 configuration precedence · FR-OBS-001 event envelope · FR-GIT-001
git engine · FR-GH-001 traceability automation · FR-REL-001 release pipeline.

Any agent may REFERENCE any keystone; only the owner defines it.

## 8. Agent assignments

Brief line ranges refer to `/Users/maia/Documents/lyra/andromeda/prompt.md`. Word targets are
for the agent's whole file set; stay within ±20%.

### Volume 4 — `volume-04-agent-runtime/`
- **Agent 4A** (writes 00-index) — mints FR/NFR/RISK-AGT-*, E-AGT-*; ADR 040–047; ~12k words.
  Brief: 400–428, 2090–2125. Files: `00-index.md`, `01-agent-engine.md` (the agent loop —
  keystone FR-AGT-001; plan/act/observe cycle over ports; turn handling; interruption/resume),
  `02-planner.md` (plan production/revision, plan states per Vol 2), `03-execution-engine.md`
  (task dispatch over ToolPort/SchedulerPort, approvals, retries), `04-prompt-engine.md`
  (versioned templates, rendering, profile parameters), `05-core-state-machines.md` (FULL
  machines: Session, Run, Agent, Plan, Task — frozen names), `98-register-a.md`.
- **Agent 4B** — mints FR/NFR/RISK-WF-*, E-WF-*; ADR 048–054; ~11k words. Brief: 2045–2087.
  Files: `06-workflow-engine-and-sdd.md` (keystone FR-WF-001; the 14 SDD workflow stages
  intake→requirements→research→planning→architecture→task-decomposition→implementation→
  validation→testing→review→security-review→documentation→completion→release-preparation, each
  with inputs/outputs/states/transitions/agents/tools/permissions/entry-exit criteria/retry/
  cancellation/resume/rollback/artifacts/events/timeouts/errors/human-approval/audit),
  `07-workflow-run-state-machine.md` (FULL machine), `08-skill-engine-runtime.md` (execution
  semantics only — format is Volume 6's), `09-task-scheduler.md` (SchedulerPort semantics,
  supervision per ADR-023), `98-register-b.md`.

### Volume 5 — `volume-05-providers-and-auth/`
- **Agent 5A** (writes 00-index) — mints FR/NFR/RISK-PROV-001–079, E-PROV-*; ADR 055–061;
  ~12k words. Brief: 430–459, 1738–1788. Files: `00-index.md`, `01-provider-contract.md`
  (keystone FR-PROV-001; ProviderPort behavioral contract; adapter declaration set: identity,
  adapter version, auth method, endpoints, discovery, capabilities, context window,
  modalities, tool calling, streaming, structured outputs, reasoning, vision, audio,
  embeddings, token usage, costs, rate limits, error mapping, retry/timeout policy,
  cancellation, idempotency, health, metadata, deprecations), `02-capabilities-model-
  discovery.md` (capability enum formalization from §5, negotiation, explicit detection, no
  silent simulation, degradation strategies, user notification on any provider/model change),
  `03-streaming-toolcalling-structured-outputs.md`, `04-token-and-cost-accounting.md`,
  `05-resilience-routing-fallback.md` (rate limits, retries, timeouts, circuit breakers,
  routing, fallback + rules preventing unsafe/costly fallback), `06-error-normalization.md`,
  `98-register-a.md`.
- **Agent 5B** — mints FR-PROV-080+, FR/NFR/RISK-AUTH-*, E-AUTH-*; ADR 062–069; ~11k words.
  Brief: 1791–1831, 915–940. Files: `07-authentication-layer.md` (keystone FR-AUTH-001; API
  keys, OAuth, device authorization flow, official service accounts/managed identity, account
  auth ONLY where officially documented for third-party clients, enterprise proxies, temporary
  credentials, rotation, revocation, multiple profiles; prohibited mechanisms list),
  `08-credential-lifecycle.md` (SecretStorePort/AuthPort usage per ADR-014; never plaintext),
  `09-provider-adapters-catalog.md` (all 19: OpenAI, Anthropic, Google Gemini, OpenRouter,
  Azure OpenAI, Groq, Together, DeepSeek, xAI, Mistral, Ollama, LM Studio, vLLM, LiteLLM,
  llama.cpp, LocalAI, FastChat, Text Generation WebUI, generic OpenAI-compatible — per
  adapter: phase, auth method, transport, capability notes as PENDING VALIDATION where
  undocumented; MVP = generic OpenAI-compatible + Anthropic + Ollama), `10-local-and-offline-
  operation.md` (Vol 1 offline guarantees mapped to provider behavior), `11-state-machines.md`
  (FULL: Authentication Session, Provider connection), `98-register-b.md`.

### Volume 6 — `volume-06-tools-mcp-skills-plugins/`
- **Agent 6A** (writes 00-index) — mints FR/NFR/RISK-TOOL-*, FR/NFR-SDK-*, E-TOOL-*;
  ADR 070–076; ~12k words. Brief: 461–484, 1834–1891. Files: `00-index.md`,
  `01-tool-sdk-and-contract.md` (keystones FR-TOOL-001, FR-SDK-001; full tool declaration:
  name, namespace, description, version, author, origin, trust level, input/output JSON
  Schema per ADR-024, permissions from §5, risks, timeout, retries, idempotency, streaming,
  cancellation, lifecycle, sandbox profile, errors, events, telemetry, compatibility, tests,
  examples), `02-tool-lifecycle-permissions-trust.md`, `03-builtin-tools-catalog.md` (20
  tools with phases: filesystem read/write/search/replace/patch/diff, git, terminal, process,
  http, browser, docker, kubernetes, sqlite, github, gitlab, jira, notion, slack, linear —
  each: purpose, schema sketch, permissions, phase; NOT all MVP), `04-tool-invocation-state-
  machine.md` (FULL machine), `98-register-a.md`.
- **Agent 6B** — mints FR/NFR/RISK-MCP-*, -SKILL-*, -PLUG-*, E-MCP/SKILL/PLUG-*; ADR 077–084;
  ~13k words. Brief: 1977–2009, 2012–2042. Files: `05-mcp-client-and-runtime.md` (keystone
  FR-MCP-001; official SDK per ADR-010: connections, transports, discovery, registration,
  installation, configuration, authorization, tools/resources/prompts, lifecycle, health,
  logs, update, uninstall, versioning, compatibility, errors, timeouts),
  `06-mcp-security-and-conformance.md` (trust model, isolation, supply chain, conformance
  tests), `07-skill-format-and-system.md` (keystone FR-SKILL-001; manifest, identity,
  versioning, I/O, prompts, required tools/capabilities, compatible providers, workflows,
  dependencies, inheritance, composition, overrides, testing, fixtures, publication,
  installation, signing, trust, future marketplace, deprecation), `08-plugin-runtime-and-
  arp.md` (keystone FR-PLUG-001; the Andromeda Runtime Protocol: JSON-RPC 2.0 framing,
  handshake/version negotiation, method surface, permissions, sandbox, lifecycle),
  `09-package-manager-supply-chain.md` (install/discover/register/update/uninstall,
  dependencies, versioning, signatures per ADR-013 patterns, marketplace future),
  `10-state-machines.md` (FULL: Plugin, MCP Client Connection, Package installation),
  `98-register-b.md`.

### Volume 7 — `volume-07-memory-context-indexing/` (single agent; writes 00-index and 99)
Mints FR/NFR/RISK-MEM/CTX/IDX-*, E-MEM/CTX/IDX-*; ADR 085–099 (use sparingly); ~14k words.
Brief: 486–508, 2128–2167, 2170–2196. Files: `00-index.md`, `01-memory-model.md` (keystone
FR-MEM-001; session/workspace/long-term/semantic/episodic memory, user preferences, project
knowledge, decision memory; provenance, trust, source attribution; MUST NOT store secrets
without explicit authorization), `02-memory-lifecycle.md` (ingestion, normalization,
compression, summarization, expiration, retention, deletion, export, encryption, redaction,
conflict resolution, versioning, offline), `03-context-manager.md` (keystone FR-CTX-001;
sources, priority, ranking, token budgeting, dedup, compression, summarization, freshness,
trust, provenance, conflict detection, user pinning, exclusion, per-model windows/limits,
tool-result handling, large/binary file handling, snapshots, reproducibility),
`04-indexing-engine.md` (keystone FR-IDX-001; chunking, embeddings per ADR-020, incremental
updates, invalidation, operation without embeddings and without Internet),
`05-index-state-machine.md` (FULL machine), `99-volume-register.md`.

### Volume 8 — `volume-08-cli-and-tui/`
- **Agent 8A** (writes 00-index) — mints FR/NFR/RISK-CLI-*, UX-001–039, E-CLI-*; ADR 100–104;
  ~14k words. Brief: 510–547, 2387–2439. Files: `00-index.md`, `01-cli-architecture.md`
  (keystone FR-CLI-001), `02-cli-conventions.md` (argument/flag grammar, global flags, exit
  codes per ADR-016, stdout/stderr discipline, human + `--json` output for EVERY command,
  quiet/verbose/debug, non-interactive `--yes`/`--no-input`, CI mode, confirmation behavior,
  `ANDROMEDA_*` env vars, shell completion), `03-cli-commands-core.md` (`andromeda`, `run`,
  `plan`, `exec`, `init`, `config`, `auth`), `04-cli-commands-platform.md` (`provider`,
  `model`, `tool`, `plugin`, `skill`, `workflow`, `mcp`), `05-cli-commands-data.md` (`memory`,
  `context`, `index`, `git`, `logs`, `trace`, `export`), `06-cli-commands-maintenance.md`
  (`doctor`, `update`, `version`, `completion`) — EVERY command: syntax, args, flags,
  defaults, examples (valid + invalid), errors, exit codes, JSON schema of output,
  permissions touched — `98-register-a.md`.
- **Agent 8B** — mints TUI-001–059, UX-040–069, E-TUI-*; ADR 105–109; ~12k words. Brief:
  510–547, 2319–2384. Files: `07-tui-architecture.md` (keystone FR-TUI-001; Bubble Tea v2
  model per ADR-006, panel system, navigation, focus, keyboard/mouse, resize, small-terminal
  ≥80×24 behavior), `08-theming-and-design-tokens.md` (ADR-026 tokens→ANSI mapping table,
  truecolor/256/16/no-color tiers, light-terminal fallback, FIX the Danger red with WCAG-style
  contrast ratios ≥4.5:1 on Neutral, `[tui.theme]` keys), `09-wireframes-core.md` (ASCII
  wireframes ~80×24, each with prose: start/splash with cat mascot + wordmark + tagline,
  workspace selection, session, plan, execution, tool call, permission prompt, diff, git,
  files, context, memory, logs, costs/tokens), `98-register-b.md`.
- **Agent 8C** — mints TUI-060+, UX-070+; ADR 110–114; ~11k words. Brief: 2319–2384. Files:
  `10-wireframes-platform.md` (provider, model, configuration, plugins, skills, workflows,
  MCP, errors, help, command palette, quick actions, update, recovery — ASCII + prose),
  `11-interaction-patterns.md` (streaming rendering, spinners, progress bars, modals,
  confirmations, toasts, empty/loading/offline/degraded states, copy/paste, search,
  filtering, pagination, virtualization), `12-accessibility-and-compatibility.md`
  (accessibility, no-color mode, Unicode fallback, screen readers where viable, SSH, CI,
  terminal compatibility matrix), `98-register-c.md`.

### Volume 9 — `volume-09-security/`
- **Agent 9A** — mints RISK-SEC-001+ (threats) ONLY, no FR/NFR; ADR 115–120 (only if truly
  needed); ~13k words. Brief: 549–582, 2442–2492. Files: `01-threat-model-overview.md`
  (assets, trust boundaries, actors, attack vectors, risk matrix), `02-threats-injection.md`,
  `03-threats-extensions-supply-chain.md`, `04-threats-system.md` — the 27 mandated threats
  (prompt injection, indirect prompt injection, tool injection, tool poisoning, MCP poisoning,
  malicious plugins, malicious skills, malicious repositories, malicious files, command
  injection, path traversal, symlink attacks, secret exfiltration, credential theft,
  dependency attacks, compromised providers, compromised local models, malicious model
  output, sandbox escape, privilege escalation, supply chain, CI compromise, release
  compromise, update compromise, log leakage, memory poisoning, index poisoning, social
  engineering — group the extras sensibly), EACH as a `### RISK-SEC-NNN — Name` with the full
  RISK template PLUS subsections: Asset, Actor, Vector, Preconditions, Impact, Prevention,
  Response, Recovery, Residual risk, Tests. `98-register-a.md`.
- **Agent 9B** (writes 00-index) — mints ALL FR/NFR-SEC-* (from SEC-100 up for FRs), E-SEC-*;
  ADR 121–129; ~12k words. Brief: 1895–1938, 1941–1974. Files: `00-index.md`,
  `05-permission-model.md` (keystone FR-SEC-100; formalizes §5 enums; inheritance,
  precedence, revocation, persistence, audit), `06-sandbox-specification.md` (keystone
  FR-SEC-101; per ADR-021 layers; filesystem isolation, working dir, ro/rw mounts, network
  policy, process/CPU/memory/time limits, env filtering, secret filtering, command
  allow/denylists, child process control, symlink handling, temp dirs, cleanup, audit;
  sandbox tiers for process/tool/workflow/plugin/MCP server), `07-credential-and-secret-
  management.md` (keystone FR-SEC-102; ADR-014; encryption, redaction in logs/errors/memory),
  `08-audit-and-incident-response.md` (Audit Log semantics, security events, recovery,
  incident response, disclosure policy pointer to Vol 15 governance), `09-approval-state-
  machine.md` (FULL machine), `98-register-b.md`.

### Volume 10 — `volume-10-config-storage-observability/`
- **Agent 10A** (writes 00-index) — mints FR/NFR/RISK-CFG-*, E-CFG-*; ADR 130–136; ~11k
  words. Brief: 584–616, 2199–2230. Files: `00-index.md`, `01-configuration-model.md`
  (keystone FR-CFG-001; COMPLETE `andromeda.toml` example covering every table from Volume 0
  ch 03; global/project/workspace configs, profiles, env-var mapping algorithm, CLI/runtime
  overrides, includes, defaults, validation per ADR-024, schema versioning, migrations,
  deprecations, EXACT precedence order: CLI flag > env var > runtime override > project >
  workspace > profile > global > defaults — adjust and justify if you find a better order),
  `02-config-errors-and-redaction.md` (minimal example, invalid examples fenced `toml
  invalid`, exact error messages with E-CFG codes, secret redaction rules),
  `98-register-a.md`.
- **Agent 10B** — mints FR/NFR/RISK-OBS-*, E-OBS-*; ADR 137–144; ~12k words. Brief: 2495–2541.
  Files: `03-logging.md` (slog JSON per ADR-011, levels, rotation, redaction, offline-first),
  `04-events-and-envelope.md` (keystone FR-OBS-001; the event envelope: name, version,
  producer, consumers, payload, correlation/session/run/tool-call/provider-request IDs,
  timestamp, ordering, delivery semantics, persistence, retention, privacy, redaction,
  compatibility, failure behavior), `05-traces-metrics-costs.md` (OTel mapping, latency/token/
  cost/retry/rate-limit metrics, file changes, commands, git changes, memory ops, indexing,
  security events), `06-telemetry-and-consent.md` (local observability MUST work offline;
  remote telemetry strictly separated, disabled by default pending an explicit ADR-level
  consent decision you mint; collected vs prohibited data, redaction, retention, export,
  deletion, enterprise policy, OTLP), `98-register-b.md`.

### Volume 11 — `volume-11-git-and-github/` (single agent; writes 00-index and 99)
Mints FR/NFR/RISK-GIT-*, -GH-*, E-GIT/GH-*; ADR 145–159; ~15k words. Brief: 618–642,
1213–1241, 1244–1292, 1295–1337, 1340–1375, 1378–1437, 1440–1491, 2233–2272, 2275–2316,
2669–2700 (commit conventions). Files: `00-index.md`, `01-git-engine.md` (keystone
FR-GIT-001; ADR-025; repository discovery, status, diff, staging, commits, branches, tags,
remotes, fetch/pull/push, rebase, merge, cherry-pick, revert, reset, blame, log, worktrees,
hooks, conflicts, submodules, sparse checkout, LFS, ignore rules, signing, protected
branches — permissions + confirmations; NO silent destructive operations),
`02-github-gitlab-product-integrations.md` (product-side integrations; PRs/MRs),
`03-repository-structure-and-branching.md` (the andromeda repo: full tree per ADR-003/031,
ADR-004 branching, CODEOWNERS, templates, security policy, contributing, changelog,
governance), `04-pull-requests.md` (template, size limits, required checks, human review
MANDATORY, no self-approval, AI-generated changes labeled at PR level — commit messages carry
change info ONLY, never AI/vendor attribution per ADR-015; a commit-msg hook + CI check
enforce this), `05-issues-projects-roadmap.md` (issue types, labels taxonomy, GitHub Projects
fields/views/automation, milestones), `06-github-actions.md` (pipelines: format, lint,
compile, unit, integration, E2E, macOS+Linux arch matrix, CodeQL, dependency audit, secret
scanning, coverage, benchmarks, packaging, signing, SBOM, provenance, release, docs,
conventional-commit check, traceability check, license compliance, upgrade tests — YAML
sketches with minimal permissions, pinned action versions, fork PRs treated as untrusted),
`07-traceability-automation.md` (keystone FR-GH-001; the full chain objective→artifact with
automated validations), `99-volume-register.md`.

### Volume 12 — `volume-12-performance-and-reliability/` (single agent; writes 00-index, 99)
Mints FR/NFR/RISK-PERF-*, E-PERF-*; ADR 160–174 (sparingly); ~10k words. Brief: 644–666,
2544–2580. Files: `00-index.md`, `01-performance-targets.md` (NFR-PERF-* for: cold start,
warm start, TUI render, input latency, first token, streaming update, tool dispatch,
filesystem scan, git status, indexing, memory retrieval, search, patch generation, diff
rendering, session restore, RAM, CPU, disk, concurrency, large repos ≥100k files, large
files, long sessions — EACH with metric, percentile, reference hardware (define two reference
machines), dataset, threshold, measurement method, phase; align with Volume 1 ch 06 metrics),
`02-reliability-and-degradation.md` (backpressure, timeouts, recovery, resilience,
availability, degraded modes), `03-benchmarks-and-operational-limits.md` (benchmark suite
definition, limits), `99-volume-register.md`.

### Volume 13 — `volume-13-testing-and-quality/` (single agent; writes 00-index, 99)
Mints FR/NFR/RISK-TEST-*, E-TEST-*; ADR 175–189 (sparingly); ~11k words. Brief: 668–695,
2583–2637. Files: `00-index.md`, `01-strategy-and-pyramid.md`, `02-test-types-catalog.md`
(unit, integration, E2E, golden, snapshot, property-based, fuzzing, contract, conformance,
security, acceptance, regression, smoke, performance, load, stress, soak, compatibility,
offline, migration, upgrade, rollback, release, provider-contract, tool-contract,
MCP-conformance, CLI, TUI, git, sandbox, permissions — each: scope, tools per ADR-017,
gates, phase), `03-fixtures-fakes-determinism.md` (fixtures, mocks, fakes, emulators, test
providers, determinism, flaky policy, coverage thresholds, mutation testing, test data,
secret handling, CI parallelization), `04-release-qualification-and-gates.md`,
`99-volume-register.md`.

### Volume 14 — `volume-14-distribution/` (single agent; writes 00-index, 99)
Mints FR/NFR/RISK-REL-*, E-REL-*; ADR 190–204 (sparingly); ~11k words. Brief: 697–720,
2640–2666, 2669–2700. Files: `00-index.md`, `01-distribution-channels.md` (keystone
FR-REL-001; binaries incl. universal macOS when applicable, Homebrew tap, GitHub Releases,
shell installer, checksums, cosign signatures, notarization PENDING VALIDATION, SBOM,
provenance, package formats, air-gapped installation, offline update), `02-updater-and-
rollback.md` (channels stable/beta/nightly/rc, automatic/manual, rollback, UpdaterPort
semantics), `03-install-uninstall-data.md` (data preservation/removal, cleanup),
`04-versioning-support-backports.md` (SemVer application per ADR-015, support windows,
backports, release branches, changelog), `05-state-machines.md` (FULL: Update, Release),
`99-volume-register.md`.

## 9. Completion protocol (every agent)

1. Write your files (one Write per file; do not re-read what you wrote; do not touch anything
   outside your list).
2. Run `cd /Users/maia/Documents/lyra/andromeda && python3 scripts/spec_lint.py --errors-only`.
   Fix every error attributable to YOUR files. Ignore all other findings — other agents are
   writing concurrently. Note: `register-absent` errors for your volume are expected in
   multi-agent volumes (the orchestrator merges fragments later); do not create
   `99-volume-register.md` there.
3. Do not run git commands. The orchestrator commits.
4. Report per the structured output schema you were given.
