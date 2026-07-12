package app

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/workflow"
)

func TestRunSDDDrivesAgentPerStage(t *testing.T) {
	ctx := context.Background()
	// Each agent turn finishes immediately (no tool calls), so every stage completes in one turn.
	prov := &scriptedProvider{responses: []ports.ChatResponse{
		assistantMsg("stage artifact produced"),
	}}
	var stages []string
	rs, err := RunSDD(ctx, SDDOptions{
		WorkspaceRoot: t.TempDir(),
		Objective:     "add a health-check endpoint",
		Provider:      prov,
		Model:         "m",
		AutoApprove:   true,
		OnStage:       func(stage, _ string) { stages = append(stages, stage) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if rs.State != workflow.StateCompleted {
		t.Fatalf("state = %q", rs.State)
	}
	if len(stages) != 14 {
		t.Fatalf("ran %d stages, want 14", len(stages))
	}
	if len(rs.History) != 14 || rs.Artifacts["intake"] != "stage artifact produced" {
		t.Fatalf("run state = %+v", rs)
	}
}

func TestRunSDDHaltsAtGateWithoutAutoApprove(t *testing.T) {
	ctx := context.Background()
	prov := &scriptedProvider{responses: []ports.ChatResponse{assistantMsg("done")}}
	rs, err := RunSDD(ctx, SDDOptions{
		WorkspaceRoot: t.TempDir(),
		Objective:     "x",
		Provider:      prov,
		Model:         "m",
		AutoApprove:   false,
	})
	if err != nil {
		t.Fatal(err)
	}
	// The second stage (requirements) is a gate; without approval the run halts there.
	if rs.State != workflow.StateAwaitingApproval {
		t.Fatalf("state = %q, want awaiting_approval", rs.State)
	}
}
