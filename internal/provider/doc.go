// Package provider is layer L3: the Provider Layer. It defines the shared error family and
// HTTP helpers every model-provider adapter uses, and the Router that itself implements
// ports.ProviderPort so consumers are indifferent to routing and fallback (FR-PROV-001,
// Principle 1). Concrete adapters live in subpackages (openaicompat, anthropic, ollama) and
// speak documented public APIs only (ADR-019); no provider-specific logic leaks outside them.
package provider
