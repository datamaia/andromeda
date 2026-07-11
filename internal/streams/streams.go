// Package streams provides reusable ports.Stream[T] implementations (slice-backed and
// channel-backed). Layer L2 infrastructure; imports internal/ports only.
package streams

import (
	"context"
	"sync"

	"github.com/datamaia/andromeda/internal/ports"
)

// Slice returns a Stream that yields the given items in order, then ports.ErrEndOfStream.
func Slice[T any](items []T) ports.Stream[T] {
	return &sliceStream[T]{items: items}
}

type sliceStream[T any] struct {
	mu     sync.Mutex
	items  []T
	i      int
	closed bool
}

func (s *sliceStream[T]) Next(ctx context.Context) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.i >= len(s.items) {
		return zero, ports.ErrEndOfStream
	}
	v := s.items[s.i]
	s.i++
	return v, nil
}

func (s *sliceStream[T]) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	return nil
}

// Chan returns a channel-backed Stream plus a send function and a done signal. Sending after
// Close or ctx cancellation is a no-op. The stream ends when the send function's returned
// close is called (draining remaining buffered items first) or the context is cancelled.
func Chan[T any](buffer int) (ports.Stream[T], func(T) bool, func()) {
	ch := make(chan T, buffer)
	done := make(chan struct{})
	var once sync.Once
	closeFn := func() { once.Do(func() { close(done) }) }
	send := func(v T) bool {
		// Check closure first so a closed stream deterministically rejects sends even when
		// the buffer still has space (a plain select would pick a ready case at random).
		select {
		case <-done:
			return false
		default:
		}
		select {
		case <-done:
			return false
		case ch <- v:
			return true
		}
	}
	return &chanStream[T]{ch: ch, done: done, closeFn: closeFn}, send, closeFn
}

type chanStream[T any] struct {
	ch      chan T
	done    chan struct{}
	closeFn func()
}

func (s *chanStream[T]) Next(ctx context.Context) (T, error) {
	var zero T
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case v, ok := <-s.ch:
		if !ok {
			return zero, ports.ErrEndOfStream
		}
		return v, nil
	case <-s.done:
		// Drain any buffered items before ending.
		select {
		case v, ok := <-s.ch:
			if ok {
				return v, nil
			}
		default:
		}
		return zero, ports.ErrEndOfStream
	}
}

func (s *chanStream[T]) Close() error {
	s.closeFn()
	return nil
}
