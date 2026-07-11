// Package agent is layer L3: the Agent Engine implementing the plan–act–observe loop
// (Volume 4, FR-AGT-001). One mode-invariant loop drives a provider through turns: it sends
// the conversation plus tool declarations, and when the model returns tool calls it executes
// them through the mediated Tool Runtime (permissions, sandbox, denial-as-data) and feeds the
// results back, iterating until the model answers without tool calls or a budget is reached.
// Runs are persisted through SessionStorePort so work is recoverable (PRD-010).
package agent
