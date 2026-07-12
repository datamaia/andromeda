package workflow

import (
	"context"
	"errors"
	"testing"
)

func linearDef(names ...string) Definition {
	def := Definition{Name: "test"}
	for _, n := range names {
		nm := n
		def.Stages = append(def.Stages, Stage{Name: nm, Run: func(_ context.Context, _ *RunState) (StageResult, error) {
			return StageResult{Summary: nm + " done", Artifacts: map[string]string{nm: "ok"}}, nil
		}})
	}
	return def
}

func TestExecuteRunsAllStages(t *testing.T) {
	ctx := context.Background()
	e := New()
	rs, err := e.Execute(ctx, linearDef("a", "b", "c"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if rs.State != StateCompleted {
		t.Fatalf("state = %q", rs.State)
	}
	if len(rs.History) != 3 || rs.Artifacts["c"] != "ok" {
		t.Fatalf("run state = %+v", rs)
	}
}

func TestStageFailureFailsRun(t *testing.T) {
	ctx := context.Background()
	def := Definition{Name: "t", Stages: []Stage{
		{Name: "ok", Run: func(context.Context, *RunState) (StageResult, error) { return StageResult{}, nil }},
		{Name: "boom", Run: func(context.Context, *RunState) (StageResult, error) { return StageResult{}, errors.New("kaboom") }},
		{Name: "never", Run: func(context.Context, *RunState) (StageResult, error) {
			t.Fatal("should not run")
			return StageResult{}, nil
		}},
	}}
	rs, err := New().Execute(ctx, def, nil, 0)
	if err == nil || rs.State != StateFailed {
		t.Fatalf("expected failure, got state=%q err=%v", rs.State, err)
	}
}

func TestGateHaltsWithoutApprover(t *testing.T) {
	ctx := context.Background()
	def := Definition{Name: "t", Stages: []Stage{
		{Name: "a", Run: func(context.Context, *RunState) (StageResult, error) { return StageResult{}, nil }},
		{Name: "gate", Gate: true, Run: func(context.Context, *RunState) (StageResult, error) {
			t.Fatal("gate stage ran without approval")
			return StageResult{}, nil
		}},
	}}
	rs, err := New().Execute(ctx, def, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if rs.State != StateAwaitingApproval || rs.StageIdx != 1 {
		t.Fatalf("expected awaiting_approval at stage 1, got %+v", rs)
	}
}

type autoApprover struct{ calls int }

func (a *autoApprover) ApproveGate(context.Context, string, string) (bool, error) {
	a.calls++
	return true, nil
}

func TestGateApprovedProceeds(t *testing.T) {
	ctx := context.Background()
	ap := &autoApprover{}
	def := Definition{Name: "t", Stages: []Stage{
		{Name: "gate", Gate: true, Run: func(context.Context, *RunState) (StageResult, error) { return StageResult{Summary: "gated ran"}, nil }},
	}}
	rs, err := New(WithApprover(ap)).Execute(ctx, def, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if rs.State != StateCompleted || ap.calls != 1 {
		t.Fatalf("gate approval flow wrong: state=%q calls=%d", rs.State, ap.calls)
	}
}

func TestResumeFromStage(t *testing.T) {
	ctx := context.Background()
	def := linearDef("a", "b", "c")
	rs := &RunState{RunID: "r1", Workflow: def.Name, Artifacts: map[string]string{}}
	// Resume from stage 2 (index) — only "c" should run.
	out, err := New().Execute(ctx, def, rs, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.History) != 1 || out.History[0].Stage != "c" {
		t.Fatalf("resume history = %+v", out.History)
	}
}

func TestSDDDefinitionHasFourteenStages(t *testing.T) {
	names := SDDStageNames()
	if len(names) != 14 || names[0] != "intake" || names[13] != "release-preparation" {
		t.Fatalf("SDD stages = %v", names)
	}
	def := SDDDefinition(func(_ context.Context, stage string, _ *RunState) (StageResult, error) {
		return StageResult{Summary: stage}, nil
	})
	rs, err := New(WithAutoApproveGates()).Execute(context.Background(), def, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if rs.State != StateCompleted || len(rs.History) != 14 {
		t.Fatalf("SDD run = %+v", rs)
	}
}
