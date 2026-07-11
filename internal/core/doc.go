// Package core is layer L0 — the pure domain. It holds identifiers, closed enumerations,
// and domain invariants with no I/O and no dependency on any other internal package.
//
// The dependency rule (ADR-030, ADR-033): nothing in internal/core may import internal/ports
// or any higher layer (runtime engines, adapters, CLI/TUI). Enforcement is the import-graph
// test in internal/arch and the depguard configuration in .golangci.yml.
package core
