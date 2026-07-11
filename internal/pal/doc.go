// Package pal is the Platform Abstraction Layer (Volume 3, chapter 07): the single place
// OS-specific behavior is encapsulated (ADR-030). Every platform dependency lives behind one
// of the surface interfaces here; platform checks MUST NOT be scattered through the codebase
// (FR-PORT-001). Unix is the reference behavior; a future Windows backend implements the same
// surfaces (Volume 3 chapter 07, Windows-future encapsulation rules).
//
// Layer: L1 (a contract/infra layer). pal may import internal/core only.
package pal
