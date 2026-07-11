// Package git is layer L3: the Git Engine implementing ports.GitPort (FR-GIT-001, ADR-025).
// It shells out to the system git (>= 2.40) behind an encapsulated adapter, parsing porcelain
// and NUL-terminated formats. Mutating methods are side-effecting: callers (Tool Runtime,
// drivers) MUST hold a PermissionPort decision before invoking them and produce the Volume 2
// File Change / Patch / Command Execution records; the engine performs no silent destructive
// operations and reports partial completion honestly.
package git
