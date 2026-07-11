// Package scheduler is layer L2 infrastructure: the Task Scheduler implementing
// ports.SchedulerPort (Volume 3 chapter 08, ADR-023). It supervises concurrent work on named,
// bounded pools with panic capture, cancellation wiring, and structured groups (first-error
// propagation, joined shutdown — errgroup semantics). SchedTaskID identifies scheduler work
// items, distinct from the domain Task entity.
package scheduler
