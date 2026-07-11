// Package arch holds the architecture dependency manifest and its enforcement (ADR-033).
//
// The layer manifest is the single source of truth for the allowed/prohibited dependency
// matrix (Volume 3, chapter 01). The import-graph test in this package enforces it over the
// whole module with no external tooling, so the rule holds even where golangci-lint's depguard
// is unavailable. golangci-lint mirrors a subset for editor feedback.
package arch

// Layer is an architecture layer, lowest (L0) to highest.
type Layer int

const (
	L0Core           Layer = iota // pure domain
	L1Contract                    // ports, platform abstraction
	L2Infrastructure              // persistence, event bus, config, telemetry, secret store, ...
	L3Engine                      // agent/execution/workflow/provider/tool/... engines
	L4Application                 // runtime composition, IPC server
	L5Driver                      // CLI, TUI
)

// PackageLayer maps an internal package's import path suffix (after
// "github.com/datamaia/andromeda/internal/") to its layer. Packages absent from the map are
// unclassified and rejected by the enforcement test, so adding a package forces a manifest
// entry in the same change (FR-ARCH-001).
var PackageLayer = map[string]Layer{
	"core":      L0Core,
	"buildinfo": L0Core, // build metadata is pure, dependency-free
	"ports":     L1Contract,
	"pal":       L1Contract,
	"arch":      L1Contract, // the manifest itself is a contract artifact
	"storage":   L2Infrastructure,
	"config":    L2Infrastructure,
	"streams":   L2Infrastructure,
}

// ModulePath is the module's import path prefix.
const ModulePath = "github.com/datamaia/andromeda"

// InternalPrefix is the prefix of internal package import paths.
const InternalPrefix = ModulePath + "/internal/"
