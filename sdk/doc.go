// Package sdk is the root of the Andromeda Extension SDK: the public, versioned contract
// surface that extension authors build against (Extension SDK component, Volume 3).
//
// The SDK mirrors the frozen port contracts of the main module's internal/ports package
// (FR-ARCH-003) without importing them — internal packages are, by Go's rules and ADR-031,
// invisible to this separate module. Concrete mirror types (ProviderExtension, ToolExtension,
// SkillManifest, PluginProtocol, …) are introduced by the Extension SDK epics; this file
// establishes the module boundary so the two-module layout and its version independence exist
// from the start.
package sdk

// Version is the SDK's own semantic version, independent of the product version (ADR-015).
const Version = "0.0.0-dev"
