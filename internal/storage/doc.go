// Package storage is layer L2 infrastructure: the SQLite-backed persistence foundation
// (ADR-007, ADR-028, ADR-029). It opens databases in WAL mode through the pure-Go
// modernc.org/sqlite driver, runs forward-only migrations with a pre-migration backup and
// post-migration integrity check, and refuses to open databases written by a newer schema.
//
// ADR-028 database split: one workspace database at <workspace>/.andromeda/state.db and one
// global database at <data_dir>/global.db. This package provides both; higher layers (Session
// store, Memory store, Event store) build their schemas as migrations on top.
//
// Layer L2 may import internal/core and internal/ports.
package storage
