package eventbus

import (
	"context"
	"testing"
	"time"

	"github.com/datamaia/andromeda/internal/ports"
)

func TestPublishSubscribeByPrefix(t *testing.T) {
	ctx := context.Background()
	b := New()
	defer b.Close()

	sub, err := b.Subscribe(ctx, ports.TopicSelector{Prefixes: []string{"run."}}, ports.SubscribeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	if err := b.Publish(ctx, NewEvent("run.completed", "runtime")); err != nil {
		t.Fatal(err)
	}
	if err := b.Publish(ctx, NewEvent("tool.invocation.denied", "tool-runtime")); err != nil {
		t.Fatal(err)
	}

	got, err := sub.Events().Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "run.completed" {
		t.Fatalf("received %q, want run.completed (tool.* should not match run. prefix)", got.Name)
	}
}

func TestExactNameSelector(t *testing.T) {
	ctx := context.Background()
	b := New()
	defer b.Close()
	sub, _ := b.Subscribe(ctx, ports.TopicSelector{Names: []string{"provider.request.failed"}}, ports.SubscribeOptions{})
	_ = b.Publish(ctx, NewEvent("provider.request.failed", "provider"))
	ev, err := sub.Events().Next(ctx)
	if err != nil || ev.Name != "provider.request.failed" {
		t.Fatalf("got %q,%v", ev.Name, err)
	}
}

func TestOverflowDropOldestDoesNotBlockPublisher(t *testing.T) {
	ctx := context.Background()
	b := New()
	defer b.Close()
	sub, _ := b.Subscribe(ctx, ports.TopicSelector{}, ports.SubscribeOptions{BufferSize: 4})
	// Publish far more than the buffer; publisher must never block.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			_ = b.Publish(ctx, NewEvent("area.thing.happened", "test"))
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publisher blocked on a full subscriber buffer")
	}
	s := sub.(*subscription)
	if s.Dropped() == 0 {
		t.Error("expected some events to be dropped by the overflow policy")
	}
}

func TestPublishToClosedBus(t *testing.T) {
	b := New()
	_ = b.Close()
	err := b.Publish(context.Background(), NewEvent("area.thing.happened", "t"))
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != "E-OBS-001" {
		t.Fatalf("want E-OBS-001, got %v", err)
	}
}

func TestContextCancelClosesSubscription(t *testing.T) {
	b := New()
	defer b.Close()
	ctx, cancel := context.WithCancel(context.Background())
	sub, _ := b.Subscribe(ctx, ports.TopicSelector{}, ports.SubscribeOptions{})
	cancel()
	// Give the auto-close goroutine a moment, then Next should end.
	time.Sleep(20 * time.Millisecond)
	if _, err := sub.Events().Next(context.Background()); err != ports.ErrEndOfStream {
		t.Errorf("want ErrEndOfStream after context cancel, got %v", err)
	}
}

func TestEventNameGrammar(t *testing.T) {
	valid := []string{"run.completed", "tool.invocation.denied", "provider.request.failed", "index.build.started"}
	for _, n := range valid {
		if !ValidName(n) {
			t.Errorf("%q should be valid", n)
		}
	}
	invalid := []string{"Run.Completed", "run", "run.", ".run", "a.b.c.d", "run completed"}
	for _, n := range invalid {
		if ValidName(n) {
			t.Errorf("%q should be invalid", n)
		}
	}
}
