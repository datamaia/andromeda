// Package tool is layer L3: the Tool Runtime (Volume 6, FR-TOOL-001). It mediates all tool
// access — agents never hold a ports.ToolPort directly. For each invocation the Runtime
// validates input against the tool's schema, evaluates the tool's declared permissions through
// PermissionPort (denial-as-data: a refused invocation returns a terminal error ToolEvent, not
// a transport failure), then drives the tool. Built-in tools live in the builtin subpackage.
package tool
