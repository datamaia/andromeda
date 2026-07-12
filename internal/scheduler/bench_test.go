package scheduler

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// BenchmarkSubmitAwait measures scheduler submission overhead: submit a trivial task and await
// its outcome. Relates to NFR-PERF-020 (concurrency capacity and scheduler overhead).
// Micro-benchmark tier, Volume 12 chapter 03.
func BenchmarkSubmitAwait(b *testing.B) {
	ctx := context.Background()
	s := New(nil)
	spec := ports.TaskSpec{Pool: "tools", Run: func(context.Context) (ports.TaskOutcome, error) {
		return ports.TaskOutcome{OK: true}, nil
	}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h, err := s.Submit(ctx, spec)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := h.Await(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
