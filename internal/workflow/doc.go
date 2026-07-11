// Package workflow is layer L3: the Workflow Engine implementing specification-driven
// development (Volume 4, FR-WF-001). It executes a workflow definition — an ordered set of
// stages, some gated by human approval — driving the frozen Workflow Run state machine
// (pending → running → awaiting_approval → completed/failed/cancelled/interrupted), emitting
// stage events, and supporting resume from a stage boundary. The 14-stage SDD pipeline is a
// built-in definition.
package workflow
