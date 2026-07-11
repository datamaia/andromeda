// Package indexer is layer L3: the Indexing Engine implementing ports.IndexerPort (Volume 7,
// FR-IDX-001). This MVP provides a lexical inverted index over workspace files. Indexes are
// rebuildable caches (INV-IDX-02): they are held in memory and rebuilt from source on demand,
// never a source of data loss. The Index lifecycle uses the frozen Volume 2 states (created,
// building, ready, updating, stale, failed, removed). Semantic indexing (embeddings, ADR-020)
// is a later increment.
package indexer
