package app

import (
	"context"
	"fmt"

	"github.com/datamaia/andromeda/internal/agent"
	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/permission"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/datamaia/andromeda/internal/tool"
	"github.com/datamaia/andromeda/internal/tool/builtin"
	"github.com/datamaia/andromeda/internal/workflow"
)

// SDDOptions parameterizes an agent-driven SDD run.
type SDDOptions struct {
	WorkspaceRoot string
	Objective     string
	Provider      ports.ProviderPort
	Model         string
	AutoApprove   bool
	OnStage       func(stage, result string) // optional progress callback
}

// RunSDD executes the 14-stage specification-driven-development workflow with each stage driven
// by the Agent Engine: the stage's action runs an agent goal scoped to that stage and the
// overall objective. Gate stages require approval (auto-approved when AutoApprove is set).
func RunSDD(ctx context.Context, opts SDDOptions) (*workflow.RunState, error) {
	db, err := storage.OpenWorkspaceDB(ctx, opts.WorkspaceRoot)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	pm := permission.NewManager(permission.NewStore(db), permission.WithActor("sdd"))
	_, _ = pm.GrantPermission(ctx, permission.Grant{Permission: core.PermRead, Scope: core.ScopePath, Selector: "*", Effect: permission.EffectAllow})
	rt := tool.NewRuntime(pm)
	_ = rt.Register(ctx, builtin.FSRead{})
	_ = rt.Register(ctx, builtin.FSSearch{})

	eng := agent.New(opts.Provider, rt, nil, nil)

	action := func(ctx context.Context, stage string, _ *workflow.RunState) (workflow.StageResult, error) {
		goal := fmt.Sprintf("Perform the %q stage of specification-driven development for this objective: %s", stage, opts.Objective)
		res, err := eng.Run(ctx, agent.RunInput{
			Goal:      goal,
			System:    "You are Andromeda executing one SDD stage. Be concise and produce the stage's artifact.",
			Model:     opts.Model,
			ToolNames: []string{"fs_read", "fs_search"},
		})
		if err != nil {
			return workflow.StageResult{}, err
		}
		if opts.OnStage != nil {
			opts.OnStage(stage, res.FinalText)
		}
		return workflow.StageResult{Summary: res.FinalText, Artifacts: map[string]string{stage: res.FinalText}}, nil
	}

	wopts := []workflow.Option{}
	if opts.AutoApprove {
		wopts = append(wopts, workflow.WithAutoApproveGates())
	}
	def := workflow.SDDDefinition(action)
	return workflow.New(wopts...).Execute(ctx, def, nil, 0)
}
