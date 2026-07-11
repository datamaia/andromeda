package scheduler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// DefaultPools are the pool sizes used when none are configured (Volume 12 tunes these).
var DefaultPools = map[string]int{
	"interactive": 8,
	"tools":       8,
	"background":  4,
	"io":          16,
}

// Scheduler implements ports.SchedulerPort.
type Scheduler struct {
	mu     sync.Mutex
	pools  map[string]*pool
	tasks  map[core.ULID]*task
	closed bool
}

type pool struct {
	sem       chan struct{}
	active    int64
	queued    int64
	completed int64
	rejected  int64
}

type task struct {
	id     core.ULID
	cancel context.CancelFunc
	done   chan struct{}
	out    ports.TaskOutcome
	err    error
}

// New returns a Scheduler with the given pool sizes (nil uses DefaultPools).
func New(sizes map[string]int) *Scheduler {
	if sizes == nil {
		sizes = DefaultPools
	}
	pools := map[string]*pool{}
	for name, n := range sizes {
		if n < 1 {
			n = 1
		}
		pools[name] = &pool{sem: make(chan struct{}, n)}
	}
	return &Scheduler{pools: pools, tasks: map[core.ULID]*task{}}
}

var _ ports.SchedulerPort = (*Scheduler)(nil)

func (s *Scheduler) poolFor(name string) *pool {
	if name == "" {
		name = "background"
	}
	if p, ok := s.pools[name]; ok {
		return p
	}
	return s.pools["background"]
}

// Submit schedules one supervised unit of work onto a named pool.
func (s *Scheduler) Submit(ctx context.Context, spec ports.TaskSpec) (ports.TaskHandle, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, &ports.PortError{Code: "E-ARCH-006", Category: "architecture", Message: "scheduler shut down"}
	}
	p := s.poolFor(spec.Pool)
	if p == nil {
		s.mu.Unlock()
		return nil, &ports.PortError{Code: "E-ARCH-007", Category: "architecture", Message: "unknown pool: " + spec.Pool}
	}
	taskCtx, cancel := context.WithCancel(ctx)
	tk := &task{id: core.NewULID(), cancel: cancel, done: make(chan struct{})}
	s.tasks[tk.id] = tk
	s.mu.Unlock()

	atomic.AddInt64(&p.queued, 1)
	go func() {
		// Acquire a pool slot (blocks under load; cancellation aborts before running).
		select {
		case p.sem <- struct{}{}:
		case <-taskCtx.Done():
			atomic.AddInt64(&p.queued, -1)
			tk.finish(ports.TaskOutcome{OK: false, Message: "cancelled before start"}, taskCtx.Err())
			s.remove(tk.id)
			return
		}
		atomic.AddInt64(&p.queued, -1)
		atomic.AddInt64(&p.active, 1)
		defer func() {
			atomic.AddInt64(&p.active, -1)
			atomic.AddInt64(&p.completed, 1)
			<-p.sem
			s.remove(tk.id)
		}()
		out, err := runWithRecover(taskCtx, spec.Run)
		tk.finish(out, err)
	}()

	return &handle{task: tk}, nil
}

func runWithRecover(ctx context.Context, fn func(context.Context) (ports.TaskOutcome, error)) (out ports.TaskOutcome, err error) {
	defer func() {
		if r := recover(); r != nil {
			out = ports.TaskOutcome{OK: false, Message: fmt.Sprintf("panic: %v", r)}
			err = fmt.Errorf("task panicked: %v", r)
		}
	}()
	if fn == nil {
		return ports.TaskOutcome{OK: true}, nil
	}
	return fn(ctx)
}

func (tk *task) finish(out ports.TaskOutcome, err error) {
	tk.out, tk.err = out, err
	close(tk.done)
}

func (s *Scheduler) remove(id core.ULID) {
	s.mu.Lock()
	delete(s.tasks, id)
	s.mu.Unlock()
}

// NewGroup creates a structured group bound to a parent context.
func (s *Scheduler) NewGroup(ctx context.Context, spec ports.GroupSpec) (ports.TaskGroup, error) {
	gctx, cancel := context.WithCancel(ctx)
	return &group{sched: s, pool: spec.Pool, ctx: gctx, cancel: cancel}, nil
}

// Cancel cancels a task by ID with a recorded reason.
func (s *Scheduler) Cancel(ctx context.Context, id ports.SchedTaskID, _ ports.CancelReason) error {
	s.mu.Lock()
	tk, ok := s.tasks[id]
	s.mu.Unlock()
	if !ok {
		return nil // already finished or unknown; cancellation is best-effort
	}
	tk.cancel()
	return nil
}

// Stats returns per-pool occupancy.
func (s *Scheduler) Stats(ctx context.Context) (ports.SchedulerStats, error) {
	st := ports.SchedulerStats{Pools: map[string]ports.PoolStats{}}
	for name, p := range s.pools {
		st.Pools[name] = ports.PoolStats{
			Active:     int(atomic.LoadInt64(&p.active)),
			QueueDepth: int(atomic.LoadInt64(&p.queued)),
			Completed:  atomic.LoadInt64(&p.completed),
			Rejected:   atomic.LoadInt64(&p.rejected),
		}
	}
	return st, nil
}

// Shutdown marks the scheduler closed; in-flight tasks continue but new submits are rejected.
func (s *Scheduler) Shutdown() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
}

// handle implements ports.TaskHandle.
type handle struct{ task *task }

func (h *handle) Await(ctx context.Context) (ports.TaskOutcome, error) {
	select {
	case <-ctx.Done():
		return ports.TaskOutcome{}, ctx.Err()
	case <-h.task.done:
		return h.task.out, h.task.err
	}
}

func (h *handle) ID() ports.SchedTaskID { return h.task.id }

// group implements ports.TaskGroup with first-error propagation.
type group struct {
	sched   *Scheduler
	pool    string
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex
	handles []ports.TaskHandle
}

func (g *group) Go(ctx context.Context, spec ports.TaskSpec) (ports.TaskHandle, error) {
	if spec.Pool == "" {
		spec.Pool = g.pool
	}
	h, err := g.sched.Submit(g.ctx, spec)
	if err != nil {
		return nil, err
	}
	g.mu.Lock()
	g.handles = append(g.handles, h)
	g.mu.Unlock()
	return h, nil
}

func (g *group) Wait(ctx context.Context) error {
	g.mu.Lock()
	handles := append([]ports.TaskHandle(nil), g.handles...)
	g.mu.Unlock()
	var firstErr error
	for _, h := range handles {
		_, err := h.Await(ctx)
		if err != nil && firstErr == nil {
			firstErr = err
			g.cancel() // first error cancels the group
		}
	}
	return firstErr
}

func (g *group) Cancel(_ ports.CancelReason) { g.cancel() }
