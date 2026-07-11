package app

import (
	"context"
	"path/filepath"

	"github.com/datamaia/andromeda/internal/agent"
	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/permission"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/datamaia/andromeda/internal/terminal"
	"github.com/datamaia/andromeda/internal/tool"
	"github.com/datamaia/andromeda/internal/tool/builtin"
)

// RunAgentOptions parameterizes a composed agent run.
type RunAgentOptions struct {
	WorkspaceRoot string
	Goal          string
	System        string
	Model         string
	Provider      ports.ProviderPort
	AllowWrite    bool // grant write within the workspace (safe-by-default is read-only)
	AllowExec     bool // grant command execution (terminal_run)
	MaxIterations int
}

// RunAgent composes the workspace, permission manager (with safe-by-default grants scoped to
// the workspace), the Tool Runtime with the built-in filesystem tools, and the Agent Engine,
// then runs the goal to completion. It is the composition behind `andromeda run`.
func RunAgent(ctx context.Context, opts RunAgentOptions) (agent.RunResult, error) {
	root, err := filepath.Abs(opts.WorkspaceRoot)
	if err != nil {
		return agent.RunResult{}, err
	}
	db, err := storage.OpenWorkspaceDB(ctx, root)
	if err != nil {
		return agent.RunResult{}, err
	}
	defer db.Close()

	pm := permission.NewManager(permission.NewStore(db), permission.WithActor("cli"))
	// Safe by default: read is granted within the workspace subtree; write only when asked.
	if _, err := pm.GrantPermission(ctx, permission.Grant{
		Permission: core.PermRead, Scope: core.ScopePath, Selector: root + "/**", Effect: permission.EffectAllow,
	}); err != nil {
		return agent.RunResult{}, err
	}
	// fs_search defaults to root "." — grant read on the relative root too.
	_, _ = pm.GrantPermission(ctx, permission.Grant{
		Permission: core.PermRead, Scope: core.ScopePath, Selector: ".", Effect: permission.EffectAllow,
	})
	if opts.AllowWrite {
		_, _ = pm.GrantPermission(ctx, permission.Grant{
			Permission: core.PermWrite, Scope: core.ScopePath, Selector: root + "/**", Effect: permission.EffectAllow,
		})
	}

	rt := tool.NewRuntime(pm)
	toolNames := []string{"fs_read", "fs_search"}
	tools := []ports.ToolPort{builtin.FSRead{}, builtin.FSSearch{}}
	if opts.AllowWrite {
		tools = append(tools, builtin.FSWrite{})
		toolNames = append(toolNames, "fs_write")
	}
	if opts.AllowExec {
		_, _ = pm.GrantPermission(ctx, permission.Grant{
			Permission: core.PermExecute, Scope: core.ScopeCommand, Selector: "*", Effect: permission.EffectAllow,
		})
		tools = append(tools, builtin.NewTerminalRun(terminal.New()))
		toolNames = append(toolNames, "terminal_run")
	}
	for _, tl := range tools {
		if err := rt.Register(ctx, tl); err != nil {
			return agent.RunResult{}, err
		}
	}

	sessions := storage.NewSessionStore(db)
	sessionID := core.NewULID()
	_ = sessions.SaveSession(ctx, ports.SessionSnapshot{ID: sessionID, State: "active"})

	eng := agent.New(opts.Provider, rt, sessions, nil)
	return eng.Run(ctx, agent.RunInput{
		SessionID:     sessionID,
		Goal:          opts.Goal,
		System:        opts.System,
		Model:         opts.Model,
		ToolNames:     toolNames,
		MaxIterations: opts.MaxIterations,
	})
}
