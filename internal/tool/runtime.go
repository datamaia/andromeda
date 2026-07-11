package tool

import (
	"context"
	"fmt"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

// ResourceScoped is an optional tool interface: a tool that can derive the concrete permission
// queries its input implies (e.g. the specific paths it will read/write) implements it so the
// Runtime evaluates path-level permissions rather than coarse tool-level ones.
type ResourceScoped interface {
	Resources(input ports.JSON) ([]ports.PermissionQuery, error)
}

// Runtime registers tools and mediates their invocation under validation and permissions.
type Runtime struct {
	perms ports.PermissionPort
	tools map[string]ports.ToolPort
}

// NewRuntime returns a Tool Runtime that decides permissions through perms.
func NewRuntime(perms ports.PermissionPort) *Runtime {
	return &Runtime{perms: perms, tools: map[string]ports.ToolPort{}}
}

// Register adds a tool by its declared name.
func (r *Runtime) Register(ctx context.Context, t ports.ToolPort) error {
	d, err := t.Describe(ctx)
	if err != nil {
		return err
	}
	if d.Name == "" {
		return fmt.Errorf("tool has no name")
	}
	r.tools[d.Name] = t
	return nil
}

// Names returns the registered tool names.
func (r *Runtime) Names() []string {
	out := make([]string, 0, len(r.tools))
	for n := range r.tools {
		out = append(out, n)
	}
	return out
}

// Describe returns a registered tool's descriptor.
func (r *Runtime) Describe(ctx context.Context, name string) (ports.ToolDescriptor, error) {
	t, ok := r.tools[name]
	if !ok {
		return ports.ToolDescriptor{}, toolErr("E-TOOL-001", "unknown tool: "+name)
	}
	return t.Describe(ctx)
}

// Invoke validates input, evaluates permissions, and drives the tool, returning its ToolEvent
// stream. A validation failure returns a Go error (the invocation could not start). A
// permission denial is delivered as data: a single terminal error event.
func (r *Runtime) Invoke(ctx context.Context, name string, subjectCtx ports.PermissionQuery, input ports.JSON) (ports.Stream[ports.ToolEvent], error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, toolErr("E-TOOL-001", "unknown tool: "+name)
	}
	desc, err := t.Describe(ctx)
	if err != nil {
		return nil, err
	}
	vr, err := t.Validate(ctx, input)
	if err != nil {
		return nil, err
	}
	if !vr.Valid {
		return nil, toolErr("E-TOOL-002", "input validation failed for "+name)
	}

	// Evaluate permissions. Prefer path-level queries when the tool provides them.
	queries := r.permissionQueries(t, desc, subjectCtx, input)
	for _, q := range queries {
		d, err := r.perms.Check(ctx, q)
		if err != nil || d.Outcome != core.OutcomeAllow {
			return streams.Slice([]ports.ToolEvent{{
				Kind: "terminal", Terminal: true, Outcome: "error",
				Text: fmt.Sprintf("permission denied: %s on %s", q.Permission, q.Subject),
			}}), nil
		}
	}

	req := ports.ToolExecuteRequest{
		InvocationID: core.NewULID(),
		Input:        input,
		Granted:      desc.Permissions,
	}
	return t.Execute(ctx, req)
}

func (r *Runtime) permissionQueries(t ports.ToolPort, desc ports.ToolDescriptor, subj ports.PermissionQuery, input ports.JSON) []ports.PermissionQuery {
	if rs, ok := t.(ResourceScoped); ok {
		if qs, err := rs.Resources(input); err == nil && len(qs) > 0 {
			// carry the subject's session/workspace context onto each derived query
			for i := range qs {
				qs[i].SessionID, qs[i].WorkspaceID = subj.SessionID, subj.WorkspaceID
			}
			return qs
		}
	}
	qs := make([]ports.PermissionQuery, 0, len(desc.Permissions))
	for _, p := range desc.Permissions {
		qs = append(qs, ports.PermissionQuery{
			Permission: p, Scope: core.ScopeTool, Subject: desc.Name,
			SessionID: subj.SessionID, WorkspaceID: subj.WorkspaceID,
		})
	}
	return qs
}

func toolErr(code, msg string) error {
	return &ports.PortError{Code: code, Category: "tool", Severity: "error", Message: msg}
}
