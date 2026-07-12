// Package jsonrpc is layer L2 infrastructure: a minimal JSON-RPC 2.0 client over a
// newline-delimited byte stream. It is the shared transport for the Model Context Protocol
// stdio transport (ADR-010) and the Andromeda Runtime Protocol for plugins (ADR-009). Messages
// are single-line JSON objects terminated by '\n' and MUST NOT contain embedded newlines, per
// the MCP stdio framing. A background read loop demultiplexes responses to concurrent callers.
package jsonrpc
