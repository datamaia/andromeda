package app

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/datamaia/andromeda/internal/agent"
	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/git"
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
	AllowNetwork  bool // grant network access (http_request)
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
	defer func() { _ = db.Close() }()

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
	// sqlite_query is a read tool by default; mutating statements are gated per-statement by the
	// write grant below (and refused entirely unless the caller passes read_only=false).
	toolNames := []string{"fs_read", "fs_search", "fs_diff", "sqlite_query"}
	tools := []ports.ToolPort{builtin.FSRead{}, builtin.FSSearch{}, builtin.FSDiff{}, builtin.SQLiteQuery{}}
	if opts.AllowWrite {
		tools = append(tools, builtin.FSWrite{}, builtin.FSReplace{}, builtin.FSPatch{})
		toolNames = append(toolNames, "fs_write", "fs_replace", "fs_patch")
		// The Git built-in operates on the workspace repository. Read is implied by the
		// workspace read grant above; mutating operations request git_mutation, granted here so
		// a write-enabled run can stage/commit (destructive actions remain the caller's to gate).
		_, _ = pm.GrantPermission(ctx, permission.Grant{
			Permission: core.PermRead, Scope: core.ScopeRepository, Selector: "*", Effect: permission.EffectAllow,
		})
		_, _ = pm.GrantPermission(ctx, permission.Grant{
			Permission: core.PermGitMutation, Scope: core.ScopeRepository, Selector: "*", Effect: permission.EffectAllow,
		})
		tools = append(tools, builtin.NewGitExec(git.New("")))
		toolNames = append(toolNames, "git_exec")
	}
	if opts.AllowExec {
		_, _ = pm.GrantPermission(ctx, permission.Grant{
			Permission: core.PermExecute, Scope: core.ScopeCommand, Selector: "*", Effect: permission.EffectAllow,
		})
		_, _ = pm.GrantPermission(ctx, permission.Grant{
			Permission: core.PermProcessSpawn, Scope: core.ScopeHost, Selector: "*", Effect: permission.EffectAllow,
		})
		// One engine shared by both tools so process_control supervises what terminal_run starts.
		termEngine := terminal.New()
		tools = append(tools, builtin.NewTerminalRun(termEngine), builtin.NewProcessControl(termEngine))
		toolNames = append(toolNames, "terminal_run", "process_control")
	}
	if opts.AllowNetwork {
		_, _ = pm.GrantPermission(ctx, permission.Grant{
			Permission: core.PermNetwork, Scope: core.ScopeDomain, Selector: "*", Effect: permission.EffectAllow,
		})
		tools = append(tools, builtin.NewHTTPRequest(nil, nil))
		toolNames = append(toolNames, "http_request")
	}
	for _, tl := range tools {
		if err := rt.Register(ctx, tl); err != nil {
			return agent.RunResult{}, err
		}
	}

	sessions := storage.NewSessionStore(db)
	sessionID := core.NewULID()
	_ = sessions.SaveSession(ctx, ports.SessionSnapshot{ID: sessionID, State: "active"})

	// Precedence: an explicit option/flag wins; otherwise the resolved config's
	// agent.loop.max_iterations applies; the Agent Engine falls back to its own default at 0.
	maxIter := opts.MaxIterations
	if maxIter == 0 {
		if cfg, err := LoadedConfig(ctx, root); err == nil {
			if v, ok := configInt(cfg.Values["agent.loop.max_iterations"]); ok {
				maxIter = v
			}
		}
	}

	eng := agent.New(opts.Provider, rt, sessions, nil)
	return eng.Run(ctx, agent.RunInput{
		SessionID:     sessionID,
		Goal:          opts.Goal,
		System:        opts.System,
		Model:         opts.Model,
		ToolNames:     toolNames,
		MaxIterations: maxIter,
	})
}

// configInt coerces a resolved config value (int64 from TOML, or a numeric string from an env
// override) to an int.
func configInt(v any) (int, bool) {
	switch n := v.(type) {
	case int64:
		return int(n), true
	case int:
		return n, true
	case float64:
		return int(n), true
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
			return i, true
		}
	}
	return 0, false
}
