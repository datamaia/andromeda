package eventbus

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// BenchmarkPublish measures event-bus publish-and-deliver overhead with one live subscriber
// draining in the background. Micro-benchmark tier, Volume 12 chapter 03 (ADR-012 in-process bus).
func BenchmarkPublish(b *testing.B) {
	ctx := context.Background()
	bus := New()
	defer bus.Close()

	sub, err := bus.Subscribe(ctx, ports.TopicSelector{Prefixes: []string{"bench."}}, ports.SubscribeOptions{})
	if err != nil {
		b.Fatal(err)
	}
	defer sub.Close()

	drained := make(chan struct{})
	go func() {
		for {
			if _, err := sub.Events().Next(ctx); err != nil {
				close(drained)
				return
			}
		}
	}()

	ev := NewEvent("bench.tick", "benchmark")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := bus.Publish(ctx, ev); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}
