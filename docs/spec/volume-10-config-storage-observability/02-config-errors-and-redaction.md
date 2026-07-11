# 02 — Configuration Errors and Secret Redaction

This chapter completes the configuration contract of chapter
[01](01-configuration-model.md): the exact error-reporting contract for configuration
failures, worked examples (a minimal valid document, invalid documents with their exact
messages), the complete E-CFG error catalog under the ADR-016 envelope, and the secret
hygiene and redaction rules that govern every configuration surface. Storage-integrity
errors minted here carry exit code 9 per ADR-016 and ADR-029; all other E-CFG errors carry
exit code 3.

## Error reporting contract

### FR-CFG-010 — Configuration error reporting

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Configuration Manager (Volume 10)
- Affected components: Configuration Manager, CLI, TUI, Logging
- Dependencies: ADR-008, ADR-016; FR-CFG-007; FR-CLI-001 (output discipline)
- Related risks: RISK-CFG-002

#### Description

Every configuration failure MUST be reported with: the stable E-CFG code; a user message
in TOML vocabulary (table, key, value class — never internal type names); the position
(file path, line, column) and layer for file-borne findings, or the environment variable
name, flag name, or profile name otherwise; a concrete remedy; and the exit code. The
human rendering follows this template exactly (one block per finding, all findings in one
run per FR-CFG-007):

```text
andromeda: <error|warning> E-CFG-NNN: <user message>
  --> <source>:<line>:<column> (layer: <layer>)
  <zero or more detail lines>
  remedy: <recommended action>
```

The final line of a failing invocation is `exit status <code>`. Machine consumption uses
the structured output rules of Volume 8 (`--json` envelope, FR-CLI-001); the JSON payload
carries the same fields (`code`, `severity`, `message`, `source`, `line`, `column`,
`layer`, `remedy`) plus `technical` (the technical message). All renderings pass FR-CFG-011
redaction before emission. Messages MUST be stable enough to test: the catalog below fixes
each error's user-message template, and golden fixtures pin the rendered form.

#### Motivation

Exit code 3 failures are a primary first-run experience; a fixed shape with positions and
remedies converts them from support tickets into self-service fixes, and stable codes let
scripts and dashboards branch without parsing prose (ADR-016).

#### Actors

Users reading stderr; scripts consuming `--json`; the TUI error surfaces; support.

#### Preconditions

A validation report or loading failure exists (FR-CFG-002, FR-CFG-007).

#### Main flow

1. Findings are collected (all of them — never first-only).
2. Each renders per the template with redaction applied.
3. The process exits with the highest-severity finding's exit code (9 beats 3).

#### Alternative flows

- Warning-only report (deprecations, heuristic detections): rendered the same way, no
  termination, exit unaffected.
- IPC surfaces (ADR-032): the envelope travels as structured data; no template rendering.

#### Edge cases

- A finding without a file position (environment variable, profile row): the `-->` line
  carries the variable or profile name and omits line/column.
- Terminal narrower than the template: lines wrap; the code and position stay on their
  own lines (Volume 8 rendering rules).
- Hundreds of findings: all render; the summary line states the total (no truncation of
  the machine report; human rendering MAY page per Volume 8).

#### Inputs

Validation reports; loading failures; storage procedure failures.

#### Outputs

Rendered findings on stderr; structured findings under `--json`; exit status.

#### States

None.

#### Errors

This requirement is the rendering contract for every error in the catalog below.

#### Constraints

The template is fixed; positions are 1-based; paths render as written by the user with
`~` abbreviation for the home directory.

#### Security

Rendering happens strictly after FR-CFG-011 redaction; technical messages never contain
more value content than user messages — both are redacted.

#### Observability

`config.validation.failed` carries finding counts by code; every rendered failure also
lands in the structured log (chapter 03) with the same code and redacted fields.

#### Performance

Rendering is linear in findings; no measurable budget beyond NFR-CFG-001.

#### Compatibility

The template and JSON field set are public contract under SM-20; additions are new fields,
never renames.

#### Acceptance criteria

- Given a file with two findings, when validation fails, then both render per the
  template, each with code, position, layer, and remedy, and exit status is 3.
- Given `--json`, when the same failure occurs, then the structured payload contains both
  findings with identical codes and positions.
- Given a storage integrity failure and a configuration error in one invocation, when it
  terminates, then exit status is 9.
- Negative case: given a finding whose value fired a secret detector, when rendered, then
  the value appears as a redaction token in both human and JSON forms.
- Permission case: rendering requires no permission and causes no side effects beyond
  output and logs.
- Observability case: finding counts in `config.validation.failed` match the rendered
  report.

#### Verification method

Golden-fixture tests for every catalog entry's rendered form (human and JSON); property
test that machine and human reports always carry identical finding sets; redaction
assertions over seeded-secret fixtures (NFR-CFG-004 method).

#### Traceability

ADR-008, ADR-016; FR-CFG-007, FR-CFG-011; FR-CLI-001; RISK-CFG-002.

## Minimal example

The smallest useful `andromeda.toml` — one provider, nothing else; every other key
resolves from built-in defaults:

```toml
config_version = 1

[providers.local]
adapter  = "ollama"
enabled  = true
base_url = "http://127.0.0.1:11434"
```

## Invalid examples with exact messages

Each example below is intentionally invalid and MUST be rejected with the message shown
(paths and line numbers illustrate the template; the code, message text, and remedy are
normative per the catalog).

**Syntax error** — unclosed string (E-CFG-002):

```toml invalid
[cli]
color = "auto
pager = "auto"
```

```text
andromeda: error E-CFG-002: configuration file is not valid TOML
  --> ~/.config/andromeda/andromeda.toml:2:9 (layer: global file)
  unterminated string starting at line 2, column 9
  remedy: fix the TOML syntax at the reported position
exit status 3
```

**Unknown key** — typo with suggestion (E-CFG-003):

```toml invalid
[cli]
collor = "never"
```

```text
andromeda: error E-CFG-003: unknown configuration key "cli.collor"
  --> ~/.config/andromeda/andromeda.toml:2:1 (layer: global file)
  did you mean "cli.color"?
  remedy: correct or remove the key; the configuration reference lists every valid key
exit status 3
```

**Type mismatch** (E-CFG-004):

```toml invalid
[agent.loop]
max_iterations = "fifty"
```

```text
andromeda: error E-CFG-004: invalid type for "agent.loop.max_iterations"
  --> .andromeda/andromeda.toml:2:18 (layer: workspace file)
  expected integer, found string "fifty"
  remedy: set an integer value, for example: max_iterations = 50
exit status 3
```

**Invalid enum value** (E-CFG-005):

```toml invalid
[cli]
color = "sometimes"
```

```text
andromeda: error E-CFG-005: invalid value for "cli.color"
  --> ~/.config/andromeda/andromeda.toml:2:9 (layer: global file)
  "sometimes" is not one of: "auto", "always", "never"
  remedy: choose one of the listed values
exit status 3
```

**Cross-key violation** — fallback chain naming an undeclared provider (E-CFG-006):

```toml invalid
config_version = 1

[[providers.fallback.chains]]
name    = "escape"
from    = "anthropic"
targets = ["acme"]
```

```text
andromeda: error E-CFG-006: configuration keys conflict
  --> ~/.config/andromeda/andromeda.toml:6:1 (layer: global file)
  fallback chain "escape" targets provider "acme", but no [providers.acme] is declared
  remedy: declare [providers.acme] or remove it from the chain's targets
exit status 3
```

**Include escape** (E-CFG-007):

```toml invalid
include = ["../../home/user/private.toml"]
```

```text
andromeda: error E-CFG-007: include cannot be loaded
  --> .andromeda/andromeda.toml:1:12 (layer: workspace file)
  "../../home/user/private.toml" resolves outside the workspace root
  remedy: move the file inside the workspace or remove the include
exit status 3
```

**Inline secret material** (E-CFG-012) — the value is never echoed:

```toml invalid
[mcp.servers.tracker]
transport = "streamable_http"
url       = "https://tracker.example.com/mcp"
headers   = { Authorization = "Bearer sk-live-abcdef1234567890" }
```

```text
andromeda: error E-CFG-012: secret material detected in configuration
  --> .andromeda/andromeda.toml:4:13 (layer: workspace file)
  value of "mcp.servers.tracker.headers.Authorization" matched detector "bearer_header";
  value redacted: [redacted:bearer_header]
  remedy: store the secret with `andromeda auth` and reference it by credential label;
  configuration files never carry secrets
exit status 3
```

**Future schema version** (E-CFG-010):

```toml invalid
config_version = 9
```

```text
andromeda: error E-CFG-010: configuration schema version 9 is newer than this build
  --> ~/.config/andromeda/andromeda.toml:1:18 (layer: global file)
  this Andromeda build supports configuration schema versions up to 1
  remedy: update Andromeda, or restore the backup written when the file was migrated
exit status 3
```

## Secret hygiene and redaction

Configuration is structurally secret-free (INV-CFGP-02; ADR-014: `andromeda.toml` is never
a credential carrier; keystone FR-SEC-102 owns storage). This section defines the two
enforcement mechanisms: **detection** (secrets must not enter configuration) and
**redaction** (whatever passes through configuration surfaces cannot leak).

```mermaid
flowchart LR
    V[Configuration value] --> C{Sensitivity class}
    C -->|sensitive key| R[Redaction token]
    C -->|public key| D{Detectors}
    D -->|definite match| E[E-CFG-012 error<br/>+ redacted rendering]
    D -->|heuristic match| W[Warning finding<br/>+ redacted rendering]
    D -->|no match| P[Rendered as written]
    R --> S[Sinks: errors, logs, events,<br/>exports, listings, snapshots]
    E --> S
    W --> S
    P --> S
```

The diagram shows the per-value decision: values of keys classed `sensitive` in the schema
always render as redaction tokens in every sink; values of `public` keys pass the detector
registry, where a definite detector match is a fatal E-CFG-012 finding (and the value is
redacted in the finding itself), a heuristic match produces a warning finding with
redacted rendering, and a clean value renders as written. All five sink families — error
renderings, structured logs, events, exports, and configuration listings/snapshots — sit
behind the same redaction step; no sink bypasses it.

### FR-CFG-011 — Secret detection and redaction in configuration surfaces

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Configuration Manager (Volume 10)
- Affected components: Configuration Manager, Logging, Event Bus, CLI, TUI, Persistence Layer
- Dependencies: ADR-014, ADR-134; FR-CFG-007, FR-CFG-010; FR-SEC-102 (storage model)
- Related risks: RISK-CFG-003

#### Description

**Sensitivity classes.** Every schema key carries a sensitivity class: `public` (default)
or `sensitive`. `sensitive` keys are those whose values plausibly embed secrets even when
used correctly: `mcp.servers.<name>.headers`, `auth.proxy.url` (may embed URL userinfo),
and any key an owning volume classes sensitive in its catalog. Values of `sensitive` keys
are NEVER rendered in any sink; they appear as `[redacted:<key-class>]`, or as
`[redacted:sha256:<first 8 hex digits>]` in run configuration snapshots, where the hash of
the value enables equality comparison for reproducibility without disclosure.

**Detector registry.** A versioned registry of named detectors runs over every string
value of every layer document, profile, environment mapping, and override:

| Detector | Class | Rule |
|---|---|---|
| `pem_block` | definite | Value contains a PEM private-key header (`-----BEGIN` … `PRIVATE KEY-----`) |
| `bearer_header` | definite | Header-typed value carrying an authorization scheme followed by token material |
| `url_userinfo` | definite | URL value with a non-empty userinfo password component |
| `known_prefix` | definite | Value matches a registered third-party token prefix pattern; the concrete pattern set is PENDING VALIDATION against each vendor's published token-format documentation (open question in `99-volume-register.md`) and ships as versioned registry data |
| `high_entropy` | heuristic | String of length ≥ 20 with Shannon entropy ≥ 3.5 bits/character, not a path, URL without userinfo, or ULID |

**Behavior.** A `definite` match in any configuration document is the fatal finding
E-CFG-012 (the configuration MUST NOT be used), with the matched value redacted in the
finding itself and `config.secret.detected` emitted (key path and detector name — never
the value, never a value fragment). A `heuristic` match is a warning finding: resolution
proceeds, but the value renders redacted in every sink. Detector evaluation MUST also run
on `andromeda config set` inputs (write-path refusal, Volume 8 surface).

**Sinks.** Redaction applies before every sink: error renderings (user and technical
fields), validation reports, structured logs (chapter 03 handler), events (chapter 04
envelope payloads), entity exports (Volume 2 chapter 10 export rules), configuration
listings (`config get`/`list` renderings), and run configuration snapshots. Environment
variable values render redacted in any diagnostic that echoes the environment.

#### Motivation

The most common credential leak in developer tools is a secret pasted into a config file
that is then committed, logged, or exported. Fail-closed detection stops the entry;
universal redaction bounds the damage of anything that slips through as a heuristic-only
match.

#### Actors

Users editing configuration; the Configuration Manager; every sink component.

#### Preconditions

Schema with sensitivity classes; detector registry loaded (compiled in, versioned).

#### Main flow

1. Every loaded value passes classification, then detectors.
2. Findings join the validation report (FR-CFG-007 completeness).
3. Sinks receive only redacted renderings.

#### Alternative flows

- Value is both `sensitive`-classed and detector-matched: the class rule applies (always
  redacted) and the definite finding still fails the configuration.

#### Edge cases

- False-positive heuristic match on a legitimate opaque identifier: resolution is not
  blocked (heuristic never blocks); the redacted rendering is cosmetic; no suppression
  mechanism exists for definite detectors (fail closed is the contract).
- Secret split across an include and the including file: detectors run per-document on
  final values after layer assembly as well as per-document, so concatenation tricks
  within one layer are still scanned at merge time.
- A `sensitive`-classed key set via environment variable: permitted (environment is not a
  persisted file); the value is still redacted in every sink.

#### Inputs

Layer documents, profiles, environment mappings, overrides, `config set` inputs.

#### Outputs

Findings; redacted renderings; `config.secret.detected` events.

#### States

None.

#### Errors

E-CFG-012 (definite detection, exit 3); heuristic findings are warning-severity entries
in the validation report.

#### Constraints

Detectors are pure functions over values; registry updates ship with releases (no runtime
download); redaction tokens are ASCII and grep-stable.

#### Security

This requirement is itself a security control; its failure mode is asymmetric by design —
false positives block a configuration (recoverable), false negatives leak secrets
(unrecoverable), so definite detectors bias toward blocking. Detection events carry no
value fragments; hashes in snapshots use SHA-256 truncated to 8 hex digits, which
identifies equality without enabling practical value recovery for high-entropy secrets.

#### Observability

`config.secret.detected` (key path, detector, document); redaction counters in the
validation report; NFR-CFG-004 measures end-to-end effectiveness.

#### Performance

Detector cost is linear in total value bytes (bounded by the 1 MiB per-file cap,
FR-CFG-002); within NFR-CFG-001.

#### Compatibility

New detectors and sensitivity-class additions are additive; reclassifying a key to
`public` or removing a definite detector requires an ADR (weakening a security control).

#### Acceptance criteria

- Given a PEM private key as any string value, when validated, then E-CFG-012 names the
  key path and detector, the rendered finding contains no key material, and exit is 3.
- Given `auth.proxy.url = "https://user:hunter2@proxy.example.com"`, when validated, then
  E-CFG-012 (detector `url_userinfo`) and the password never appears in any output.
- Given a heuristic-matched opaque string, when resolved, then resolution succeeds and
  `config list` renders the value as a redaction token.
- Given a `sensitive`-classed key, when a run snapshot persists, then the snapshot stores
  the hashed token form and replay equality-checks against it.
- Negative case: given a clean document, when validated, then zero detector findings and
  zero redaction tokens appear.
- Permission case: `config set` with a definite-matched value is refused before any file
  write occurs.
- Observability case: every definite detection produces exactly one
  `config.secret.detected` with no value fragment (asserted by fixture scan).

#### Verification method

Seeded-secret corpus per NFR-CFG-004 (planted secrets across files, profiles, environment,
overrides; assertion that zero fragments appear in any sink output); detector unit tests
with true/false-positive suites; snapshot hash round-trip tests; write-path refusal tests.

#### Traceability

PRD-005, PRD-006; ADR-014, ADR-134; INV-CFGP-02; FR-SEC-102; SM-16; RISK-CFG-003.

## E-CFG error catalog

Every error declares the full ADR-016 envelope. Common values, stated once: **HTTP
mapping**: none — Andromeda has no HTTP server surface; IPC surfaces (ADR-032) carry the
envelope verbatim over JSON-RPC. **Producer**: Configuration Manager for E-CFG-001..014,
Persistence Layer for E-CFG-015..019. User-message templates are fixed; `<>` marks
substituted fields.

### E-CFG-001 — Configuration file unreadable

- Code: E-CFG-001
- Category: configuration
- Severity: error
- User message: `configuration file cannot be read: <path> (<reason: not found at explicit path | permission denied | I/O error | exceeds 1 MiB size limit>)`
- Technical message: OS error string, file size where relevant, resolution source (explicit `ANDROMEDA_CONFIG` or ADR-022 location).
- Cause: Explicit path missing; permission or I/O failure; oversize file.
- Safe context data: path, layer, reason class, size limit.
- Recoverability: user-recoverable (fix path, permissions, or size).
- Retry policy: none (deterministic until the environment changes).
- Recommended action: verify the path and its permissions; for oversize files, split via includes (FR-CFG-006).
- Exit code: 3.
- HTTP mapping: none (see catalog header).
- Telemetry event: `config.validation.failed`.
- Security implications: none beyond path disclosure in local output.

### E-CFG-002 — Configuration file is not valid TOML

- Code: E-CFG-002
- Category: configuration
- Severity: error
- User message: `configuration file is not valid TOML` with position detail from the strict parser.
- Technical message: parser diagnostic verbatim (ADR-008 position-aware).
- Cause: Syntax error in a layer or included file.
- Safe context data: path, line, column, layer; never the offending raw line (it may contain a partially typed secret).
- Recoverability: user-recoverable.
- Retry policy: none.
- Recommended action: fix the syntax at the reported position.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.validation.failed`.
- Security implications: raw-line suppression prevents accidental secret echo.

### E-CFG-003 — Unknown configuration key

- Code: E-CFG-003
- Category: configuration
- Severity: error
- User message: `unknown configuration key "<dotted path>"`, plus `did you mean "<suggestion>"?` when edit distance ≤ 2, plus the introducing or removing schema version when the key belongs to another version.
- Technical message: schema version consulted, candidate suggestions with distances.
- Cause: Typo; key from a newer schema; key removed after deprecation; foreign tool's key.
- Safe context data: key path, position, layer, suggestion, schema versions.
- Recoverability: user-recoverable.
- Retry policy: none.
- Recommended action: correct or remove the key; consult the generated configuration reference.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.validation.failed`.
- Security implications: strict rejection is a trust-boundary control (ADR-008).

### E-CFG-004 — Invalid type for configuration key

- Code: E-CFG-004
- Category: configuration
- Severity: error
- User message: `invalid type for "<key>"` with `expected <schema type>, found <TOML type> <redacted-or-shown value>`.
- Technical message: declared type, parse failure detail; for environment sources, the variable name.
- Cause: Wrong TOML type; unparseable environment value; flag value of the wrong shape.
- Safe context data: key, expected type, found type, position or variable name; the found value only when the key is `public` and no detector matched.
- Recoverability: user-recoverable.
- Retry policy: none.
- Recommended action: set a value of the documented type; the message shows a valid example.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.validation.failed`.
- Security implications: value echo is gated by FR-CFG-011.

### E-CFG-005 — Invalid value for configuration key

- Code: E-CFG-005
- Category: configuration
- Severity: error
- User message: `invalid value for "<key>"` with the violated rule rendered (`"<value>" is not one of: <enum list>` / `must be between <min> and <max>` / `must match <pattern description>` / duration grammar violations).
- Technical message: rule identifier and parameters.
- Cause: Enum, range, pattern, or duration-grammar violation on a correctly-typed value.
- Safe context data: key, rule, allowed set/range, position; value under FR-CFG-011 gating.
- Recoverability: user-recoverable.
- Retry policy: none.
- Recommended action: choose a value satisfying the rendered rule.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.validation.failed`.
- Security implications: none beyond value-echo gating.

### E-CFG-006 — Configuration keys conflict

- Code: E-CFG-006
- Category: configuration
- Severity: error
- User message: `configuration keys conflict` with the specific cross-key rule rendered (undeclared reference target, overlapping project roots, deprecated key and replacement set to different values, rename conflict during migration).
- Technical message: rule identifier, both key paths and positions.
- Cause: Cross-key constraint violation on the merged document (FR-CFG-007 rule 4).
- Safe context data: rule, key paths, positions, layers.
- Recoverability: user-recoverable.
- Retry policy: none.
- Recommended action: rule-specific remedy rendered per FR-CFG-010.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.validation.failed`.
- Security implications: reference-target checks prevent typo-routing to unintended providers.

### E-CFG-007 — Include cannot be loaded

- Code: E-CFG-007
- Category: configuration
- Severity: error
- User message: `include cannot be loaded` with sub-class detail: `"<path>" does not exist` / `include cycle: <chain>` / `include depth exceeds 8` / `more than 64 files in one layer` / `"<path>" resolves outside <containment root>` / `reserved key "<key>" not allowed in an included file`.
- Technical message: resolved absolute paths, symlink resolution trail for containment findings.
- Cause: FR-CFG-006 bound or containment violation.
- Safe context data: include chain, bounds, containment root.
- Recoverability: user-recoverable.
- Retry policy: none.
- Recommended action: sub-class-specific remedy (move file inside root, break the cycle, flatten nesting).
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.validation.failed`.
- Security implications: containment refusal is the control against include-based exfiltration of files outside the workspace.

### E-CFG-008 — Configuration profile cannot be resolved

- Code: E-CFG-008
- Category: configuration
- Severity: error
- User message: `configuration profile "<name>" not found` with the scopes searched, or `profile "<name>" was deleted` for a selection that no longer resolves.
- Technical message: selector source (flag, `ANDROMEDA_PROFILE`, file key, entity default), scopes searched, nearest-name suggestion.
- Cause: Unknown, deleted, or out-of-scope profile name (FR-CFG-003 fail-closed rule).
- Safe context data: profile name, selector source, scopes searched.
- Recoverability: user-recoverable.
- Retry policy: none.
- Recommended action: list available profiles (Volume 8 surface) and select an existing one, or clear the selector.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.validation.failed`.
- Security implications: fail-closed selection prevents silently running without a deliberately chosen (possibly restrictive) profile.

### E-CFG-009 — Environment variable mapping failure

- Code: E-CFG-009
- Category: configuration
- Severity: error
- User message: `environment variable <NAME> <does not map to any configuration key | is ambiguous | duplicates <OTHER NAME> | addresses a table, which cannot be set from the environment>`; the ambiguous form lists every candidate with its explicit `__` form; the unknown form carries a nearest-key suggestion.
- Technical message: candidate set, declared dynamic-table entries consulted.
- Cause: FR-CFG-004 rules 3–5 violations.
- Safe context data: variable name, candidates, suggestion; never the variable's value.
- Recoverability: user-recoverable.
- Retry policy: none.
- Recommended action: use the listed explicit `__` form, correct the name, or unset the variable.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.validation.failed`.
- Security implications: rejecting unknown `ANDROMEDA_*` variables surfaces stale or attacker-planted environment rather than ignoring it.

### E-CFG-010 — Configuration schema version unsupported

- Code: E-CFG-010
- Category: configuration
- Severity: error
- User message: `configuration schema version <declared> is newer than this build` plus `this Andromeda build supports configuration schema versions up to <current>`.
- Technical message: document path, declared and supported versions, backup file path when one exists beside the document.
- Cause: Document written by a newer Andromeda (sync or downgrade scenario).
- Safe context data: versions, path, backup path.
- Recoverability: user-recoverable (upgrade Andromeda or restore the backup).
- Retry policy: none.
- Recommended action: update Andromeda, or restore the `.bak.v<n>` backup written at migration time.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.validation.failed`.
- Security implications: refusing partial reads of future schemas prevents silently dropping future security-relevant keys.

### E-CFG-011 — Configuration migration failed

- Code: E-CFG-011
- Category: configuration
- Severity: error
- User message: `configuration file migration failed for <path>: <step> (original file unchanged)`.
- Technical message: failing transform identifier, source and target versions, I/O detail.
- Cause: I/O failure or transform conflict during an explicit file rewrite (FR-CFG-008 rule 3).
- Safe context data: path, versions, transform identifier, backup path.
- Recoverability: recoverable — the original file and backup are intact.
- Retry policy: manual retry after addressing the cause.
- Recommended action: resolve the reported cause (disk space, permissions, key conflict) and re-run the migration.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.migration.failed`.
- Security implications: partial outputs are removed, so no half-migrated file can be loaded later.

### E-CFG-012 — Secret material detected in configuration

- Code: E-CFG-012
- Category: configuration
- Severity: error
- User message: `secret material detected in configuration` naming the key path and detector, with the value rendered only as `[redacted:<detector>]`.
- Technical message: detector name and registry version; match length class; never content.
- Cause: A definite detector matched a configuration value (FR-CFG-011).
- Safe context data: key path, detector name, document, position.
- Recoverability: user-recoverable.
- Retry policy: none.
- Recommended action: remove the secret from the file, store it via the authentication surfaces (Secret Store, ADR-014), reference it by credential label, and rotate the exposed secret.
- Exit code: 3.
- HTTP mapping: none.
- Telemetry event: `config.secret.detected`.
- Security implications: the exposed secret MUST be treated as compromised (it exists in a plaintext file); the remedy says to rotate it. The error output itself is leak-free by construction.

### E-CFG-013 — Deprecated configuration key in use

- Code: E-CFG-013
- Category: configuration
- Severity: warning
- User message: `configuration key "<old>" is deprecated since schema version <n>` plus `use "<replacement>"` when one exists and the removal release.
- Technical message: deprecation metadata (since, replacement, removal release).
- Cause: A document sets a key marked deprecated (FR-CFG-008 rule 5).
- Safe context data: old key, replacement, versions, position.
- Recoverability: user-recoverable; resolution proceeds meanwhile.
- Retry policy: none.
- Recommended action: rename the key (or run the migration surface, which rewrites it).
- Exit code: 3 — applied only when a validation surface is invoked with warnings promoted to errors (Volume 8 surface semantics); during normal startup this finding never terminates.
- HTTP mapping: none.
- Telemetry event: `config.deprecation.detected`.
- Security implications: none.

### E-CFG-014 — Runtime override refused

- Code: E-CFG-014
- Category: configuration
- Severity: error
- User message: `runtime override of "<key>" is not permitted: <table> settings can only be changed in configuration files or profiles`.
- Technical message: protected-table rule reference, override origin (TUI or IPC peer).
- Cause: Override targeting `[permissions]`, `[sandbox]`, `[security]`, `[auth]`, or `[telemetry]` (FR-CFG-005).
- Safe context data: key, table, origin class.
- Recoverability: not applicable — the refusal is the correct outcome; the persistent path is the alternative.
- Retry policy: none.
- Recommended action: change the value in the appropriate file or profile so the change is attributable and durable.
- Exit code: 3 (surfaces only on explicit override requests; never at startup).
- HTTP mapping: none.
- Telemetry event: `config.override.rejected`.
- Security implications: this refusal is the control preventing ephemeral, low-visibility weakening of security and consent posture.

### E-CFG-015 — Database schema is newer than this build

- Code: E-CFG-015
- Category: storage integrity
- Severity: critical
- User message: `<database> was written by a newer Andromeda (schema <found>, this build supports <supported>)` with the remedy naming the paired-rollback rule.
- Technical message: database path, `user_version`, `schema_migrations` head, binary target version.
- Cause: Older binary opening a newer database (ADR-029 rule 5).
- Safe context data: paths, versions.
- Recoverability: recoverable via upgrade, or via restoring the paired binary + database backup (ADR-029 rule 6).
- Retry policy: none.
- Recommended action: update Andromeda, or restore the database backup taken when it was migrated.
- Exit code: 9.
- HTTP mapping: none.
- Telemetry event: `storage.integrity.failed`.
- Security implications: refusing partial reads prevents misinterpreting future security-relevant rows (permissions, audit).

### E-CFG-016 — Database migration failed

- Code: E-CFG-016
- Category: storage integrity
- Severity: critical
- User message: `database migration <n> failed on <database>; the database and its backup are preserved` naming the newest verified backup.
- Technical message: failing migration number and checksum, SQLite error, paths of database and backup.
- Cause: Migration script failure, checksum mismatch, or post-check failure (ADR-029 rule 4).
- Safe context data: migration number, versions, paths.
- Recoverability: recoverable — restore-from-backup, then retry after upgrading or reporting the defect.
- Retry policy: manual after cause resolution; the chain re-runs idempotently from the recorded version.
- Recommended action: restore the named backup (command surface Volume 8) or retry after freeing resources; report checksum mismatches as integrity defects.
- Exit code: 9.
- HTTP mapping: none.
- Telemetry event: `storage.migration.failed`.
- Security implications: checksum verification detects tampered migration scripts.

### E-CFG-017 — Database integrity check failed

- Code: E-CFG-017
- Category: storage integrity
- Severity: critical
- User message: `<database> failed its integrity check; no changes were made` naming the newest verified backup when one exists.
- Technical message: `PRAGMA integrity_check` / `foreign_key_check` output summary; `user_version` vs `schema_migrations` disagreement where that is the trigger.
- Cause: File corruption; version-ledger disagreement (ADR-029 rule 1).
- Safe context data: database path, check name, backup path.
- Recoverability: recoverable via backup restore; cache databases are exempt (rebuilt, never this error).
- Retry policy: none (deterministic until repaired).
- Recommended action: restore the newest backup; if none exists, the message directs to the recovery documentation.
- Exit code: 9.
- HTTP mapping: none.
- Telemetry event: `storage.integrity.failed`.
- Security implications: audit-chain verification (Volume 9) independently detects tampering; this error covers structural corruption.

### E-CFG-018 — Database backup could not be created

- Code: E-CFG-018
- Category: storage integrity
- Severity: critical
- User message: `cannot create the pre-migration backup for <database> (<reason: insufficient disk space | write failure | verification failure>); migration was not started`.
- Technical message: space required vs available, backup path, verification detail.
- Cause: FR-CFG-009 pre-flight or backup verification failure.
- Safe context data: paths, sizes, reason class.
- Recoverability: recoverable — the database is untouched.
- Retry policy: manual after freeing space or fixing the backup location.
- Recommended action: free disk space or point `storage.backups.dir` at a writable location, then retry.
- Exit code: 9.
- HTTP mapping: none.
- Telemetry event: `storage.migration.failed`.
- Security implications: refusing to migrate without a verified backup preserves the ADR-029 recovery guarantee.

### E-CFG-019 — Workspace database is locked

- Code: E-CFG-019
- Category: concurrency
- Severity: error
- User message: `workspace is in use by another Andromeda process (pid <pid>, started <time>); waited <ms> ms`.
- Technical message: lock file path, holder metadata, wait budget.
- Cause: Write-lock contention on `.andromeda/state.lock` beyond `storage.lock_wait_ms` (FR-CFG-009).
- Safe context data: pid, start time, lock path, wait duration.
- Recoverability: recoverable — close the other process or wait; read-only surfaces remain available.
- Retry policy: automatic bounded wait already applied; further retries are manual.
- Recommended action: finish or close the other session; stale locks (dead holder) are broken automatically with a diagnostic.
- Exit code: 1.
- HTTP mapping: none.
- Telemetry event: `storage.lock.denied`.
- Security implications: lock metadata discloses only pid and start time.

## Non-functional requirements

### NFR-CFG-003 — Validation finding completeness

- Category: Usability
- Priority: P1
- Phase: MVP
- Metric: Finding recall of a single validation pass over the seeded-defect corpus — configurations planted with N independent findings (unknown keys, type errors, value rule violations, cross-key conflicts, include violations, deprecations) across multiple files and layers
- Target: 100% of seeded findings reported in one pass, each with correct code, file, line, and layer
- Minimum threshold: 100% (fail-fast validation is a defect, not a degradation)
- Measurement method: Seeded-defect corpus in the test suite (≥ 50 configurations, 1–20 seeded findings each); assertion that the report equals the seed manifest exactly (no misses, no duplicates, correct positions)
- Test environment: CI on Tier 1 platforms
- Measurement frequency: Every merge
- Owner: Configuration Manager (Volume 10)
- Dependencies: FR-CFG-007, FR-CFG-010
- Risks: RISK-CFG-002
- Acceptance criteria: The corpus suite passes with exact-match reports on every entry; adding a new finding kind to the schema adds corpus entries in the same change (checked by review checklist).

### NFR-CFG-004 — Redaction effectiveness

- Category: Security
- Priority: P0
- Phase: MVP
- Metric: Count of planted secret fragments (≥ 8 contiguous characters of any planted secret) appearing in any configuration-surface output — error renderings (human and JSON), validation reports, structured logs, events, exports, configuration listings, and run snapshots — across the seeded-secret corpus
- Target: 0 fragments across all sinks
- Minimum threshold: 0 (identity property; a single leak is a release-blocking defect)
- Measurement method: Seeded-secret corpus (secrets planted in files, includes, profiles, environment variables, and override requests; every detector class and sensitivity class covered); automated scan of every sink output for planted fragments; run per release and on every change to detector or redaction code
- Test environment: CI on Tier 1 platforms; sink capture harness
- Measurement frequency: Every merge touching configuration/redaction code; every release
- Owner: Configuration Manager (Volume 10)
- Dependencies: FR-CFG-011; chapter 03 logging redaction (shared handler)
- Risks: RISK-CFG-003
- Acceptance criteria: Zero fragments detected in any sink for every corpus entry on both Tier 1 platforms; the corpus includes at least one secret per detector class and per sink family.

## Risks

### RISK-CFG-003 — Secrets committed to version control through workspace configuration

- Category: Security
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Structural: configuration is never a credential carrier (INV-CFGP-02, ADR-014) and the schema provides reference indirection (credential labels, `api_key_env`); preventive: fail-closed definite detectors block loading and `config set` writes (FR-CFG-011, E-CFG-012) and the remedy instructs rotation; damage-bounding: universal sink redaction (NFR-CFG-004 target 0 leaks)
- Detection: E-CFG-012 frequency in local diagnostics; `config.secret.detected` events; repository secret scanning in the project's own CI (Volume 11)
- Owner: Configuration Manager (Volume 10)
- Status: Open

Workspace files (`.andromeda/andromeda.toml`) are designed to be shareable and therefore
committable; a pasted key in a committed file is the highest-likelihood credential leak
this product can cause. Detection cannot be perfect — the residual risk is a secret shape
no detector recognizes — so the structural rule (references only) and the rotation
instruction in E-CFG-012 carry the real weight.

## Events minted by this chapter

Envelope per chapter 04 of this volume (keystone FR-OBS-001).

| Event | Producer | Emitted when |
|---|---|---|
| `config.validation.completed` | Configuration Manager | A validation pass finishes with no error-severity findings |
| `config.validation.failed` | Configuration Manager | A validation pass finishes with at least one error-severity finding |
| `config.secret.detected` | Configuration Manager | A definite detector matches a configuration value |
