# volume-08-cli-and-tui — Volume Register

Merged from per-agent register fragments at the Phase B gate.

## Requirements index

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-CLI-001 | CLI command grammar | Core | Grammar golden tests over the full tree; argv fuzzing; SM-20 contract-diff; consolidation audit of chapters 03–06 |
| FR-CLI-002 | Runtime mediation and driver parity | Core | ADR-033 dependency checks; CLI/IPC parity record comparison; permission mediation tests (SM-16b) |
| FR-CLI-003 | Root command and TUI hand-off | MVP | teatest launch tests; pipe/CI matrix asserting exit 2 and clean stdout; `--version` alias golden |
| FR-CLI-004 | Extension-contributed commands | Beta | SDK extension-command conformance suite; namespace-closure tests; permission mediation over extension origins |
| FR-CLI-005 | Global flags and invocation modes | Core | Global-flag matrix over the tree; cancellation tests; ConfigPort attribution assertions |
| FR-CLI-006 | Structured JSON output for every command | Core | Schema conformance matrix (success + failure per command); NDJSON strict parsing; SM-20 contract-diff; redaction leak tests |
| FR-CLI-007 | Stream discipline: stdout, stderr, exit code | Core | Stream-classification matrix; EPIPE fault injection; redirection goldens |
| FR-CLI-008 | Verbosity modes: quiet, verbose, debug | MVP | Recording-parity tests; redaction leak tests; per-level stderr goldens |
| FR-CLI-009 | Non-interactive and CI modes | MVP | Decision-table unit tests; NFR-CLI-003 prompt-free matrix; policy fixture parity tests |
| FR-CLI-010 | Confirmation behavior | MVP | PTY prompt-driving tests; non-interactive matrix; audit-record assertions; golden prompt texts |
| FR-CLI-011 | Environment variables | MVP | Environment matrix; precedence tests vs. ConfigPort attribution; truthy-parser unit tests |
| FR-CLI-012 | Shell completion | MVP | Per-shell completion harness in CI; dynamic-completion fixtures; silent-empty fault injection |
| FR-CLI-013 | Core command family behavior | MVP | Per-command goldens and schema conformance; PTY prompt tests; non-interactive matrix; UC-01/UC-07/UC-11 E2E |
| FR-CLI-014 | Platform command family behavior | MVP | Grammar goldens; tampered-artifact and stopped-server fixtures; confirmation matrix; UC-12 E2E at Beta |
| FR-CLI-015 | Data command family behavior | MVP | Offline suite (SM-05); export schema conformance; permission-denial fixtures; redaction leak tests |
| FR-CLI-016 | Maintenance command family behavior | MVP | Doctor fixture matrix; update E2E with tampered artifacts and interrupts (SM-18/SM-19); cold-start benchmark (SM-06a); offline suite |
| FR-UX-001 | Error presentation standard | MVP | Golden-format tests per error family; leak tests; JSON/human parity assertions |
| FR-UX-002 | Terminal capability adaptation and paging | MVP | Byte-classification over TTY/pipe matrix; styled/plain parity diff; pager fault injection |
| FR-UX-003 | Progress reporting outside the TUI | MVP | PTY/pipe capture with byte classification; heartbeat cadence with mock clocks; recording-parity assertions |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-CLI-001 | Structured-output schema stability | Beta | Contract-diff of published schemas per release; CI validation of emitted output against schemas |
| NFR-CLI-002 | Help and reference completeness | MVP | Automated help-coverage walk of the tree in CI; docs generation gate |
| NFR-CLI-003 | Prompt-free non-interactive operation | MVP | Piped full-command matrix instrumented for TTY reads, with confirmation/approval fixtures |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-CLI-001 | Grammar and surface sprawl across releases | High | Open |
| RISK-CLI-002 | TUI hand-off misdetection corrupting scripted output | Medium | Open |
| RISK-CLI-003 | Structured-output drift breaking automation | High | Open |
| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-TUI-001 | TUI shell | MVP | teatest/v2 golden frames; PTY integration tests incl. panic/kill restoration; event assertions |
| FR-TUI-002 | Panel system and layout manager | MVP | Golden frames per layout class; geometry property tests |
| FR-TUI-003 | Navigation and focus model | MVP | Scripted interaction suites; focus-trap and typed-ahead-discard tests |
| FR-TUI-004 | Keyboard command map | MVP | Scripted keystroke matrix; random-sequence fuzz for panics/unintended actions |
| FR-TUI-005 | Mouse input | MVP | Synthetic mouse-event tests; PTY capture of reporting sequences |
| FR-TUI-006 | Resize and small-terminal behavior | MVP | Resize scripts with frame capture; SIGWINCH integration; purity property test |
| FR-TUI-007 | Runtime event rendering and streaming pipeline | MVP | Load tests with mock streaming provider; overflow/fault injection; byte-equality vs persisted records |
| FR-TUI-008 | Theme configuration and resolution | MVP | Configuration-matrix PTY tests; watch-repaint test; validation exit-code tests |
| FR-TUI-009 | Core screen inventory and content contract | MVP | Golden-frame enumeration suite; permission-mediation instrumentation; fault-injection screen matrix |
| FR-UX-040 | Closed design-token vocabulary with fixed Danger | MVP | Golden-frame color scanning; color-literal grep gate; contrast recomputation script |
| FR-UX-041 | Token-to-ANSI degradation tiers | MVP | Per-tier golden frames; SGR-legality scanners |
| FR-UX-042 | Light-terminal fallback theme | MVP | Light-mode golden frames; scripted OSC 11 responder; contrast scanner |
| FR-UX-043 | Splash and identity surfaces | MVP | Golden frames for splash variants; policy-matrix start tests; timing assertions |
| NFR-TUI-001 | Small-terminal functional completeness | MVP | Scripted 80×24 traversal of all screens/actions; compact-class smoke traversal |
| NFR-TUI-002 | Rendering determinism for golden-frame testing | MVP | Double-run frame diff across the golden matrix in CI |
| NFR-UX-040 | Contrast and non-color redundancy | MVP | Automated contrast computation over tables and scanned frames; danger-marker frame scan |
| RISK-TUI-001 | Presentation-state monolith erodes maintainability and latency | MVP | Golden-suite runtime and SM-07 trend monitoring; review discipline |
| RISK-TUI-002 | Event flood renders the TUI unresponsive or misleading | MVP | Load tests at 10× delta rates; overflow counters |
| RISK-TUI-003 | Wireframe and implementation drift | MVP | FR-TUI-009 enumeration suite; consolidation/release audits |
| RISK-UX-040 | User-palette variance at the ansi16 tier breaks contrast or brand | MVP | Terminal-matrix testing; doctor tier reporting; user reports |

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-TUI-060 | Platform screen catalog and management frame | MVP | teatest golden frames per screen/tier/glyph set; registry-coverage test; driver-parity audit |
| FR-TUI-061 | Command palette | MVP | teatest interaction scripts; registry coverage; latency measurement per NFR-UX-077 |
| FR-TUI-062 | Quick actions | Beta | teatest ranking/staleness/confirmation scripts |
| FR-TUI-063 | Help overlay and keybinding reference | MVP | Registry-diff coverage test; golden frames; offline assertion; keymap-rebind test |
| FR-TUI-064 | Error center and recovery screens | MVP | Crash-injection suite (SM-11 method) via teatest; envelope golden tests; clipboard audit assertion |
| FR-TUI-065 | Accessible output mode | Beta | Byte-stream sequence denylist; registry-parity traversal; AT validation pass (PENDING VALIDATION) |
| FR-TUI-066 | No-color and monochrome operation | MVP | Byte classification; parity diff across tiers; attribute-free terminfo fixture |
| FR-TUI-067 | Glyph tiers and Unicode fallback | MVP | Byte-inventory scans per set; resolution matrix unit tests; width fixtures |
| FR-TUI-068 | SSH, multiplexer, and non-TTY operation | MVP | Remote-PTY golden suite over SSH/tmux/screen; clipboard fault scripts; CI refusal tests |
| FR-UX-070 | Streaming output rendering | MVP | Mock-stream teatest scripts; SM-08 harness; escape-injection fixture corpus |
| FR-UX-071 | Spinners and progress bars | MVP | Time-controlled teatest scripts; reduced-motion golden frames; stall injection |
| FR-UX-072 | Modal overlays and confirmation tiers | MVP | Per-tier interaction scripts; keystroke-buffer injection; z-order golden frames |
| FR-UX-073 | Toasts | MVP | Timing scripts (dismissal, queue, coalescing, modal pause); storm injection |
| FR-UX-074 | Canonical view states | MVP | Signal-injection fixtures; offline suite assertions; per-panel state scripts |
| FR-UX-075 | Copy and paste | MVP | Per-kind scripts; sanitization corpus; mechanism fakes; event payload assertions |
| FR-UX-076 | Data navigation: search, filtering, pagination, virtualization | MVP | Synthetic 100k-row stores; memory accounting; facet/smart-case unit tests; fetch-fault injection |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-TUI-069 | Keyboard reachability and focus visibility | MVP | Registry traversal in teatest; color-stripped focus-marker analysis on golden corpus |
| NFR-TUI-070 | Terminal compatibility conformance | Beta | Compatibility suite on the Tier A terminal set; probe-versus-matrix comparison per release |
| NFR-UX-077 | Interaction feedback deadline | MVP | SM-07 replay harness extended with feedback classification |
| NFR-UX-078 | Virtualized view memory ceiling | MVP | Benchmark-harness process accounting on 100 vs 100,000-record stores; render-state inspection |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-TUI-071 | Assistive-technology experience remains impractical despite accessible output mode | High | Open |
| RISK-TUI-072 | Character-width divergence corrupts layout in edge terminals | Medium | Open |
| RISK-UX-079 | Clipboard exposure of sensitive content | High | Open |
| RISK-UX-080 | Stale or masked view states misrepresenting reality | High | Open |

## ADRs minted

Block 100–114 belongs to Volume 8; fragment A used 100–104 (fragment B mints 105–109,
fragment C 110–114).

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-100](../annexes/adr/ADR-100.md) | CLI grammar style: noun-group command tree with a closed verb vocabulary | Accepted | Resource-noun groups + closed shared verbs + action leaves; singular nouns; depth ≤ 3; no aliases except governed renames; closed top level |
| [ADR-101](../annexes/adr/ADR-101.md) | CLI structured output: human default, `--json` everywhere, NDJSON streams, versioned schemas | Accepted | Boolean `--json`; fixed result envelope; NDJSON for streams; `andromeda.cli.<path>.v<major>` schemas under SM-20 |
| [ADR-102](../annexes/adr/ADR-102.md) | Interactivity resolution order and the scope of `--yes` | Accepted | Ordered resolution (flag > env > `CI` > TTY probe); `--yes` covers destructive confirmations only, never permission approvals |
| [ADR-103](../annexes/adr/ADR-103.md) | CLI color and terminal-capability policy | Accepted | Six-step per-stream styling resolution with `auto` default; JSON never styled; CLI bound to the 16-color-safe token subset |
| [ADR-104](../annexes/adr/ADR-104.md) | Extension-contributed commands mount under the reserved `x` namespace | Accepted | `andromeda x <extension> <command>`; closed top level preserved; conventions and permission mediation inherited mechanically |
| ADR | Title | One-line decision |
|---|---|---|
| ADR-105 | Danger token fixed at `#FF6B6B` with a contrast-validated light-theme variant | Danger = `#FF6B6B` (6.65:1 on Neutral); light theme renders it as `#B3261E` (5.85:1 on Tertiary); closes the ADR-026 PENDING VALIDATION |
| ADR-106 | Terminal capability detection: conservative signal ladder with configuration override | Config override → ADR-103 off-switches → `COLORTERM` → `TERM` 256color → terminfo → none; mode via bounded OSC 11 query, dark default |
| ADR-107 | TUI shell shape: single-window screen-per-concern model with a fixed panel system | One active screen, 1–4 panels, three layout classes; ring + go-to chords + palette navigation |
| ADR-108 | Keyboard model: non-modal fixed keymap with go-to chords at MVP; user remapping at Beta | Non-modal three-layer resolution; reserved overlay-local decision keys; `[tui.keymap]` remapping deferred to Beta with validation |
| ADR-109 | Mouse support as enhancement-only, with guaranteed native-selection escape hatches | Mouse never required; modifier bypass plus `tui.mouse = false` guarantee native selection |

Block 100–114 belongs to Volume 8; fragment C used 110–114 (100–104 fragment A;
105–109 fragment B).

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-110](../annexes/adr/ADR-110.md) | Single TUI action registry | Accepted | Every TUI operation is a named registered action; palette, quick actions, keymap, and help render from one catalog; extensions mount under `x.` |
| [ADR-111](../annexes/adr/ADR-111.md) | Accessible output mode as a first-class linear rendering profile | Accepted | Linear append-only rendering of the same model, explicitly activated; screen-reader sufficiency PENDING VALIDATION with real AT |
| [ADR-112](../annexes/adr/ADR-112.md) | Glyph capability tiers | Accepted | Two closed glyph sets (`unicode`, `ascii`) with declared parity, locale/TERM auto-selection, single-cell chrome only, no emoji in chrome |
| [ADR-113](../annexes/adr/ADR-113.md) | Clipboard integration policy | Accepted | PAL-native first, OSC 52 fallback, 1 MiB bound, user-initiated writes audited, agent access mediated by the `clipboard` permission |
| [ADR-114](../annexes/adr/ADR-114.md) | Terminal support tiers and capability probing | Accepted | Tier A reference set gates releases; probing order config → conventions → terminfo → conservative baseline; degrade, never crash |

## Error codes minted

| Code | Name | Exit code |
|---|---|---|
| E-CLI-001 | Unknown command or flag | 2 |
| E-CLI-002 | Invalid argument or flag value | 2 |
| E-CLI-003 | Confirmation required but unavailable | 2 |
| E-CLI-004 | Extension command unavailable | 1 |
| E-CLI-005 | Interactive terminal required | 1 |
| E-CLI-006 | Conflicting flags | 2 |
| E-CLI-007 | Input read failure | 1 |
| E-CLI-008 | Output write failure | 1 |
| E-CLI-009 | Invocation deadline exceeded | 8 |
| Code | Name | Exit code | Defined in |
|---|---|---|---|
| E-TUI-001 | TUI initialization failure | 1 | 07-tui-architecture.md |
| E-TUI-002 | Terminal capabilities insufficient | 1 | 07-tui-architecture.md |
| E-TUI-003 | Terminal state restoration failure | 1 (non-overriding) | 07-tui-architecture.md |
| E-TUI-004 | Invalid theme configuration | 3 | 08-theming-and-design-tokens.md |
| E-TUI-005 | Invalid keymap configuration | 3 | 07-tui-architecture.md |
| E-TUI-006 | Render pipeline failure | 1 | 07-tui-architecture.md |
| E-TUI-007 | Runtime unavailable to the interface | 1 | 07-tui-architecture.md |
| E-TUI-008 | Event subscription overflow | none (in-TUI, recoverable) | 07-tui-architecture.md |

None. Fragment C mints no error codes: TUI-internal failures use the E-TUI family minted
by fragment B (chapter 07); CLI-boundary refusals reuse fragment A's E-CLI codes
(E-CLI-005 for the non-TTY TUI request).

## Events minted

Envelope and delivery semantics per Volume 10; grammar per Volume 0 chapter 03. Payload
contracts in chapter 02's events table.

| Event | Emitted by |
|---|---|
| `cli.command.started` | CLI |
| `cli.command.completed` | CLI |
| `cli.command.failed` | CLI |
| `cli.confirmation.resolved` | CLI |
| `cli.tui.launched` | CLI |
| `cli.update.notified` | CLI |
| Event | Version | Defined in |
|---|---|---|
| `tui.shell.started` | 1 | 07-tui-architecture.md |
| `tui.shell.exited` | 1 | 07-tui-architecture.md |
| `tui.screen.changed` | 1 | 07-tui-architecture.md |
| `tui.resize.applied` | 1 | 07-tui-architecture.md |
| `tui.render.failed` | 1 | 07-tui-architecture.md |
| `tui.theme.resolved` | 1 | 08-theming-and-design-tokens.md |

Envelope semantics per Volume 10; grammar per Volume 0 chapter 03.

| Event | Emitted by | Notes |
|---|---|---|
| `tui.palette.opened` | TUI | No payload beyond envelope; recency context is local |
| `tui.action.invoked` | TUI | Action identifier, source (`palette`, `quick_actions`, `keybinding`, `screen`, `help`), disposition (`executed`, `cancelled`, `refused`) |
| `tui.clipboard.copied` | TUI | Content kind and byte count; never content (ADR-113) |
| `tui.render_profile.changed` | TUI | Resolved profile: color tier, glyph set, accessible flag, deciding signals |

## Config keys minted

`[cli]` table content (schema, precedence, and validation are Volume 10's): `cli.color`
(`"auto"`), `cli.pager` (`"auto"`), `cli.pager_command` (`""`), `cli.editor` (`""`),
`cli.update_notice` (`true`), `cli.default_timeout` (`"0s"`). The `[tui]` and `[tui.theme]`
tables are fragment B's.
| Key | Type | Default | Defined in |
|---|---|---|---|
| `tui.mouse` | bool | `true` | 07-tui-architecture.md |
| `tui.splash` | string enum | `"auto"` | 07-tui-architecture.md |
| `tui.default_screen` | string enum | `"session"` | 07-tui-architecture.md |
| `tui.theme.mode` | string enum | `"auto"` | 08-theming-and-design-tokens.md |
| `tui.theme.tier` | string enum | `"auto"` | 08-theming-and-design-tokens.md |

`[tui.keymap]` is reserved for Beta remapping (ADR-108); its keys are minted when that
feature is specified.

`[tui]` table content (schema/precedence Volume 10's; `[tui.theme]` is fragment B's):
`tui.glyphs` ("auto"), `tui.reduce_motion` (false), `tui.accessible_output` (false),
`tui.clipboard` ("auto"), `tui.clipboard_max_bytes` (1048576), `tui.toast_duration_ms`
(4000), `tui.search_debounce_ms` (150), `tui.list_page_size` (100).

## Glossary additions

| Term | One-line meaning |
|---|---|
| Invocation-mode record | The immutable per-invocation resolution of interactivity, CI mode, color, output format, and verbosity that every pipeline stage consults (chapter 01). |
| Result envelope | The fixed eight-field JSON document (`schema`, `command`, `ok`, `exit_code`, `data`, `error`, `warnings`, `meta`) every `--json` command emits (FR-CLI-006). |
| Stream document | One NDJSON line wrapping a Volume 10 event envelope, emitted by streaming commands before their result envelope (ADR-101). |
| Destructive confirmation | A CLI-local consent checkpoint for operations that destroy or overwrite user-visible state; distinct from permission Approvals (FR-CLI-010). |
| Shared verb vocabulary | The closed subcommand verb set (`list`, `show`, `add`, `remove`, `install`, `uninstall`, `enable`, `disable`, `test`, `status`, `validate`, `search`, `export`) of ADR-100. |
| Extension mount (`x` namespace) | The reserved command group under which all extension-contributed commands appear (ADR-104). |
| Update notice | The throttled, suppressible post-command stderr line announcing a cached newer release (chapter 06). |
| Term | Meaning |
|---|---|
| Screen | One full-frame TUI view dedicated to a concern (session, plan, diff, …); exactly one is active (ADR-107) |
| Panel | A rectangular region of a screen with title, border, optional viewport, and focus state |
| Layout class | One of `wide` / `standard` / `compact` — the size-derived arrangement a screen renders in (FR-TUI-002) |
| Go-to chord | A `g`-prefixed two-key sequence addressing a screen directly (FR-TUI-004) |
| Degradation tier | One of `truecolor` / `ansi256` / `ansi16` / `none` — the color capability level frames render at (FR-UX-041) |
| Derived palette | The closed set of theme values derived from the five brand tokens (chapter 08) |
| Splash screen | The identity screen rendering the mascot, wordmark, and tagline per `tui.splash` policy (FR-UX-043) |
| Action registry | The single TUI catalog of named, context-predicated, permission-annotated operations that palette, quick actions, keymap, and help render from (ADR-110). |
| Management frame | The shared list-pane/detail-pane layout every platform screen instantiates (chapter 10). |
| View state | One of the six canonical per-panel UI states: `loading`, `content`, `empty`, `error`, `offline`, `degraded` (FR-UX-074). |
| Glyph set | One of the two closed chrome character inventories, `unicode` or `ascii`, with declared parity (ADR-112). |
| Accessible output mode | The linear, append-only rendering profile for assistive and transcript use (ADR-111, FR-TUI-065). |
| Confirmation tier | The three-level destructiveness-bound confirmation scheme: none / default-No modal / typed name (FR-UX-072). |
| Render profile | The resolved per-process rendering facts: color tier, glyph set, motion, and accessibility flags (ADR-114). |
| Toast | A non-focus-stealing, severity-classed notice with bounded stacking and error double-recording (FR-UX-073). |

## Assumptions

Local list per Volume 0 chapter 05 (global numbers at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | The `CI` environment variable is set truthy by the CI systems users run Andromeda in (de-facto convention, not a standard) | Field reports; CI-vendor documentation review at Beta (ADR-102 review condition) | Explicit tiers (`--no-input`, `ANDROMEDA_NO_INPUT`) already guarantee correctness; extend detection per open question V8A-OQ-1 |
| Technical assumption | The `NO_COLOR` convention (any non-empty value disables color) remains stable in the ecosystem | Convention page and ecosystem behavior review at Beta | Adjust the one-line check in the ADR-103 resolution; decision table under change control |
| Technical assumption | `PAGER`, `VISUAL`/`EDITOR`, and `TERM` semantics follow their established Unix conventions on Tier 1 platforms | Tier 1 platform test matrix | PAL-level adjustment; command behavior tables unchanged |
| Technical assumption | A six-character ULID prefix is a usable and sufficiently collision-resistant interactive identifier shorthand | Usage friction reports; collision incidence in fixtures | Raise the minimum prefix length (additive tightening of FR-CLI-001 rule 4) |
| Product hypothesis | A closed shared verb vocabulary with no convenience aliases yields higher first-guess correctness than muscle-memory alias compatibility | Usability passes at Beta (ADR-100 review condition) | Introduce governed aliases through the FR-CLI-001 change procedure |

1. **`COLORTERM` convention.** `COLORTERM=truecolor|24bit` is a widely deployed but
   unstandardized convention; the tier ladder (ADR-106) treats it as a positive signal
   only. If the convention decays, one ladder rung changes.
2. **Reference palettes for quantized tiers.** Contrast figures for the ansi256 tier use
   the standard xterm 256-color cube/grayscale values; ansi16 reference figures use the
   de-facto VGA palette. User-redefined palettes vary (RISK-UX-040).
3. **Modifier-bypass selection.** Most terminals bypass application mouse mode while a
   platform modifier is held, restoring native text selection; treated as a convenience,
   with `tui.mouse = false` as the guaranteed path (ADR-109).
4. **WCAG 2.x contrast arithmetic** is the accepted accessibility criterion for terminal
   text rendering; ratios in chapter 08 are computed with the WCAG relative-luminance
   formula.

Local list per Volume 0 chapter 05 (global numbers at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | Locale variables (`LC_ALL`/`LC_CTYPE`/`LANG`) are a sufficient UTF-8 signal for glyph-set auto-selection | Compatibility suite across the Tier A set and the Linux console | Extend the ADR-112 resolution with additional signals via its review conditions; `tui.glyphs` override already covers users |
| Technical assumption | Bracketed paste is available across the Tier A terminal set | Compatibility suite paste fixtures | Paste guard (FR-UX-075) still applies; unguarded terminals documented in the matrix |
| Technical assumption | The 33/100 ms coalescing windows satisfy the SM-08 ≤ 50 ms p95 added-overhead budget | SM-08 benchmark harness (Volume 12) | Tune windows; the mechanism, not the requirement, changes |
| Technical assumption | teatest/v2 can drive remote-PTY and multiplexer scenarios for FR-TUI-068 verification | Volume 13 harness spike at suite build-out | Supplement with a PTY-level harness; verification method amended, requirements unchanged |

## Open questions

Every PENDING VALIDATION in chapters 00–06 maps to a row here.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V8A-OQ-1 | Detection of vendor-specific CI environment variables beyond `CI` (which variables, which vendors) — PENDING VALIDATION | FR-CLI-009 (chapter 02); ADR-102 | No — explicit tiers and the TTY probe are sufficient for correctness | Survey major CI vendors' documented environments at Beta; extend the FR-CLI-009 ladder additively if warranted | Open |

1. **OSC 11 background-query support matrix — PENDING VALIDATION.** Theme mode `auto`
   (FR-TUI-008, ADR-106) queries the terminal background color via OSC 11 with a 50 ms
   budget and defaults to dark when unanswered. The support matrix across Tier 1-relevant
   emulators, multiplexers, and SSH paths is unvalidated; chapter 12's terminal matrix
   work should confirm it. Until validated, the dark default keeps behavior defined
   everywhere.
2. **Multiplexer truecolor passthrough.** Whether default tmux/screen configurations pass
   `COLORTERM` through (affecting how often the conservative ansi256 landing occurs) is
   environment-dependent; to be characterized alongside the chapter 12 matrix. No
   normative behavior depends on the answer — only fidelity frequency.

Every PENDING VALIDATION in chapters 10–12 and ADRs 110–114 maps to a row here.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V8C-OQ-1 | Screen-reader sufficiency of accessible output mode with VoiceOver + Terminal.app and Orca + GNOME Terminal — PENDING VALIDATION | ADR-111; FR-TUI-065; RISK-TUI-071 | No — mode contract is normative regardless; validation gates Beta accessibility claims | Structured AT test pass at Beta with recorded findings; rendering adjustments before v1 | Open |
| V8C-OQ-2 | OSC 52 availability, size limits, and multiplexer passthrough behavior per reference terminal — PENDING VALIDATION per terminal | ADR-113; FR-UX-075; compatibility matrix | No — native path and visible failure cover the gaps | Volume 13 compatibility suite verifies per terminal per release; matrix cells updated | Open |
| V8C-OQ-3 | Expected capability cells of the terminal compatibility matrix (color, mouse, paste, OSC 52 per emulator) — PENDING VALIDATION until first verified suite run | Chapter 12 matrix; ADR-114; NFR-TUI-070 | No — runtime probing (ADR-114) never trusts the matrix | Compatibility suite replaces expectations with verified values; Tier A mismatches gate releases | Open |

## Cross-volume references

Volume 0: exit-code scheme, error envelope fields, event grammar, identifier ownership.
Volume 1: PRD objectives, use cases, MVP minimum items 1/22/23/27, tension rules, SM-05/06/
07/08/18/19/20 bindings, signing viability note. Volume 2: frozen state vocabularies
rendered verbatim (Session, Run, Plan, Task, Tool Invocation, Authentication Session,
Provider connection, Plugin, MCP Client Connection, Package installation, Index, Update);
canonical export forms. Volume 3: L4 driver position, Runtime API, port signatures consumed
(ConfigPort, AuthPort, GitPort, MemoryStorePort, IndexerPort, UpdaterPort, PackagePort,
EventBusPort, PermissionPort), FR-ARCH-003/004 discipline. Volume 4: run/plan/session
semantics behind `run`/`plan`/`session`/`workflow` (keystones FR-AGT-001, FR-WF-001).
Volume 5: provider contract and capability enum behind `provider`/`model` (FR-PROV-001);
authentication flows behind `auth` (FR-AUTH-001). Volume 6: tool/plugin/skill/MCP contracts
behind their command groups (FR-TOOL-001, FR-PLUG-001, FR-SKILL-001, FR-MCP-001); extension
manifest for ADR-104 mounting. Volume 7: memory/context/index semantics behind their groups
(FR-MEM-001, FR-CTX-001, FR-IDX-001). Volume 9: permission model and names (FR-SEC-100),
secret storage (FR-SEC-102), redaction rules. Volume 10: configuration schema and
precedence (FR-CFG-001), event envelope (FR-OBS-001), logging and trace query surfaces.
Volume 11: Git Engine semantics behind `git` (FR-GIT-001); commit-content rules (ADR-015).
Volume 12: SM-06/07/08 formalization, completion and listing latency budgets. Volume 13:
the CLI test suite this fragment's verification methods name. Volume 14: update semantics,
channels, release verification behind `update` (FR-REL-001). Within Volume 8: fragments B/C
(TUI architecture FR-TUI-001, theming, wireframes, interaction patterns, accessibility)
consume chapter 01's hand-off contract and chapter 02's conventions.

- **Volume 2 chapter 09**: all frozen state vocabularies rendered by TUI screens
  (Session, Run, Plan, Task, Tool Invocation, Approval, Index) and recorded status
  vocabularies (Patch, Tool Result, Memory Record, Credential).
- **Volume 3**: TUI component boundaries (chapter 06); EventBusPort, PermissionPort,
  GitPort, WorkspacePort, ToolPort, ConfigPort, MemoryStorePort, IndexerPort contracts;
  FR-ARCH-003/FR-ARCH-004 port disciplines.
- **Volume 4**: run/turn semantics, plan approval and revision semantics, budget policy.
- **Volume 5**: provider/model identity in headers, cost-source honesty rules rendered by
  the costs screen.
- **Volume 6**: tool declaration fields rendered by the tool-call screen (origin, trust
  level, timeout, risks).
- **Volume 7**: context assembly transparency (context screen), memory layers and
  provenance (memory screen), token estimation markers.
- **Volume 9**: permission enum, scope, and decision names rendered verbatim in the
  permission prompt; Approval machine ownership; redaction rules.
- **Volume 10**: event envelope and delivery semantics for the `tui.*` family;
  configuration schema/precedence/env mapping for `[tui]`/`[tui.theme]`; log pipeline
  behind the logs screen.
- **Volume 11**: Git operation semantics and no-silent-destructive-operations rules
  behind the git and diff screens.
- **Volume 12**: formalization of SM-06(b), SM-07, SM-08 budgets this volume's TUI
  requirements reference as constraints.
- **Volume 13**: teatest/v2 golden-frame suites, TUI load/fault suites, permission-
  mediation tests (SM-16(b)).
- **ADR-026** (Volume 0 register): its Danger PENDING VALIDATION item is resolved by
  ADR-105; the ADR-026 review condition "resolve when Volume 8 fixes the mapping" is
  satisfied at consolidation.
- **Chapter 02 of this volume (fragment A)**: ADR-103 styling resolution consumed by the
  tier ladder; E-CLI-005 and FR-CLI-003 govern the CLI side of the hand-off boundary.
- **Chapters 10–12 of this volume (fragment C)**: platform wireframes, interaction
  patterns (search, clipboard, virtualization, empty/loading/offline/degraded states),
  accessibility and terminal compatibility matrix referenced throughout chapters 07–09.

Volume 2: frozen state vocabularies rendered verbatim in every screen (Provider
connection, Plugin, MCP Client Connection, Workflow Run, Update process, `interrupted`
semantics). Volume 3: driver rules and component boundaries (chapter 06); PAL surfaces
consumed — Clipboard, Notifications, PTY (chapter 07); reference emulators, shells, and
constrained-environment contracts; FR-ARCH-004 cancellation; SchedulerPort-independent
render loop per ADR-006. Volume 4: run/plan/task semantics behind the workflow and
recovery screens; resume rules (PRD-010). Volume 5: capability enum names rendered in
provider/model screens; provider/model change notification rule. Volume 6: plugin, skill,
MCP lifecycles and trust vocabularies behind their screens; PackagePort flows. Volume 9:
permission enum names in the action registry (`network`, `clipboard`,
`package_installation`, `system_modification`, `external_service_access`); approval
presentation semantics; redaction rules applied to copies and envelopes. Volume 10: event
envelope for `tui.*` events; configuration schema/precedence for the `[tui]` keys;
ConfigPort source attribution rendered by the configuration screen. Volume 12:
formalization of SM-06/07/08/09 targets these chapters' mechanisms serve; reference
hardware for NFR-UX-077/078. Volume 13: TUI suite, compatibility suite, crash-injection
(SM-11), offline suite (SM-05). Volume 14: Update process semantics behind the update
screen. Within Volume 8: fragment A's CLI conventions (ADR-102/103, FR-CLI-009/010,
FR-UX-001/002/003, E-CLI-005); fragment B's shell contract, theming tiers, and E-TUI
family (chapters 07–09).
