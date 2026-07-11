// Package permission is layer L3: the Permission Manager implementing ports.PermissionPort
// (FR-SEC-100). It is the single decision path for side-effecting actions (Principle 8). The
// closed enums (13 permissions, 10 scope qualifiers, 7 decisions) live in internal/core; this
// package implements the evaluation algorithm (deny > ask > allow > else ask; fail-closed per
// ADR-125), grant storage, decision persistence, and audit. Evaluation never touches the
// network (PRD-003).
package permission
