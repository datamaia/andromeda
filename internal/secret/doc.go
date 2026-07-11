// Package secret is layer L3: the Secret Store implementing ports.SecretStorePort (FR-SEC-102,
// ADR-014). Credential material lives behind references (secret_ref); only this package touches
// material. The primary backend is the OS keychain (via the PAL CredentialStore); an opt-in,
// explicitly lower-security age-encrypted file fallback (ADR-014) serves platforms without a
// keychain. All operations are local-only — no network (Principle 3). Secrets never appear in
// logs, events, or errors (Volume 9 redaction).
package secret
