package core

// ULID is the canonical 26-character Crockford base32 identifier used for every entity
// reference that crosses a port (ADR-027). It is a string alias so it interoperates freely
// with JSON, SQL, and log fields while documenting intent at call sites.
type ULID = string

// Phase is a delivery phase from the Volume 1 phase model.
type Phase string

// Phase enumeration from the Volume 1 delivery-phase model.
const (
	PhaseCore       Phase = "Core"
	PhaseMVP        Phase = "MVP"
	PhaseBeta       Phase = "Beta"
	PhaseV1         Phase = "v1"
	PhaseV2         Phase = "v2"
	PhaseFuture     Phase = "Future"
	PhaseOutOfScope Phase = "Out of Scope"
)
