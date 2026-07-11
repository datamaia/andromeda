# 03 — Fixtures, Fakes, Determinism, Coverage, and Mutation

This chapter defines the test infrastructure: the test-double taxonomy, the provider doubles
of ADR-176, determinism controls, the flaky-test policy of ADR-177, coverage and mutation
gates, test data, secret handling, and CI parallelization.

## Test-double taxonomy

| Double | Definition | Permitted use |
|---|---|---|
| Fixture | Versioned static input data (files, databases, cassettes, goldens) under `testdata/` | All levels |
| Stub | Canned single-response stand-in with no behavior | Level 1 only, incidental collaborators |
| Mock | Expectation-scripted double asserting on how it is called | Discouraged for ports; permitted only where the *interaction itself* is the contract (for example, verifying PermissionPort is consulted before a side effect) |
| Fake | Working lightweight port implementation with real behavior (in-memory Secret Store, scripted provider) | Preferred double at every port boundary |
| Emulator | Local process imitating an external system's wire protocol (HTTP provider fakes, in-repo MCP test server) | Level 2 and above |
| Test provider | The ADR-176 provider doubles (scripted fake, HTTP fakes, recordings) | Per ADR-176 lanes |

Rules:

1. Fakes at port boundaries MUST pass the same contract kit as real implementations
   (FR-TEST-004) — a fake that cannot pass its kit specifies wishful behavior.
2. Real local infrastructure is used where hermetic and within budget: SQLite in
   `t.TempDir()`, the real filesystem, real system git, real processes through the Sandbox
   Engine. Emulating these is prohibited — fidelity is free.
3. Expectation-scripted mocks stay limited as tabled; behavioral assertions belong on
   outputs and recorded effects, not call sequences.

## Test providers

Per ADR-176: scripted fakes (in-process ProviderPort fake plus loopback HTTP fake servers
for the OpenAI-compatible, Anthropic, and Ollama wire families) as the gating primary;
sanitized recordings for adapter parsing regression; scheduled non-gating live lanes for
drift detection. Scenario scripts (deterministic content, chunk pacing, tool-call sequences,
injected errors, usage figures) are versioned fixtures under E-TEST-001 integrity rules;
invalid scripts are rejected with E-TEST-003; replay divergence raises E-TEST-002.

## Emulators and real-infrastructure policy

| System | Policy |
|---|---|
| SQLite (ADR-007) | Real driver, temp databases, WAL on as in production |
| Filesystem | Real, under `t.TempDir()` |
| Git (ADR-025) | Real system git; hermetic env (pinned identities and dates, `GIT_CONFIG_GLOBAL`/`GIT_CONFIG_SYSTEM` pointed at fixtures); a floor lane pins git 2.40 |
| OS keychain (ADR-014) | Never touched by tests: in-memory SecretStorePort fake (kit-passing); the age fallback exercised with throwaway keys; real keychain only in manual platform verification |
| MCP servers (ADR-010) | In-repo test server on the official Go SDK; public servers only in the T2 interop lane |
| PTY / processes | Real, through SandboxPort/TerminalPort |
| Network | Prohibited outside loopback (FR-TEST-005 sentinel discipline for gating suites) |

## Determinism controls

FR-TEST-007 makes these normative:

1. **Injected time.** Code under test receives a clock interface; tests use a fake clock.
   Real sleeps are prohibited; waiting is event- or fake-clock-driven.
2. **Injected identifiers.** ULID generation (ADR-027) is injectable; seeded sequences make
   persisted artifacts byte-stable for goldens.
3. **Seeded randomness.** All randomness behind injected sources; rapid replay seeds
   committed on failure.
4. **Hermetic environment.** Environment cleared to an allow-list; `TZ=UTC`, pinned locale;
   all paths under `t.TempDir()`.
5. **Order independence.** `-shuffle=on` in the determinism lane; no inter-test state; no
   package-level shared state in test code.
6. **Stable iteration.** No map-order dependence; goldens canonicalized (sorted keys per
   Volume 2's canonical JSON rules).
7. **Concurrency discipline.** Synchronize on events, not timing; goroutine-leak checks at
   suite teardown.

## Flaky-test policy

Per ADR-177: detection via the nightly determinism lane and trunk triage; quarantine by the
`quarantine` build tag with a dated, issue-linked comment; a non-gating quarantine lane keeps
the test running; a 14-day time box enforced by a repository check; no automatic
retry-on-failure in gating CI. Budgets are NFR-TEST-005.

## Coverage

- **Measurement.** `go test -covermode=atomic` with merged profiles on the Linux x86_64
  merge-gate lane, compared against thresholds (SM-14 → NFR-TEST-003).
- **Scopes.** Overall product source, and the strict scope (Core Domain, `internal/ports`,
  `sdk/` contract packages).
- **Exclusions.** Generated code and test helpers, via a versioned, reviewed exclusion list —
  never a CI-console action.
- **Ratchet.** Enforced thresholds never decrease; increases land through the change
  procedure.
- Coverage is a floor, not a goal: mutation testing exists because coverage alone is gameable
  (RISK-TEST-003).

## Mutation testing

Per ADR-175: scoped to Core Domain, contract, and parser/codec packages; weekly T2 lane plus
Beta/v1 phase gates; score = killed ÷ (killed + survived) after recorded equivalent-mutant
exclusions; JSON report as gate evidence; tooling PENDING VALIDATION (register entry) with
fixed selection criteria and a hand-seeded fault-injection fallback. Thresholds are
NFR-TEST-004.

## Test data

1. **Synthetic by construction.** Real user data, third-party repository contents, and
   production captures are prohibited as fixtures.
2. **Reference repository generator.** The scale dataset (Volume 1's reference repository:
   ~5,000 files / ~1,000,000 lines with Git history) comes from a seeded generator; same
   seed, byte-identical output — the corpus itself is not versioned.
3. **Large and binary fixtures.** Anything above 1 MiB is generated at test time or fetched
   by checksum from repository release assets — never committed.
4. **Licensing.** Authored fixtures carry the repository license (ADR-002); third-party
   samples only with a recorded license.
5. **Integrity.** Manifest checksums where corruption would confuse (database fixtures,
   cassettes); mismatches fail fast with E-TEST-001.

## Secret handling in tests

1. Real credentials MUST NOT appear in fixtures, goldens, cassettes, scripts, or test source
   — including revoked ones.
2. Fake credentials use the reserved prefix `andromeda-test-`; secret scanning (Volume 11)
   covers `testdata/` and test source and blocks merges.
3. Recording sanitization is deny-by-default: only allow-listed fields survive (ADR-176);
   the recorder refuses to write a cassette when coverage of auth-bearing fields is
   unproven.
4. Live-lane credentials live only in the CI secret store, scoped to scheduled lanes, never
   exposed to fork-triggered jobs (Volume 11 treats fork PRs as untrusted).
5. Redaction tests use fake secrets injected through SecretStorePort fakes and assert
   absence in logs, errors, events, and persisted records (Volume 9 rules).

## CI parallelization

1. **Sharding unit is the package**, balanced by recorded timing data; the shard map derives
   from FR-TEST-003 classification, never hand-maintained.
2. **Isolation enables parallelism.** Hermetic tests (temp dirs, port-0 loopback listeners,
   no shared state) make `-parallel` and sharding safe; a test that cannot run in parallel
   is a determinism defect.
3. **Platform matrix.** T0: Linux x86_64 plus macOS arm64; T1/T3: the Tier 1 matrix.
   Coverage and distribution accounting come from the Linux x86_64 lane only.
4. **Caching.** Build/module caches keyed by toolchain and dependency digests; test-result
   caching disabled for gating runs — every gate run executes.

## Requirements

### FR-TEST-006 — Test doubles and test providers

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Design
- Owner: Testing and quality (Volume 13)
- Affected components: all components; Provider Layer; Extension SDK (doubles shipped per Volume 6)
- Dependencies: ADR-176; FR-TEST-004
- Related risks: RISK-TEST-004, RISK-TEST-005

#### Description

Test doubles follow this chapter's taxonomy and rules: fakes preferred at port boundaries
and required to pass their port's contract kit; expectation-scripted mocks limited to
interaction-is-the-contract cases; real local infrastructure per the policy table; provider
testing through the three ADR-176 lanes with versioned scenario scripts. The ProviderPort
fake, the loopback HTTP fake servers, the in-memory SecretStorePort fake, and the MCP test
server are first-class packages with owners, kit compliance, and their own unit tests.

#### Motivation

Double quality bounds suite meaning: an unfaithful fake verifies fiction (RISK-TEST-004).
Kit-passing fakes and a specified three-lane provider system turn faithfulness into a
checked property.

#### Actors

Test authors; double maintainers; extension authors consuming SDK-shipped doubles.

#### Preconditions

Contract kits exist (FR-TEST-004); the scenario-script format is versioned.

#### Main flow

1. A test takes the standard fake for a port collaborator.
2. The fake's own T0 kit run guarantees contract compliance.
3. Provider-dependent tests select the ADR-176 lane.

#### Alternative flows

- No standard fake exists for a port yet: the author contributes one (kit-passing) rather
  than hand-rolling a stub for a non-incidental collaborator.

#### Edge cases

- Fault injection: fakes expose scripted failure modes (latency, errors, malformed payloads)
  for chaos/stress suites, using the port's declared error families only.
- Recordings that no longer replay fail with E-TEST-002 and route to re-recording, not to
  loosened assertions.

#### Inputs

Port contracts; scenario scripts; cassettes.

#### Outputs

Deterministic collaborator behavior; divergence diagnostics.

#### States

Fakes honor frozen state vocabularies where their port exposes them, minting none.

#### Errors

E-TEST-001 (fixture integrity), E-TEST-002 (replay divergence), E-TEST-003 (script
rejected).

#### Constraints

Doubles are test-only code, never importable from shipped binaries (ADR-033 checks).

#### Security

Doubles never contain real credentials; the Secret Store fake zeroizes like the real one so
redaction tests are honest.

#### Observability

Double failures carry the E-TEST envelope and emit their telemetry events (chapter 04).

#### Performance

Fakes fit Level 1 budgets; scripted pacing is fake-clock driven, adding no wall time.

#### Compatibility

Doubles are platform-neutral; platform-specific behavior is real infrastructure's job per
the policy table.

#### Acceptance criteria

- Given every standard fake, when T0 runs, then each passes its port's contract kit.
- Error case: given a scenario script with an unknown directive, when loaded, then the fake
  rejects it with E-TEST-003 naming the directive; nothing runs against a partial script.
- Negative case: given a shipped-binary build, when the ADR-033 dependency check runs, then
  no test-double package appears in the import graph.
- Error case: given a stale cassette, when the parsing regression suite runs, then it fails
  with E-TEST-002 including the first divergent frame.
- Observability case: divergences emit `test.replay.diverged` correlated to the suite run.

#### Verification method

Contract kits over fakes at T0; doubles' own unit tests; ADR-033 import-graph check;
scheduled live-lane cross-checks (ADR-176).

#### Traceability

ADR-176; FR-TEST-004; RISK-TEST-004; SM-04, SM-10.

### FR-TEST-007 — Determinism controls and flaky-test quarantine

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Design
- Owner: Testing and quality (Volume 13)
- Affected components: all components (test suites); CI pipelines (Volume 11)
- Dependencies: ADR-177; NFR-TEST-002, NFR-TEST-005
- Related risks: RISK-TEST-002

#### Description

Gating tests MUST satisfy the seven determinism controls of this chapter. Flaky tests are
handled exclusively through the ADR-177 mechanism: build-tag quarantine with dated,
issue-linked comments, a non-gating quarantine lane, a 14-day time box enforced by a
repository check, and no automatic retry-on-failure in gating CI. The nightly determinism
lane (repeat + shuffle) is the detection instrument (NFR-TEST-002).

#### Motivation

A gate that sometimes lies trains contributors to re-run instead of read (RISK-TEST-002).
Determinism is enforceable only as concrete controls plus a bounded, mechanical response.

#### Actors

Test authors; the determinism lane; triagers; the quarantine check.

#### Preconditions

Harness helpers (fake clock, env allow-list, leak checks) exist as shared packages.

#### Main flow

1. Tests are written against the controls; review rejects violations (real sleeps,
   environment reads, wall-clock assertions).
2. The nightly lane repeats and shuffles; divergence flags a flake.
3. Quarantine lands by pull request within one working day; the time box runs.

#### Alternative flows

- The flake reproduces a real product race: the issue converts to a product defect; the test
  returns only with the fix.

#### Edge cases

- Timeout behavior tests through the fake clock; behavior inexpressible deterministically is
  redesigned or moved to a non-gating operational suite, never merged flaky.
- Quarantining the last test verifying a requirement triggers the FR-TEST-002 uncovered
  finding at the next gate.

#### Inputs

Test sources; lane outcome sets; quarantine annotations.

#### Outputs

Lane reports; the generated quarantine registry; dwell-time metrics.

#### States

Not applicable — process machinery.

#### Errors

Quarantine-check violations (expired date, missing issue link) fail the merge gate; lane
infrastructure failure is E-TEST-006.

#### Constraints

No CI-level rerun-until-green in T0/T1/T3; bounded fake-clock polling is the only sanctioned
wait.

#### Security

None specific; quarantine metadata is public.

#### Observability

Quarantines emit `test.flake.quarantined`; lane reports and the registry feed NFR-TEST-005.

#### Performance

The 50-repeat lane runs at T2, outside gating budgets.

#### Compatibility

Controls are platform-neutral; platform behavioral differences belong to compatibility
tests, not tolerated nondeterminism.

#### Acceptance criteria

- Given the unit and integration suites, when the nightly lane runs 50 shuffled repeats,
  then outcome sets are identical (NFR-TEST-002).
- Negative case: given a test using a real 2-second sleep, when reviewed and linted, then
  the change is rejected per control 1.
- Negative case: given a quarantined test older than 14 days, when any pull request runs T0,
  then the quarantine check fails the gate.
- Error case: given a gating CI job configured with retry-on-failure, when the pipeline
  definition check (Volume 11) runs, then the configuration is rejected.
- Observability case: each quarantine appears in the registry with issue link and date, and
  `test.flake.quarantined` is emitted.

#### Verification method

Nightly determinism lane; quarantine repository check; pipeline-definition review;
NFR-TEST-005 accounting at T4.

#### Traceability

ADR-177; NFR-TEST-002, NFR-TEST-005; RISK-TEST-002.

### FR-TEST-008 — Test data and secret handling

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Design
- Owner: Testing and quality (Volume 13)
- Affected components: all components (test suites); CI pipelines (Volume 11)
- Dependencies: ADR-176 (sanitization); ADR-002 (licensing)
- Related risks: RISK-TEST-005

#### Description

Test data follows this chapter's five data rules (synthetic-only, seeded reference
generator, 1 MiB committed-fixture ceiling, licensing recorded, manifest checksums); secrets
follow the five secret rules (no real credentials, `andromeda-test-` prefix,
deny-by-default sanitization, fork-isolated live credentials, redaction tests via injected
fakes). Secret scanning covers all test trees and blocks merges.

#### Motivation

Fixtures are the least-reviewed, longest-lived files in a repository — where leaked
credentials and unlicensed content hide (RISK-TEST-005).

#### Actors

Test authors; the recorder/sanitizer; secret scanning; reviewers.

#### Preconditions

Scanner configured over test trees; generator packages exist.

#### Main flow

1. An author reaches for a generator or authors a synthetic fixture.
2. Recordings pass through the sanitizer at capture time.
3. Scanning validates every pull request.

#### Alternative flows

- A third-party sample is genuinely needed: admitted with its license recorded alongside;
  scanning still applies.

#### Edge cases

- Goldens containing error envelopes show redacted placeholders, proving redaction.
- The seeded generator's output is regenerable: cache loss is never data loss.

#### Inputs

Generators, seeds, sanitizer allow-lists.

#### Outputs

Fixtures, manifests, scan reports.

#### States

Not applicable.

#### Errors

Manifest mismatch fails with E-TEST-001; sanitizer refusal blocks cassette creation.

#### Constraints

The 1 MiB ceiling; exceptions require recorded justification in review.

#### Security

This requirement is itself a security control; a found real secret triggers Volume 9
incident handling (rotate, then purge history per policy).

#### Observability

Scan results are required checks; fixture failures emit `test.fixture.failed`.

#### Performance

Generation cost stays inside suite budgets by caching generated corpora per seed within a
run.

#### Compatibility

Generators are platform-neutral; path-length and case-sensitivity edge fixtures are
deliberate compatibility inputs, generated per platform.

#### Acceptance criteria

- Given any fixture tree, when secret scanning runs, then zero findings; a planted
  credential-shaped string without the reserved prefix is detected (scanner self-check).
- Given the reference generator with a fixed seed, when run twice, then outputs are
  byte-identical.
- Negative case: given a pull request committing a 5 MiB binary fixture, when T0 runs, then
  the size check fails absent a recorded exception.
- Error case: given a database fixture with a checksum mismatch, when loaded, then the load
  fails fast with E-TEST-001.
- Permission case: given a fork pull request, when its CI jobs run, then live-lane
  credentials are not exposed to them.
- Observability case: fixture failures emit `test.fixture.failed` with the manifest path.

#### Verification method

Secret-scanning checks; generator determinism tests; size checks in T0; sanitizer unit tests
with planted-secret corpora.

#### Traceability

RISK-TEST-005; ADR-176; Volume 9 redaction rules (by name); Volume 11 secret scanning (by
name).

### NFR-TEST-003 — Coverage thresholds

- Category: Maintainability
- Priority: P0
- Phase: MVP
- Metric: Line coverage per SM-14's two scopes: overall product source, and the strict scope (Core Domain, `internal/ports`, `sdk/` contract packages), measured per this chapter's Coverage section
- Target: MVP: ≥ 70% overall and ≥ 85% strict scope; v1: ≥ 80% overall and ≥ 90% strict scope
- Minimum threshold: The MVP levels are the floor from MVP exit onward; the ratchet rule forbids decreases
- Measurement method: `go test -covermode=atomic` merged profiles on the Linux x86_64 lane, evaluated as a merge gate against the current floor; per-release report in the evidence bundle
- Test environment: Linux x86_64 CI lane (Volume 11 pipelines)
- Measurement frequency: Every merge gate; reported per release; audited at T4
- Owner: Testing and quality (Volume 13)
- Dependencies: FR-TEST-001; this chapter's exclusion list
- Risks: RISK-TEST-003
- Acceptance criteria: The merge gate blocks below-floor changes; the per-release report shows both scopes at or above the phase levels; coverage measured on a package subset counts as unmeasured per Volume 1 metric governance and blocks dependent gates.

### NFR-TEST-004 — Mutation score on scoped packages

- Category: Maintainability
- Priority: P1
- Phase: Beta
- Metric: Mutation score per ADR-175 (killed ÷ (killed + survived), after recorded equivalent-mutant exclusions), per scoped package and aggregated
- Target: ≥ 75% aggregate at the v1 gate
- Minimum threshold: ≥ 60% aggregate at the Beta gate; no scoped package below 50% at either gate
- Measurement method: The ADR-175 mutation lane's JSON report; thresholds evaluated at phase gates; weekly trend at T2
- Test environment: Linux x86_64 CI lane; tool and version pinned per ADR-175 once validated
- Measurement frequency: Weekly lane; gating evaluation at Beta and v1 gates
- Owner: Testing and quality (Volume 13)
- Dependencies: ADR-175 (tooling PENDING VALIDATION — not evaluable until resolved; see the open-questions register)
- Risks: RISK-TEST-003
- Acceptance criteria: Beta gate evidence includes a mutation report meeting the minimum threshold; v1 gate evidence meets the target; surviving-mutant lists are attached as actionable findings.

### NFR-TEST-005 — Flake rate and quarantine dwell time

- Category: Reliability
- Priority: P1
- Phase: MVP
- Metric: (a) Fraction of gating trunk CI runs over a rolling 4-week window whose failure is reclassified as flake; (b) quarantine dwell time per test; (c) quarantine lane population
- Target: (a) ≤ 0.5%; (b) 100% resolved within 14 days; (c) ≤ 5 tests at any time
- Minimum threshold: (a) ≤ 1%; (b) no quarantine exceeds 14 days (hard, check-enforced); (c) ≤ 10 tests
- Measurement method: CI outcome records with triage labels; the quarantine registry generated from `quarantine` tags with dates; weekly report
- Test environment: All gating CI lanes
- Measurement frequency: Weekly; audited at T4
- Owner: Testing and quality (Volume 13)
- Dependencies: FR-TEST-007; ADR-177
- Risks: RISK-TEST-002
- Acceptance criteria: Weekly reports within targets; breaching any minimum threshold for two consecutive weeks triggers the ADR-177 review condition (determinism-controls re-audit) recorded as an issue.

## Errors

### E-TEST-001 — Fixture integrity failure

- Code: E-TEST-001
- Category: Test infrastructure
- Severity: Error
- User message: "Test fixture failed integrity verification."
- Technical message: Fixture at `<path>` failed manifest checksum or structural validation; expected `<digest>`, found `<digest>`.
- Cause: Corrupted, hand-edited, or incompletely regenerated fixture; harness setup failure.
- Safe context data: Fixture path, manifest path, digests.
- Recoverability: Not recoverable in-run.
- Retry policy: No retry; regenerate from the generator or restore from version control.
- Recommended action: Regenerate with the recorded seed; if the change is intentional, update the manifest in the same change.
- Exit code: 9 (integrity error) when surfaced by harness tooling.
- HTTP mapping: Not applicable (no HTTP surface).
- Telemetry event: `test.fixture.failed`
- Security implications: None; digests and paths only — fixture content is never echoed.

### E-TEST-002 — Recorded interaction replay divergence

- Code: E-TEST-002
- Category: Test infrastructure
- Severity: Error
- User message: "Recorded provider interaction no longer matches adapter behavior."
- Technical message: Replay of cassette `<name>` diverged at frame `<n>`: request did not match the recorded request (sanitized field-level diff attached).
- Cause: Adapter behavior changed, cassette stale, or sanitization altered a match-relevant field.
- Safe context data: Cassette name, frame index, sanitized diff.
- Recoverability: Not recoverable in-run.
- Retry policy: No retry; re-record via the ADR-176 scheduled path after confirming intent.
- Recommended action: Intended change → re-record and review the diff; otherwise fix the regression.
- Exit code: 1 (general error) when surfaced by harness tooling.
- HTTP mapping: Not applicable.
- Telemetry event: `test.replay.diverged`
- Security implications: Diffs are sanitizer output only; raw recorded material never appears.

### E-TEST-003 — Scenario script rejected

- Code: E-TEST-003
- Category: Test infrastructure
- Severity: Error
- User message: "Test provider scenario script is invalid."
- Technical message: Script `<name>` failed validation at `<location>`: unknown directive, malformed pacing, or capability-inconsistent sequence.
- Cause: Script authored against a different script-format version or internally inconsistent (tool-call chunks in a scenario declaring no `tool_calling`).
- Safe context data: Script name, format version, offending directive and location.
- Recoverability: Not recoverable in-run.
- Retry policy: No retry; fix the script.
- Recommended action: Validate against the versioned script schema; regenerate from a template.
- Exit code: 3 (configuration error) when surfaced by harness tooling.
- HTTP mapping: Not applicable.
- Telemetry event: `test.scenario.rejected`
- Security implications: None; scripts are synthetic by rule (FR-TEST-008).

### E-TEST-004 — Hermeticity violation

- Code: E-TEST-004
- Category: Test infrastructure
- Severity: Critical
- User message: "A hermetic test attempted prohibited external access."
- Technical message: Network sentinel observed an outbound attempt to `<host:port>` from `<test>` during a hermetic/offline run; isolation layer in effect: `<layer>`.
- Cause: Product code reaching the network inside an offline-guaranteed operation, or a test violating the loopback-only rule.
- Safe context data: Destination host/port, test identifier, isolation layer, attempt count.
- Recoverability: Not recoverable in-run.
- Retry policy: No retry — the violation is the finding (SM-05 binds at MVP exit).
- Recommended action: Product defect when raised from a guaranteed operation; test defect otherwise.
- Exit code: 1 (general error) when surfaced by harness tooling.
- HTTP mapping: Not applicable.
- Telemetry event: `test.hermeticity.violated`
- Security implications: A guaranteed-operation violation is also a privacy finding (unexpected egress), triaged with Volume 9's severity lens.

## Risks

### RISK-TEST-002 — Flaky tests eroding gate credibility

- Category: Process / technical
- Probability: High
- Impact: High
- Severity: Critical
- Mitigation: FR-TEST-007 determinism controls; ADR-177 quarantine with the hard 14-day check; NFR-TEST-005 budgets; no retry-on-failure in gating CI
- Detection: Nightly determinism lane; trunk failure triage labels; NFR-TEST-005 weekly report
- Owner: Testing and quality (Volume 13)
- Status: Open

### RISK-TEST-003 — Coverage gaming

- Category: Process
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Mutation testing on the packages that matter (ADR-175, NFR-TEST-004); review guidance directing assertions to behavior; strict-scope coverage held above overall (NFR-TEST-003)
- Detection: Mutation-score trend diverging from coverage trend; surviving-mutant reviews
- Owner: Testing and quality (Volume 13)
- Status: Open

### RISK-TEST-004 — Test-double divergence from real systems

- Category: Technical
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Fakes pass the same contract kits as real implementations (FR-TEST-004, FR-TEST-006); cassette regression plus scheduled live lanes cross-check wire fakes (ADR-176)
- Detection: Live-lane drift reports; E-TEST-002 failures; kit failures on fakes
- Owner: Testing and quality (Volume 13)
- Status: Open

### RISK-TEST-005 — Secret or sensitive-data leakage through test assets

- Category: Security
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: FR-TEST-008 rules (synthetic-only data, reserved fake-credential prefix, deny-by-default sanitization, fork-isolated live credentials); secret scanning over all test trees as a required check
- Detection: Secret scanning; sanitizer refusals; scanner self-checks with planted secrets
- Owner: Testing and quality (Volume 13)
- Status: Open
