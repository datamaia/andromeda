package ports

import "context"

// SchedulerPort is supervised concurrency — the ADR-023 model as a contract. The full
// behavioral contract is Volume 3 chapter 08; pool sizes and shed policies are Volume 12's.
// Errors: E-ARCH. SchedTaskID identifies scheduler work items, distinct from the domain Task.
type SchedulerPort interface {
	Submit(ctx context.Context, spec TaskSpec) (TaskHandle, error)
	NewGroup(ctx context.Context, spec GroupSpec) (TaskGroup, error)
	Cancel(ctx context.Context, id SchedTaskID, reason CancelReason) error
	Stats(ctx context.Context) (SchedulerStats, error)
}

// TaskGroup is a structured concurrency group bound to a parent context: first error cancels
// the group; Wait joins all members (errgroup semantics).
type TaskGroup interface {
	Go(ctx context.Context, spec TaskSpec) (TaskHandle, error)
	Wait(ctx context.Context) error
	Cancel(reason CancelReason)
}

// TaskHandle awaits one supervised unit of work.
type TaskHandle interface {
	Await(ctx context.Context) (TaskOutcome, error)
	ID() SchedTaskID
}

// SchedTaskID identifies a scheduler work item.
type SchedTaskID = string

// TaskSpec describes a unit of work to schedule onto a named pool.
type TaskSpec struct {
	Pool string // "interactive" | "tools" | "background" | "io"
	Name string
	Run  func(ctx context.Context) (TaskOutcome, error)
}

// GroupSpec parameterizes a task group.
type GroupSpec struct {
	Pool string
	Name string
}

// CancelReason records why work was cancelled.
type CancelReason string

// CancelUserInterrupt, CancelTimeout, CancelBudget, and CancelShutdown are the cancellation reasons.
const (
	CancelUserInterrupt CancelReason = "user_interrupt"
	CancelTimeout       CancelReason = "timeout"
	CancelBudget        CancelReason = "budget"
	CancelShutdown      CancelReason = "shutdown"
)

// TaskOutcome is the result of a supervised task.
type TaskOutcome struct {
	OK      bool
	Message string
}

// SchedulerStats is the observability surface for NFR-ARCH-004 and Volume 12 saturation.
type SchedulerStats struct {
	Pools map[string]PoolStats
}

// PoolStats is per-pool occupancy and queue state.
type PoolStats struct {
	Active     int
	QueueDepth int
	Completed  int64
	Rejected   int64
}
