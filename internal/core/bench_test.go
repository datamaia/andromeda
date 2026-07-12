package core

import "testing"

// BenchmarkNewULID measures identifier-generation overhead. Every entity reference that crosses
// a port mints a ULID (ADR-027), so this is on a very hot path. Micro-benchmark tier, Volume 12
// chapter 03.
func BenchmarkNewULID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewULID()
	}
}
