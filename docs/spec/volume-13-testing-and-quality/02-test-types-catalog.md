# 02 — Test Types Catalog

This chapter is the closed catalog of Andromeda's test types. Every automated test declares
exactly one primary type (FR-TEST-003); the catalog fixes, per type: scope, tooling (ADR-017
stack unless stated), gate tier ([chapter 01](01-strategy-and-pyramid.md)), and phase. Adding
or removing a type is a change to this chapter through the Volume 0 change procedure. Where a
type verifies content owned by another volume, that volume owns *what* is verified; this
volume owns *how* verification is organized, gated, and phased.

## Catalog overview

| Type | Family | Pyramid position | Gate tier | Phase |
|---|---|---|---|---|
| Unit | Foundation | Level 1 | T0 | MVP |
| Integration | Foundation | Level 2 | T0 | MVP |
| End-to-end | Foundation | Level 3 | T0 subset / T1 full | MVP |
| Golden | Deterministic output | Level 1–2 | T0 | MVP |
| Snapshot | Deterministic output | Level 2 | T0 | MVP |
| Property-based | Generative | Orthogonal | T0 (bounded) / T2 extended | MVP |
| Fuzz | Generative | Orthogonal | T1 corpus replay / T2 extended | MVP |
| Contract | Contract and conformance | Level 2 | T0 | MVP |
| Provider contract | Contract and conformance | Level 2 | T0 (fakes) / T2 (live) | MVP |
| Tool contract | Contract and conformance | Level 2 | T0 | MVP |
| MCP conformance | Contract and conformance | Level 2 | T0 (emulator) / T2 (interop) | Beta |
| Conformance | Contract and conformance | Level 2 | T0 / T2 | MVP |
| CLI | Surface | Level 2 | T0 | MVP |
| TUI | Surface | Level 2 | T0 subset / T1 full | MVP |
| Git | Surface | Level 2 | T0 | MVP |
| Sandbox | Surface | Level 2 | T0 | MVP |
| Permission | Surface | Level 2 | T0 | MVP |
| Security | Guarantee | Orthogonal | T0 scans / T3 | MVP |
| Offline | Guarantee | Orthogonal | T1 / T3 | MVP |
| Compatibility | Guarantee | Orthogonal | T3 | MVP |
| Migration | Lifecycle | Orthogonal | T1 / T3 | MVP |
| Upgrade | Lifecycle | Orthogonal | T3 | MVP |
| Rollback | Lifecycle | Orthogonal | T3 | Beta |
| Release qualification | Lifecycle | Orthogonal | T3 | MVP |
| Performance | Operational | Orthogonal | T2 trend / T3 gate | MVP |
| Load | Operational | Orthogonal | T2 | Beta |
| Stress | Operational | Orthogonal | T2 | Beta |
| Soak | Operational | Orthogonal | T2 (weekly) | Beta |
| Chaos / fault injection | Operational | Orthogonal | T2 / T3 (crash suite) | Beta |
| Acceptance | Process | Level 3 | T3 / T4 | MVP |
| Regression | Process | Any level | T0 | MVP |
| Smoke | Process | Level 3 | T0 / per artifact | MVP |

## Foundation types

### Unit tests

- Scope: One package in isolation; fakes at port boundaries; no filesystem outside
  `t.TempDir()`, no network, no subprocesses.
- Tooling: stdlib `testing`; go-cmp for structural comparison; testify for simple assertions.

### Integration tests

- Scope: Components crossing a real boundary: SQLite (modernc, WAL, ADR-007) in temp dirs,
  real filesystem, real system git ≥ 2.40 (ADR-025), real subprocesses through the Sandbox
  Engine. No network; provider boundaries use the ADR-176 loopback HTTP fakes.
- Tooling: stdlib `testing` plus chapter 03 harness helpers.

### End-to-end tests

- Scope: User journeys driving the compiled `andromeda` binary via CLI and PTY, asserting on
  `--json` output, exit codes (ADR-016), and persisted state. The main journey is UC-01
  (MVP minimum item 25); providers are scripted HTTP fakes.
- Tooling: binary harness over stdlib `os/exec` and PTY; golden JSON assertions.

## Deterministic-output types

### Golden tests

- Scope: Byte-exact comparison against versioned golden files: CLI `--json` documents,
  rendered diffs, export formats, error envelopes. Updates only via an explicit `-update`
  flag recorded in the diff.
- Tooling: go-cmp with golden helpers; goldens in `testdata/`.

### Snapshot tests

- Scope: Rendered TUI frames at fixed sizes (including the 80×24 floor, Volume 8) and color
  tiers, compared against versioned snapshots.
- Tooling: teatest/v2 frame capture (ADR-017).

## Generative types

### Property-based tests

- Scope: Invariants over generated inputs for parsers, codecs, and pure domain logic:
  configuration round-trips (ADR-008), Andromeda Runtime Protocol frames (ADR-009), schema
  validation (ADR-024), ULID ordering (ADR-027), context budgeting arithmetic.
- Tooling: rapid, bounded generators, replay seeds committed on failure.
- Gate: bounded case counts at T0; extended at T2.

### Fuzz tests

- Scope: Crash/panic exploration of untrusted inputs: configuration, protocol frames,
  provider payloads, schema documents, MCP messages, patch documents. Crashers become
  committed corpus entries replayed forever.
- Tooling: native `go test -fuzz`; corpora in `testdata/fuzz/`.
- Gate: corpus replay at T1; ≥ 15 minutes per target at T2.
- Phase: MVP for configuration/protocol surfaces; Beta for the full list.

## Contract and conformance types

### Contract tests

- Scope: Every implementation of the 18 frozen ports verified against its reusable contract
  kit (FR-TEST-004): method semantics per the contract-owner volume, error-family mapping,
  cancellation per FR-ARCH-004, stream discipline, concurrency claims.

### Provider contract tests

- Scope: Each adapter against the ProviderPort kit plus its Volume 5 declaration: capability
  honesty (undeclared capabilities refused, never simulated), streaming, token/cost
  accounting, error normalization.
- Tooling: kit over ADR-176 fakes and recordings; live lanes at T2, non-gating.
- Phase: MVP for the MVP provider seed; later adapters per Volume 5.

### Tool contract tests

- Scope: Every tool (built-in, plugin, MCP-bridged) against the ToolPort kit and its own
  declaration (Volume 6): schema validity (ADR-024), permission declaration enforcement,
  timeout/cancellation, exactly-one-terminal-event stream discipline.

### MCP conformance tests

- Scope: MCP Runtime against declared protocol revisions (ADR-010): negotiation,
  tools/resources/prompts, lifecycle, errors, timeouts; suite content per Volume 6. Interop
  lane: connect/discover/invoke against the SM-15 public reference set.
- Tooling: in-repo MCP test server on the official Go SDK; interop harness at T2.
- Phase: Beta, or earlier if Volume 6 assigns MCP support an earlier phase.

### Conformance tests

- Scope: Externally defined specifications other than MCP: OpenAI-compatible wire shape,
  JSON-RPC 2.0 framing (ADR-009, ADR-012), JSON Schema behavior (ADR-024), and the SM-04
  local-model suite (content per Volume 5) at T2 on reference hardware.

## Surface types

### CLI tests

- Scope: Every command's Volume 8 contract: grammar, defaults, exit codes, stdout/stderr
  discipline, `--json` schema (golden), non-interactive behavior, invalid-usage cases
  (exit code 2).
- Tooling: in-process cobra harness plus compiled-binary checks.

### TUI tests

- Scope: Model-update-view behavior per Volume 8: navigation, focus, resize, streaming
  rendering, permission prompts, degraded states. Snapshots cover appearance; TUI tests
  cover behavior.
- Tooling: teatest/v2 over bubbletea v2 (ADR-006, ADR-017).

### Git tests

- Scope: Git Engine over real temp repositories: porcelain parsing fidelity, version-floor
  gating (2.40 lane runs the floor exactly), permissioned mutations, no-silent-destructive
  rules (Volume 11), conflicts, worktrees.
- Tooling: system git in hermetic env (pinned identities, dates, `GIT_CONFIG_*`).

### Sandbox tests

- Scope: Sandbox Engine enforcement per Volume 9 / ADR-021: filesystem scope, env filtering,
  resource/time limits, process-tree termination on cancellation, and negative escape
  attempts (traversal, symlink escape, out-of-policy spawn) refused and audited.
- Phase: MVP process-level controls; Beta additions per ADR-021's OS layer.

### Permission tests

- Scope: Permission Manager semantics per Volume 9: every side-effecting path holds a
  decision before acting (SM-16b enforcement tests attempt unmediated side effects and MUST
  observe refusal), decision persistence and audit records, scope precedence,
  non-interactive non-allow-is-denial.
- Phase: MVP (SM-16b binds at MVP exit).

## Guarantee types

### Security tests

- Scope: Injection corpora against prompt/tool boundaries (threats per Volume 9),
  secret-redaction assertions in logs/errors/events, dependency and code scanning (CodeQL,
  secret scanning — pipelines per Volume 11), SM-16(a) release gating.
- Gate: scans at T0; full suite at T3.

### Offline tests

- Scope: The offline guarantee list plus the UC-09 journey under OS-level isolation with a
  network sentinel (FR-TEST-005; SM-05).
- Gate: T1 on Linux; full Tier 1 matrix at T3.

### Compatibility tests

- Scope: Identical behavior across Tier 1 platforms (SM-17), terminal matrix (Volume 8),
  system-git version range, SQLite file interchange across platforms (ADR-007, ADR-028).
- Phase: MVP for Tier 1; Beta for the extended terminal matrix.

## Lifecycle types

### Migration tests

- Scope: Forward-only migrations (ADR-029): fixture databases at every released schema
  version migrate with integrity verified; pre-migration backup exists; future schemas
  refused with exit code 9; backup-restore recovery works.

### Upgrade tests

- Scope: Release N−1 → N through UpdaterPort (check, download, verify, apply) against a
  local release-metadata fixture server; SM-18 timing; data preserved per Volume 14.
- Phase: MVP (first measurable at the second release).

### Rollback tests

- Scope: Restore of the previous version from locally retained artifacts, offline (SM-19);
  state integrity after rollback. Phase: Beta (semantics per Volume 14).

### Release qualification tests

- Scope: The chapter 04 pipeline itself: artifact verification (checksums, SBOM, provenance,
  signatures when enabled per ADR-013 and the Volume 1 signing viability note),
  install/uninstall on Tier 1 platforms, evidence-bundle completeness.

## Operational types

### Performance tests

- Scope: The benchmark suite measuring Volume 12's NFR-PERF targets on Volume 12's reference
  environments; SM-06..SM-09 trends from MVP, gating per Volume 12's phase bindings.
- Tooling: `go test -bench` plus the Volume 12 harness; results as evidence artifacts.

### Load tests

- Scope: Sustained operation at Volume 12's declared concurrency limits (sessions, tool
  invocations, index builds) verifying throughput and bounded queues (ADR-023).
- Tooling: scripted driver against the headless surface (ADR-032) and CLI.

### Stress tests

- Scope: Behavior beyond declared limits: scheduler rejection and backpressure (ADR-023),
  rate-limit storms, disk pressure — explicit degradation per Volume 12, never silent
  corruption.

### Soak tests

- Scope: The 8-hour scripted session of SM-09(b): memory growth bounds, goroutine and
  file-descriptor leak detection, log rotation, database growth.

### Chaos / fault-injection tests

- Scope: Crash injection at randomized points (kill −9, SIGKILL during tool execution,
  simulated power loss between writes) verifying SM-11 recovery via `MarkInterrupted`
  semantics: work marked `interrupted`, resume correctness, zero loss of persisted turns.
  Port-level fault injection (latency, errors, malformed data) verifying degradation paths.
- Gate: T2; the SM-11 crash suite also runs at T3. Phase: Beta (crash suite measurable from
  MVP).

## Process types

### Acceptance tests

- Scope: Executable Given/When/Then scenarios derived from requirement acceptance criteria,
  annotated per FR-TEST-002; the full suite gates phase exits and releases on every Tier 1
  platform (SM-17).

### Regression tests

- Scope: One reproducing test per fixed defect, added in the fixing pull request, named for
  the issue, annotated; placed at the lowest expressive level (FR-TEST-001).

### Smoke tests

- Scope: Minimal sanity per built artifact: binary starts, `version` and `doctor` succeed, a
  trivial fake-provider run completes, exit codes sane. Runs per packaged artifact per Tier 1
  platform and post-publication.

## Requirements

### FR-TEST-003 — Closed test-type catalog and classification

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: Testing and quality (Volume 13)
- Affected components: all components; CI pipelines (Volume 11)
- Dependencies: FR-TEST-001; chapter 01 tiers
- Related risks: RISK-TEST-001

#### Description

Every automated test MUST be classifiable to exactly one primary catalog type, expressed
through package placement and naming conventions documented in the repository testing guide,
so type and tier (T0–T4) are derivable mechanically from the tree. The catalog is closed:
a test fitting no type requires amending this chapter first. A classification report (types,
counts, tiers) is generated on trunk runs.

#### Motivation

Gate composition (chapter 04) selects suites by type and tier; without mechanical
classification, gate contents drift into hand-maintained CI lists that rot.

#### Actors

Test authors; CI pipeline definitions; gate evaluators.

#### Preconditions

Placement conventions documented (Volume 11 repository structure).

#### Main flow

1. A contributor places a new test per its type's convention.
2. CI derives type and tier and includes it in the right suites.
3. The classification report updates on trunk.

#### Alternative flows

- A test spans mechanisms (a golden assertion inside a CLI test): the primary type is the
  surface verified (CLI); the mechanism is a detail.

#### Edge cases

- Quarantined tests keep their type; exclusion is via the `quarantine` tag (ADR-177), not
  reclassification. Fuzz targets remain fuzz in replay-only mode.

#### Inputs

Test sources; placement conventions.

#### Outputs

Classification report (JSON) as a CI artifact.

#### States

Not applicable — repository artifact governance.

#### Errors

Unclassifiable tests fail the T0 classification check; report-generation failure is
E-TEST-006.

#### Constraints

Classification derives from the tree alone — no per-test registry file.

#### Security

None; the report is public repository data.

#### Observability

Report published per trunk run; referenced by the chapter 04 evidence bundle.

#### Performance

Static analysis within the merge gate's lint allocation.

#### Compatibility

Conventions are platform-neutral.

#### Acceptance criteria

- Given the full test tree, when the classification check runs, then 100% of tests resolve
  to exactly one type and tier.
- Negative case: given a test placed outside every convention, when T0 runs, then the check
  fails naming the file.
- Negative case: given a new type used without a catalog amendment, when reviewed and
  checked, then the change is rejected.
- Observability case: the trunk report is retrievable and matches executed suite
  composition.

#### Verification method

Classification check self-tests; T4 audit comparing report against executed CI jobs.

#### Traceability

FR-TEST-001; chapter 04 gate composition; Volume 11 pipeline definitions (by name).

### FR-TEST-004 — Port contract test kits

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Core
- Source: Design
- Owner: Testing and quality (Volume 13)
- Affected components: all port implementations; Extension SDK (kit distribution per Volume 6)
- Dependencies: FR-ARCH-003, FR-ARCH-004; ADR-017, ADR-030
- Related risks: RISK-TEST-004

#### Description

For each of the 18 frozen ports (Volume 3, chapter 02) there MUST exist exactly one reusable
contract test kit: a package exporting a function that accepts a port implementation (plus a
capability description where the contract is conditional) and runs the behavioral contract as
subtests — method semantics per the contract-owner volume, error-family mapping (no raw
driver/HTTP/OS errors), cancellation and pre-cancelled-context behavior for every method,
stream discipline (ordering, terminal events, idempotent `Close`), declared concurrency
safety. Every in-repo implementation — real adapters *and* the chapter 03 fakes — MUST pass
its kit in T0. Kits are mirrored to extension authors through the Extension SDK per Volume 6.

#### Motivation

The port freeze (FR-ARCH-003) is only as real as its enforcement; kits turn contract drift
into a T0 failure, and running the same kit over fakes bounds fake/real divergence
(RISK-TEST-004) — the property ADR-176 depends on.

#### Actors

Port implementers (in-repo and extension authors); kit maintainers; CI.

#### Preconditions

The port's contract-owner volume has specified behavioral semantics.

#### Main flow

1. An implementer wires their implementation into the kit entry point in a test.
2. `go test` runs the kit's subtests in T0.
3. Contract changes land as kit changes first; implementations follow.

#### Alternative flows

- A method unsupported in the current phase (Volume 3 phased-presence rule): the kit asserts
  the declared "unsupported" error class instead of skipping.

#### Edge cases

- Conditional surface (a provider without `streaming`): the kit asserts refusal semantics
  for absent capabilities rather than omitting coverage.
- Kits MUST be usable outside the monorepo (SDK consumers): no `internal/` imports beyond
  the ports mirror (ADR-031).

#### Inputs

A port implementation; capability/limit descriptions where applicable.

#### Outputs

Pass/fail subtests with per-method diagnostics.

#### States

Kits exercise the frozen state vocabularies the contract owner defines without minting any.

#### Errors

Kit fixture failures surface as E-TEST-001; contract violations are ordinary test failures
naming method and clause.

#### Constraints

One kit per port; kits use only exported contract types (Volume 3 no-leakage rule).

#### Security

Kits for PermissionPort, SandboxPort, and SecretStorePort encode their owners' mediation and
redaction clauses — an implementation cannot pass while bypassing them.

#### Observability

Kit results are named CI checks per implementation; compliance changes are release-notes
material (NFR-ARCH-002 regime).

#### Performance

A kit run fits T0 budgets; long-running clauses are exported separately for T2.

#### Compatibility

Kits run identically on all Tier 1 platforms; platform-conditional clauses mirror the
Volume 3 platform matrix explicitly.

#### Acceptance criteria

- Given the 18 ports, when the kit inventory check runs, then exactly one kit exists per
  port and every in-repo implementation (including every fake) references its kit.
- Given an adapter leaking a raw underlying error, when the kit runs, then the error-family
  clause fails naming the method.
- Given any method called with a pre-cancelled context, when the kit runs, then the
  implementation returns the family's cancellation class without side effects
  (FR-ARCH-004).
- Negative case: given a fake diverging from a clause, when T0 runs, then the fake fails its
  kit — divergence cannot persist silently.
- Observability case: kit results appear as named CI checks per implementation.

#### Verification method

Kit inventory check in CI; kits in T0; consolidation audit that each kit covers its
contract-owner volume's clauses.

#### Traceability

FR-ARCH-003, FR-ARCH-004, NFR-ARCH-002; ADR-176; SM-01 (the kit is the conformance bar for
new adapters); SM-02.

### FR-TEST-005 — Offline test suite and network sentinel

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Testing and quality (Volume 13)
- Affected components: all components; CI pipelines (Volume 11)
- Dependencies: Volume 1 offline guarantee list; SM-05; ADR-176
- Related risks: RISK-TEST-002

#### Description

The offline suite executes the eleven offline-guaranteed operations (Volume 1, chapter 04)
plus the UC-09 journey with a local provider under two enforced layers: (1) OS-level network
isolation — on Linux a network namespace with no external interfaces; on macOS runners the
mechanism is PENDING VALIDATION (register entry), and until validated the macOS lane runs
layer (2) only, stated explicitly in the report; (2) a network sentinel — test-build
instrumentation at the dialer seam that records and fails on any outbound attempt, counting
attempts per SM-05's "0 network-access attempts observed". Runs at T1 on Linux; full Tier 1
matrix at T3.

#### Motivation

Local-first operation binds at MVP exit (SM-05). Without isolation *and* a sentinel, "no
network attempts" is unfalsifiable: isolation proves nothing was reachable, the sentinel
proves nothing was attempted.

#### Actors

CI pipelines; the offline harness; release qualification.

#### Preconditions

A local serving path is available in CI (Ollama with a pinned model, or the local fake path
where the subject is Andromeda rather than model quality — SM-04 covers model quality).

#### Main flow

1. The harness establishes isolation (Linux) and enables the sentinel.
2. Each guaranteed operation executes and is asserted individually.
3. The UC-09 journey runs end-to-end.
4. The report records per-operation results and the sentinel count.

#### Alternative flows

- An operation legitimately degrades offline (documented per Volume 12): the check asserts
  the *documented* degraded behavior, not silent success.

#### Edge cases

- Loopback traffic to local providers/fakes is allowed; anything else counts. DNS resolution
  attempts count even when they fail fast.

#### Inputs

The guarantee list (generated from Volume 1's list, not hand-copied); local provider
endpoint.

#### Outputs

Offline report (JSON): per-operation outcome, sentinel count, isolation layer used.

#### States

Not applicable — the suite observes product states without minting any.

#### Errors

A sentinel-detected attempt fails the run with E-TEST-004; harness setup failure is
E-TEST-001.

#### Constraints

The suite MUST NOT stub Andromeda's own code paths — the binary is unmodified except for the
test-build sentinel seam.

#### Security

Sentinel records (destination host/port) are CI-internal diagnostics with no user data.

#### Observability

The report joins the evidence bundle; violations emit `test.hermeticity.violated`.

#### Performance

Completes within the T1 budget on Linux; offline *performance* measurement is Volume 12's.

#### Compatibility

Linux isolation uses network namespaces; macOS isolation PENDING VALIDATION; the sentinel is
platform-neutral.

#### Acceptance criteria

- Given the offline environment, when the eleven operations run, then 100% complete per
  documented behavior and the sentinel records 0 non-loopback attempts.
- Negative case: given a change introducing a network call into a guaranteed operation, when
  T1 runs, then the sentinel fails the run with E-TEST-004 naming the destination.
- Error case: given isolation unavailable and no sentinel, when the suite starts, then it
  refuses to report success (E-TEST-001) rather than running unenforced.
- Observability case: the report appears in the evidence bundle with the isolation layer
  recorded.

#### Verification method

The suite at T1/T3; harness self-tests injecting a deliberate network call and asserting
detection; T4 report audit.

#### Traceability

SM-05; UC-09; Volume 1 offline guarantee list; Volume 12 offline NFR formalization (by
name); E-TEST-004.
