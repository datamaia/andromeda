// Package sandbox is layer L3: the Sandbox Engine implementing ports.SandboxPort (FR-SEC-101,
// ADR-021). It is the exclusive launch path for tools, plugins, and terminal commands — direct
// process spawning outside this engine is a defect. The MVP layer applies process-level
// controls (deny-by-default environment filtering, working-directory and path policy, command
// allow/deny lists, a wall-clock time limit, and process-group teardown); OS-level isolation
// (macOS Seatbelt, Linux Landlock/namespaces) is a Beta/v1 layer (PENDING VALIDATION, ADR-021).
// The effective containment level is part of a handle's observable state and is never silently
// weakened.
package sandbox
