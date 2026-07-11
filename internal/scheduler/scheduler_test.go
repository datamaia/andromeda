package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/datamaia/andromeda/internal/ports"
)

func TestSubmitAndAwait(t *testing.T) {
	ctx := context.Background()
	s := New(nil)
	h, err := s.Submit(ctx, ports.TaskSpec{Pool: "tools", Run: func(context.Context) (ports.TaskOutcome, error) {
		return ports.TaskOutcome{OK: true, Message: "done"}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	out, err := h.Await(ctx)
	if err != nil || !out.OK || out.Message != "done" {
		t.Fatalf("outcome = %+v err=%v", out, err)
	}
}

func TestPanicIsCaptured(t *testing.T) {
	ctx := context.Background()
	s := New(nil)
	h, _ := s.Submit(ctx, ports.TaskSpec{Run: func(context.Context) (ports.TaskOutcome, error) {
		panic("boom")
	}})
	out, err := h.Await(ctx)
	if err == nil || out.OK {
		t.Fatalf("panic should surface as error: out=%+v err=%v", out, err)
	}
}

func TestConcurrencyBoundedByPool(t *testing.T) {
	ctx := context.Background()
	s := New(map[string]int{"tools": 2})
	var concurrent, maxConcurrent int64
	start := make(chan struct{})
	var handles []ports.TaskHandle
	for i := 0; i < 6; i++ {
		h, _ := s.Submit(ctx, ports.TaskSpec{Pool: "tools", Run: func(context.Context) (ports.TaskOutcome, error) {
			n := atomic.AddInt64(&concurrent, 1)
			for {
				m := atomic.LoadInt64(&maxConcurrent)
				if n <= m || atomic.CompareAndSwapInt64(&maxConcurrent, m, n) {
					break
				}
			}
			<-start
			atomic.AddInt64(&concurrent, -1)
			return ports.TaskOutcome{OK: true}, nil
		}})
		handles = append(handles, h)
	}
	time.Sleep(50 * time.Millisecond)
	close(start)
	for _, h := range handles {
		h.Await(ctx)
	}
	if maxConcurrent > 2 {
		t.Errorf("pool concurrency exceeded limit: %d > 2", maxConcurrent)
	}
}

func TestGroupFirstErrorPropagates(t *testing.T) {
	ctx := context.Background()
	s := New(nil)
	g, _ := s.NewGroup(ctx, ports.GroupSpec{Pool: "background"})
	g.Go(ctx, ports.TaskSpec{Run: func(context.Context) (ports.TaskOutcome, error) {
		return ports.TaskOutcome{OK: true}, nil
	}})
	g.Go(ctx, ports.TaskSpec{Run: func(context.Context) (ports.TaskOutcome, error) {
		return ports.TaskOutcome{OK: false}, errors.New("failed member")
	}})
	if err := g.Wait(ctx); err == nil {
		t.Fatal("expected the group to surface the first error")
	}
}

func TestCancelBeforeStart(t *testing.T) {
	ctx := context.Background()
	// A pool of size 1, saturated, so the second task queues and can be cancelled before start.
	s := New(map[string]int{"io": 1})
	block := make(chan struct{})
	s.Submit(ctx, ports.TaskSpec{Pool: "io", Run: func(context.Context) (ports.TaskOutcome, error) {
		<-block
		return ports.TaskOutcome{OK: true}, nil
	}})
	h2, _ := s.Submit(ctx, ports.TaskSpec{Pool: "io", Run: func(context.Context) (ports.TaskOutcome, error) {
		return ports.TaskOutcome{OK: true}, nil
	}})
	_ = s.Cancel(ctx, h2.ID(), ports.CancelUserInterrupt)
	out, err := h2.Await(ctx)
	if err == nil && out.OK {
		t.Error("cancelled-before-start task should not report success")
	}
	close(block)
}

func TestStatsAndShutdown(t *testing.T) {
	ctx := context.Background()
	s := New(nil)
	st, _ := s.Stats(ctx)
	if _, ok := st.Pools["tools"]; !ok {
		t.Error("expected a tools pool in stats")
	}
	s.Shutdown()
	if _, err := s.Submit(ctx, ports.TaskSpec{}); err == nil {
		t.Error("submit after shutdown should be rejected")
	}
}
