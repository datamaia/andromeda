package workflow

import "context"

// sddStages are the fourteen specification-driven-development stages (Volume 4). Gate stages
// (a subset) require human approval before proceeding.
var sddStages = []struct {
	name string
	gate bool
}{
	{"intake", false},
	{"requirements", true},
	{"research", false},
	{"planning", true},
	{"architecture", true},
	{"task-decomposition", false},
	{"implementation", false},
	{"validation", false},
	{"testing", false},
	{"review", true},
	{"security-review", true},
	{"documentation", false},
	{"completion", false},
	{"release-preparation", true},
}

// StageAction runs the work of one SDD stage. The Workflow Engine supplies the concrete action;
// in production each stage drives an agent run, in tests a stub records progress.
type StageAction func(ctx context.Context, stage string, rs *RunState) (StageResult, error)

// SDDDefinition builds the 14-stage SDD workflow using the given per-stage action.
func SDDDefinition(action StageAction) Definition {
	def := Definition{Name: "sdd"}
	for _, s := range sddStages {
		name := s.name
		def.Stages = append(def.Stages, Stage{
			Name: name,
			Gate: s.gate,
			Run: func(ctx context.Context, rs *RunState) (StageResult, error) {
				if action == nil {
					return StageResult{Summary: name + " (no-op)"}, nil
				}
				return action(ctx, name, rs)
			},
		})
	}
	return def
}

// SDDStageNames returns the ordered stage names.
func SDDStageNames() []string {
	names := make([]string, len(sddStages))
	for i, s := range sddStages {
		names[i] = s.name
	}
	return names
}
