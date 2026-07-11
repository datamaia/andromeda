// Package memory is layer L3: the Memory Manager implementing ports.MemoryStorePort (Volume 7,
// FR-MEM-001). It persists memory records across the session/workspace/long-term/semantic/
// episodic layers in the workspace database (ADR-028) with provenance and retention, and
// retrieves them lexically. Semantic retrieval via the Indexer (ADR-020) is layered on top in a
// later increment. Memory MUST NOT store secrets without explicit authorization; redaction
// happens upstream (Volume 9).
package memory
