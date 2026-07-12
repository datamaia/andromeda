// Package mcp is layer L3: the Model Context Protocol client (Volume 6, FR-MCP-001, ADR-010).
// It speaks MCP over the newline-delimited JSON-RPC 2.0 stdio transport (internal/jsonrpc):
// initialize, tools/list, and tools/call. Discovered MCP tools are bridged to ports.ToolPort so
// the Tool Runtime mediates them (permissions, trust) exactly like built-in tools. The same
// transport carries the Andromeda Runtime Protocol for plugins (ADR-009).
package mcp
