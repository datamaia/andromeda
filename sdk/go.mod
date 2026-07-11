// The Andromeda Extension SDK — the public contract surface for building providers, tools,
// skills, plugins, and other extensions (ADR-003, ADR-031). It is a separate Go module so it
// versions independently and MUST NOT import internal/ packages; it mirrors the frozen port
// contracts of internal/ports. The mirror-equivalence check (ADR-031) is a CI gate.
module github.com/datamaia/andromeda/sdk

go 1.24
