// Package ports is layer L1 — the frozen port interfaces through which engines reach
// infrastructure and extensions reach Andromeda (Volume 3, chapter 02; FR-ARCH-003).
//
// Port names, method names, and signatures in this package are frozen: owning volumes
// elaborate behavioral contracts under exactly these names and MUST NOT rename, re-sign,
// split, or merge them. Contract types may gain optional fields additively.
//
// Conventions honored by every port (not restated per method):
//  1. Context first — every method takes context.Context first and honors cancellation.
//  2. Typed errors by area — adapters map failures into the port's E-<AREA> family; a caller
//     never sees a raw driver/HTTP/OS error through a port.
//  3. Streams use Stream[T]; consumers must Close; Close is idempotent.
//  4. Entity references cross ports as core.ULID (ADR-027).
//  5. No leakage — signatures use only L1 contract types and L0 core types.
//  6. Thread safety — implementations are safe for concurrent use unless a method says otherwise.
package ports

import (
	"context"
	"errors"

	"github.com/datamaia/andromeda/internal/core"
)

// ULID re-exports the domain identifier for signature fidelity with the specification.
type ULID = core.ULID

// JSON is a canonical JSON document (Volume 2, chapter 10).
type JSON = []byte

// ErrEndOfStream terminates a Stream. Consumers detect it from Next.
var ErrEndOfStream = errors.New("ports: end of stream")

// Stream is the uniform streaming result shape. Items arrive in order; Next blocks until an
// item, ErrEndOfStream, context cancellation, or failure. Close is idempotent, cancels
// production, and releases resources. Every stream must be closed by its consumer.
type Stream[T any] interface {
	Next(ctx context.Context) (T, error)
	Close() error
}

// PortError is the in-process view of the ADR-016 error envelope. Adapters return values of
// this type (wrapped as error) so callers can branch on Code and Retryable without parsing.
type PortError struct {
	Code          string // stable E-<AREA>-NNN
	Category      string
	Severity      string
	Message       string // user-facing, redacted
	Detail        string // technical, safe-to-log
	Retryable     bool
	CorrelationID core.ULID
	Cause         error
}

func (e *PortError) Error() string {
	if e.Detail != "" {
		return e.Code + ": " + e.Detail
	}
	return e.Code + ": " + e.Message
}

func (e *PortError) Unwrap() error { return e.Cause }

// Error family prefixes (ADR-016). Concrete codes are minted by each contract owner.
const (
	FamilyProvider = "E-PROV"
	FamilyAuth     = "E-AUTH"
	FamilyTool     = "E-TOOL"
	FamilyMemory   = "E-MEM"
	FamilyIndex    = "E-IDX"
	FamilyObs      = "E-OBS"
	FamilySecurity = "E-SEC"
	FamilyConfig   = "E-CFG"
	FamilyGit      = "E-GIT"
	FamilyAgent    = "E-AGT"
	FamilyArch     = "E-ARCH"
	FamilyRelease  = "E-REL"
	FamilyPlugin   = "E-PLUG"
)
