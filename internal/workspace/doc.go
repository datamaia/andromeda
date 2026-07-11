// Package workspace is layer L3: the Workspace Engine implementing ports.WorkspacePort. It
// discovers workspace roots (the .andromeda/ marker or a repository root), opens workspaces
// (initializing .andromeda/ and the workspace database, ADR-028, and registering them in the
// machine-global registry), takes consistent read-only snapshots for run reproducibility
// (SM-12) and context assembly, and closes them cleanly. Entity semantics are Volume 2's;
// behavior is Volume 4's.
package workspace
