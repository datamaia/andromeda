# Volume 9 — Security

**Status:** Complete · **Owner:** Security (Volume 9)

Volume 9 specifies Andromeda's security posture end to end: the threat model over the
product's assets, trust boundaries, actors, and attack vectors; the permission model
(closed enums for permissions, scope qualifiers, and decisions — the single decision path
for every side-effecting action); the sandbox specification implementing ADR-021's layered
containment across five execution tiers; the credential and secret management model
implementing ADR-014 with corpus-wide redaction rules; Audit Log semantics with the
audited-action catalog, hash-chain verification, and incident response; and the full
Approval state machine. Volume 9 mints identifiers in the `SEC` area and ADRs 115–129.
Per the single-home matrix, this volume is the authoritative home of the permission model,
the sandbox model, and credential storage (provider auth flows are Volume 5's).

## Chapters

| Chapter | Contents | Status |
|---|---|---|
| 01 — Threat Model Overview (`01-threat-model-overview.md`) | Assets, trust boundaries, actors, attack vectors, risk matrix | Complete |
| 02 — Injection Threats (`02-threats-injection.md`) | Prompt injection, indirect prompt injection, tool/command injection, path traversal, symlink attacks, malicious model output, and related threats | Complete |
| 03 — Extension and Supply-Chain Threats (`03-threats-extensions-supply-chain.md`) | Tool poisoning, MCP poisoning, malicious plugins/skills/repositories/files, dependency and supply-chain attacks, CI/release/update compromise | Complete |
| 04 — System Threats (`04-threats-system.md`) | Secret exfiltration, credential theft, sandbox escape, privilege escalation, compromised providers and local models, log leakage, memory and index poisoning, social engineering | Complete |
| [05 — Permission Model](05-permission-model.md) | Keystone FR-SEC-100: permission/scope/decision enums, evaluation precedence and inheritance, revocation, persistence, audit binding | Complete |
| [06 — Sandbox Specification](06-sandbox-specification.md) | Keystone FR-SEC-101: policy model, five sandbox tiers, layered enforcement per ADR-021, filesystem/env/secret filtering, limits, symlinks, temp dirs, cleanup | Complete |
| [07 — Credential and Secret Management](07-credential-and-secret-management.md) | Keystone FR-SEC-102: Secret Store backends per ADR-014, encrypted-file fallback consent, fingerprints, orphan sweep, redaction at every sink | Complete |
| [08 — Audit and Incident Response](08-audit-and-incident-response.md) | Audited-action catalog, chain verification and tamper response, retention and export, security events, incident procedure, disclosure hooks (governance in Volume 15) | Complete |
| [09 — Approval State Machine](09-approval-state-machine.md) | Full machine for Approval under the frozen state names, with all twelve mandatory elements | Complete |
| 99 — Volume Register (`99-volume-register.md`) | Everything Volume 9 minted; assembled from the per-agent register fragments at consolidation | Pending merge |

## Reading order and dependencies

Chapters 01–04 (threat model) motivate everything that follows and are best read first;
chapters 05–08 define the enforcement mechanisms the threats' Prevention and Response rows
bind to; chapter 09 closes with the consent machine both 05 and the rest of the corpus block
on. Prerequisites from other volumes: Volume 2 chapters 04, 08, and 09 (Permission,
Approval, Audit Record entities and frozen states), Volume 3 chapter 02 (PermissionPort,
SandboxPort, SecretStorePort), and ADR-014/ADR-021 for the mechanism decisions this volume
elaborates. Downstream consumers: Volume 4 (approvals gating runs and workflows), Volume 5
(credential flows over the storage model), Volume 6 (tool/plugin/MCP permission and sandbox
binding), Volume 7 (memory redaction constraints), Volume 8 (permission prompts, non-
interactive denial semantics, incident surfaces), Volume 10 (config tables `[permissions]`,
`[sandbox]`, `[security]`; event envelope), Volume 11 (Git mutation gates), Volume 13
(security suites), Volume 14 (update verification and incident triggers).
