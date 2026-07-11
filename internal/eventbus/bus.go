package eventbus

import (
	"context"
	"strings"
	"sync"

	"github.com/datamaia/andromeda/internal/ports"
)

// Overflow policies for a subscriber whose buffer is full.
const (
	OverflowDropOldest = "drop_oldest" // default: discard the oldest buffered event
	OverflowDropNewest = "drop_newest" // discard the incoming event
)

// DefaultBuffer is the per-subscriber buffer size used when none is requested.
const DefaultBuffer = 256

// Bus is an in-process EventBusPort. It is safe for concurrent publishers and subscribers.
type Bus struct {
	mu     sync.RWMutex
	subs   map[*subscription]struct{}
	closed bool
}

// New returns an empty Bus.
func New() *Bus {
	return &Bus{subs: map[*subscription]struct{}{}}
}

var _ ports.EventBusPort = (*Bus)(nil)

// Publish delivers one event to every matching subscriber without blocking on slow ones.
func (b *Bus) Publish(ctx context.Context, event ports.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return &ports.PortError{Code: "E-OBS-001", Category: "observability", Message: "event bus is closed"}
	}
	targets := make([]*subscription, 0, len(b.subs))
	for s := range b.subs {
		if s.matches(event.Name) {
			targets = append(targets, s)
		}
	}
	b.mu.RUnlock()
	for _, s := range targets {
		s.enqueue(event)
	}
	return nil
}

// Subscribe registers a subscriber for a topic selector.
func (b *Bus) Subscribe(ctx context.Context, sel ports.TopicSelector, opts ports.SubscribeOptions) (ports.Subscription, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	buf := opts.BufferSize
	if buf <= 0 || buf > DefaultBuffer*16 {
		buf = DefaultBuffer
	}
	s := &subscription{
		sel:      sel,
		ch:       make(chan ports.Event, buf),
		overflow: OverflowDropOldest,
		bus:      b,
	}
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil, &ports.PortError{Code: "E-OBS-001", Message: "event bus is closed"}
	}
	b.subs[s] = struct{}{}
	b.mu.Unlock()

	// Auto-close the subscription when its context is cancelled.
	go func() {
		<-ctx.Done()
		_ = s.Close()
	}()
	return s, nil
}

// Close shuts the bus down and closes all subscriptions.
func (b *Bus) Close() error {
	b.mu.Lock()
	b.closed = true
	subs := make([]*subscription, 0, len(b.subs))
	for s := range b.subs {
		subs = append(subs, s)
	}
	b.mu.Unlock()
	for _, s := range subs {
		_ = s.Close()
	}
	return nil
}

// subscription is one registered subscriber with a bounded buffer.
type subscription struct {
	sel      ports.TopicSelector
	ch       chan ports.Event
	overflow string
	bus      *Bus

	mu      sync.Mutex
	closed  bool
	dropped int64
}

func (s *subscription) matches(name string) bool {
	if len(s.sel.Names) == 0 && len(s.sel.Prefixes) == 0 {
		return true // empty selector matches everything
	}
	for _, n := range s.sel.Names {
		if n == name {
			return true
		}
	}
	for _, p := range s.sel.Prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func (s *subscription) enqueue(e ports.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.ch <- e:
	default:
		// Buffer full: apply the overflow policy.
		if s.overflow == OverflowDropNewest {
			s.dropped++
			return
		}
		select {
		case <-s.ch: // drop oldest
			s.dropped++
		default:
		}
		select {
		case s.ch <- e:
		default:
			s.dropped++
		}
	}
}

// Events returns the subscription's event stream.
func (s *subscription) Events() ports.Stream[ports.Event] {
	return &subStream{s: s}
}

// Close removes the subscription from the bus and closes its channel.
func (s *subscription) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	close(s.ch)
	s.mu.Unlock()

	s.bus.mu.Lock()
	delete(s.bus.subs, s)
	s.bus.mu.Unlock()
	return nil
}

// Dropped returns the number of events dropped by the overflow policy (observability).
func (s *subscription) Dropped() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dropped
}

type subStream struct{ s *subscription }

func (st *subStream) Next(ctx context.Context) (ports.Event, error) {
	select {
	case <-ctx.Done():
		return ports.Event{}, ctx.Err()
	case e, ok := <-st.s.ch:
		if !ok {
			return ports.Event{}, ports.ErrEndOfStream
		}
		return e, nil
	}
}

func (st *subStream) Close() error { return st.s.Close() }
