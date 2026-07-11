# Security Policy

## Reporting a vulnerability

**Do not report security vulnerabilities through public GitHub issues.**

Please use GitHub's [private vulnerability reporting](https://docs.github.com/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability)
for this repository (Security → Advisories → Report a vulnerability).

We commit to a **first response within 3 business days** (SM-16(c)). After triage we will
agree a disclosure timeline with you and credit your report unless you prefer to remain
anonymous.

## Scope

Vulnerabilities in Andromeda itself: the runtime, permission model, sandbox, secret
handling, provider adapters, tool/plugin/MCP runtimes, and the release/update path.
Findings in third-party dependencies should be reported upstream; tell us too if Andromeda's
use of them is affected.

## What to include

- A description of the issue and its impact.
- Steps or a proof of concept to reproduce it.
- Affected version(s) or commit.
- Any known mitigations.

## Handling and response

Incident response, forensic preservation, and recovery are specified in
[Volume 9, chapter 08](docs/spec/volume-09-security/08-audit-and-incident-response.md).
Coordinated-disclosure governance is defined in
[Volume 15](docs/spec/volume-15-roadmap-and-execution/04-open-source-governance.md).
