package workflow

import (
	"context"
	"fmt"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/eventbus"
	"github.com/datamaia/andromeda/internal/ports"
)

// Frozen Workflow Run states (Volume 2, chapter 09).
const (
	StatePending          = "pending"
	StateRunning          = "running"
	StateAwaitingApproval = "awaiting_approval"
	StatePaused           = "paused"
	StateInterrupted      = "interrupted"
	StateCompleted        = "completed"
	StateFailed           = "failed"
	StateCancelled        = "cancelled"
)

// StageResult is what a stage produces.
type StageResult struct {
	Summary   string
	Artifacts map[string]string
}

// Stage is one step of a workflow. Gate stages request human approval before running.
type Stage struct {
	Name string
	Gate bool
	Run  func(ctx context.Context, rs *RunState) (StageResult, error)
}

// Definition is an ordered set of stages.
type Definition struct {
	Name   string
	Stages []Stage
}

// RunState is the evolving state of a workflow run.
type RunState struct {
	RunID     core.ULID
	Workflow  string
	State     string
	StageIdx  int
	Artifacts map[string]string
	History   []StageRecord
}

// StageRecord records the outcome of one stage.
type StageRecord struct {
	Stage   string
	Summary string
	OK      bool
}

// GateApprover renders an approval decision for a gate stage. Nil means non-interactive: gates
// are auto-approved only when AutoApproveGates is set, otherwise the run halts awaiting_approval.
type GateApprover interface {
	ApproveGate(ctx context.Context, workflow, stage string) (bool, error)
}

// Engine executes workflow definitions.
type Engine struct {
	bus              ports.EventBusPort
	approver         GateApprover
	autoApproveGates bool
}

// Option configures the Engine.
type Option func(*Engine)

// WithBus wires event publication.
func WithBus(b ports.EventBusPort) Option { return func(e *Engine) { e.bus = b } }

// WithApprover sets the gate approver.
func WithApprover(a GateApprover) Option { return func(e *Engine) { e.approver = a } }

// WithAutoApproveGates approves all gates automatically (non-interactive CI use).
func WithAutoApproveGates() Option { return func(e *Engine) { e.autoApproveGates = true } }

// New builds a Workflow Engine.
func New(opts ...Option) *Engine {
	e := &Engine{}
	for _, o := range opts {
		o(e)
	}
	return e
}

// Execute runs a definition to completion, awaiting_approval, or failure. Start begins at a
// stage index (0 for a fresh run; > 0 to resume).
func (e *Engine) Execute(ctx context.Context, def Definition, rs *RunState, start int) (*RunState, error) {
	if rs == nil {
		rs = &RunState{RunID: core.NewULID(), Workflow: def.Name, Artifacts: map[string]string{}}
	}
	rs.State = StateRunning
	e.emit("workflow.run.started", rs)

	for i := start; i < len(def.Stages); i++ {
		if err := ctx.Err(); err != nil {
			rs.State = StateInterrupted
			e.emit("workflow.run.interrupted", rs)
			return rs, err
		}
		stage := def.Stages[i]
		rs.StageIdx = i

		if stage.Gate {
			approved, err := e.gate(ctx, def.Name, stage.Name, rs)
			if err != nil {
				rs.State = StateFailed
				e.emit("workflow.run.failed", rs)
				return rs, err
			}
			if !approved {
				// No decision available and not auto-approved: halt for later resume.
				rs.State = StateAwaitingApproval
				e.emit("workflow.run.awaiting_approval", rs)
				return rs, nil
			}
			rs.State = StateRunning
		}

		res, err := stage.Run(ctx, rs)
		rec := StageRecord{Stage: stage.Name, Summary: res.Summary, OK: err == nil}
		rs.History = append(rs.History, rec)
		for k, v := range res.Artifacts {
			rs.Artifacts[k] = v
		}
		e.emit("workflow.stage.completed", rs)
		if err != nil {
			rs.State = StateFailed
			e.emit("workflow.run.failed", rs)
			return rs, fmt.Errorf("stage %q failed: %w", stage.Name, err)
		}
	}

	rs.State = StateCompleted
	e.emit("workflow.run.completed", rs)
	return rs, nil
}

// gate resolves a gate: an approver decision, or auto-approval, or "not approved" (halt).
func (e *Engine) gate(ctx context.Context, wf, stage string, rs *RunState) (bool, error) {
	rs.State = StateAwaitingApproval
	e.emit("workflow.gate.reached", rs)
	if e.approver != nil {
		return e.approver.ApproveGate(ctx, wf, stage)
	}
	if e.autoApproveGates {
		return true, nil
	}
	return false, nil
}

func (e *Engine) emit(name string, rs *RunState) {
	if e.bus == nil {
		return
	}
	_ = e.bus.Publish(context.Background(), eventbus.NewEvent(name, "workflow-engine", eventbus.WithRun(rs.RunID)))
}
