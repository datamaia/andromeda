package builtin

import (
	"context"
	"encoding/json"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/git"
	"github.com/datamaia/andromeda/internal/ports"
)

// GitExec runs one structured Git operation through the Git Engine (GitPort, ADR-025) — never a
// raw shell string. Read operations require `read`; mutating operations additionally require
// `git_mutation`, requested per operation. Destructive Git actions are gated by Volume 11
// confirmations, not by this tool. Phase: MVP.
type GitExec struct {
	Engine *git.Engine
}

// NewGitExec builds the git.exec tool over a Git Engine.
func NewGitExec(e *git.Engine) GitExec { return GitExec{Engine: e} }

// gitMutating is the set of operations that change repository state.
var gitMutating = map[string]bool{
	"stage": true, "unstage": true, "commit": true, "branch_create": true, "branch_switch": true,
}

// Describe returns the git_exec tool descriptor.
func (GitExec) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "git_exec", Namespace: "git", Version: "1",
		Description: "Run a structured Git operation through the Git Engine",
		InputSchema: []byte(`{"type":"object","required":["operation"],"properties":{` +
			`"operation":{"type":"string","enum":["status","log","branch_list","stage","unstage","commit","branch_create","branch_switch"]},` +
			`"args":{"type":"object"},"repo":{"type":"string"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"operation":{"type":"string"},"result":{},"warnings":{"type":"array"}}}`),
		Permissions:  []core.Permission{core.PermRead, core.PermGitMutation}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

type gitInput struct {
	Operation string          `json:"operation"`
	Args      json.RawMessage `json:"args"`
	Repo      string          `json:"repo"`
}

// Validate requires a non-empty operation.
func (GitExec) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in gitInput
	if err := json.Unmarshal(input, &in); err != nil || in.Operation == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"operation is required"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

// Resources requests read access to the repository, plus git_mutation for mutating operations.
func (GitExec) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	var in gitInput
	_ = json.Unmarshal(input, &in)
	repo := in.Repo
	if repo == "" {
		repo = "."
	}
	qs := []ports.PermissionQuery{{Permission: core.PermRead, Scope: core.ScopeRepository, Subject: repo}}
	if gitMutating[in.Operation] {
		qs = append(qs, ports.PermissionQuery{Permission: core.PermGitMutation, Scope: core.ScopeRepository, Subject: repo})
	}
	return qs, nil
}

// Execute dispatches the operation to the Git Engine and returns its result.
func (t GitExec) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in gitInput
	_ = json.Unmarshal(req.Input, &in)
	repo := ports.RepoRef{Root: in.Repo}

	result, err := t.dispatch(ctx, repo, in)
	if err != nil {
		return errEvent(err.Error()), nil
	}
	out, _ := json.Marshal(map[string]any{"operation": in.Operation, "result": result, "warnings": []string{}})
	return okEvent(string(out)), nil
}

func (t GitExec) dispatch(ctx context.Context, repo ports.RepoRef, in gitInput) (any, error) {
	switch in.Operation {
	case "status":
		return t.Engine.Status(ctx, repo)
	case "log":
		var a struct {
			Rev string `json:"rev"`
			Max int    `json:"max"`
		}
		_ = json.Unmarshal(in.Args, &a)
		st, err := t.Engine.Log(ctx, repo, ports.LogSpec{Rev: a.Rev, Max: a.Max})
		if err != nil {
			return nil, err
		}
		return collect(ctx, st)
	case "branch_list":
		return t.Engine.ListBranches(ctx, repo)
	case "stage":
		var a struct {
			Paths []string `json:"paths"`
		}
		_ = json.Unmarshal(in.Args, &a)
		return map[string]any{"staged": a.Paths}, t.Engine.Stage(ctx, repo, a.Paths)
	case "unstage":
		var a struct {
			Paths []string `json:"paths"`
		}
		_ = json.Unmarshal(in.Args, &a)
		return map[string]any{"unstaged": a.Paths}, t.Engine.Unstage(ctx, repo, a.Paths)
	case "commit":
		var a struct {
			Message string `json:"message"`
			Author  string `json:"author"`
			Signoff bool   `json:"signoff"`
		}
		_ = json.Unmarshal(in.Args, &a)
		id, err := t.Engine.Commit(ctx, repo, ports.CommitSpec{Message: a.Message, Author: a.Author, Signoff: a.Signoff})
		if err != nil {
			return nil, err
		}
		return map[string]any{"commit": id}, nil
	case "branch_create":
		var a struct {
			Name string `json:"name"`
			From string `json:"from"`
		}
		_ = json.Unmarshal(in.Args, &a)
		return map[string]any{"created": a.Name}, t.Engine.CreateBranch(ctx, repo, ports.BranchSpec{Name: a.Name, From: a.From})
	case "branch_switch":
		var a struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(in.Args, &a)
		return map[string]any{"switched": a.Name}, t.Engine.SwitchBranch(ctx, repo, a.Name)
	default:
		return nil, &ports.PortError{Code: "E-TOOL-020", Category: "tool", Severity: "error", Message: "unsupported git operation: " + in.Operation}
	}
}

// Cancel is a no-op; Git operations complete synchronously within Execute.
func (GitExec) Cancel(context.Context, core.ULID) error { return nil }

// collect drains a stream into a slice.
func collect[T any](ctx context.Context, st ports.Stream[T]) ([]T, error) {
	var out []T
	for {
		v, err := st.Next(ctx)
		if err == ports.ErrEndOfStream {
			return out, nil
		}
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}
}
