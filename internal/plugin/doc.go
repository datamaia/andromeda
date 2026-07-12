// Package plugin is layer L3: the Plugin Runtime (Volume 6, FR-PLUG-001). Plugins are external
// subprocesses that speak the Andromeda Runtime Protocol — JSON-RPC 2.0 over stdio (ADR-009),
// the same transport and method surface as MCP (ADR-010) — so the MCP client drives both. The
// runtime spawns a plugin, initializes it, bridges its tools to permission-mediated ToolPorts,
// and manages the frozen Plugin lifecycle (registered → starting → running → stopping →
// stopped; failed/disabled/removed). A subprocess is launched only through the Sandbox Engine
// in production; the transport is injectable for testing.
package plugin
