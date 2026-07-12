package streams

import (
	"context"
	"testing"
)

// BenchmarkChanRoundtrip measures per-chunk streaming overhead: one send plus one receive
// through a channel-backed stream. Relates to NFR-PERF-007 (streaming update overhead).
// Micro-benchmark tier, Volume 12 chapter 03.
func BenchmarkChanRoundtrip(b *testing.B) {
	ctx := context.Background()
	st, send, closeFn := Chan[int](1)
	defer closeFn()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		send(i)
		if _, err := st.Next(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
