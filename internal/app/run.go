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
	Effort        string          // reasoning effort (minimal|low|medium|high); empty leaves it to the provider
	History       []ports.Message // prior conversation turns to continue (empty starts fresh)
	Provider      ports.ProviderPort
	AllowWrite    bool // grant write within the workspace (safe-by-default is read-only)
	AllowExec     bool // grant command execution (terminal_run)
	AllowNetwork  bool // grant network access (http_request)
	MaxIterations int

	// Interactive registers the full toolset but pre-grants only reads: every state-changing
	// action (write, git mutation, execute, network) resolves to "ask" and is routed to Approver,
	// which prompts the user (approve once / for session / for workspace, or deny). The user's
	// standing decisions are persisted as grants, forming the session/workspace allow- and
	// deny-lists. A nil Approver makes "ask" fail closed (deny), same as a non-interactive run.
	Interactive bool
	Approver    permission.Approver

	// Sink, when non-null, receives streamed run events (content deltas, tool calls, tool results)
	// as the run proceeds, for a live transcript. See agent.RunEvent.
	Sink func(agent.RunEvent)
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

	mgrOpts := []permission.Option{permission.WithActor("cli")}
	if opts.Interactive && opts.Approver != nil {
		mgrOpts = append(mgrOpts, permission.WithApprover(opts.Approver))
	}
	pm := permission.NewManager(permission.NewStore(db), mgrOpts...)
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

	// In interactive mode the full toolset is registered but state-changing permissions are left
	// ungranted, so each is prompted for; otherwise a capability is available only when its flag
	// pre-grants it. preGrant enables the classic non-interactive behaviour.
	registerWrite := opts.AllowWrite || opts.Interactive
	registerExec := opts.AllowExec || opts.Interactive
	registerNet := opts.AllowNetwork || opts.Interactive
	preGrant := !opts.Interactive

	if opts.AllowWrite && preGrant {
		_, _ = pm.GrantPermission(ctx, permission.Grant{
			Permission: core.PermWrite, Scope: core.ScopePath, Selector: root + "/**", Effect: permission.EffectAllow,
		})
	}

	rtOpts := []tool.RuntimeOption{}
	if opts.Interactive {
		rtOpts = append(rtOpts, tool.WithInteractive())
	}
	rt := tool.NewRuntime(pm, rtOpts...)
	// sqlite_query is a read tool by default; mutating statements are gated per-statement by the
	// write grant below (and refused entirely unless the caller passes read_only=false).
	toolNames := []string{"fs_read", "fs_search", "fs_diff", "sqlite_query"}
	tools := []ports.ToolPort{builtin.FSRead{}, builtin.FSSearch{}, builtin.FSDiff{}, builtin.SQLiteQuery{}}
	if registerWrite {
		tools = append(tools, builtin.FSWrite{}, builtin.FSReplace{}, builtin.FSPatch{})
		toolNames = append(toolNames, "fs_write", "fs_replace", "fs_patch")
		// The Git built-in operates on the workspace repository. Repository read is safe and always
		// granted; mutating operations request git_mutation — pre-granted for a write-enabled
		// non-interactive run, prompted for interactively (destructive actions stay the user's).
		_, _ = pm.GrantPermission(ctx, permission.Grant{
			Permission: core.PermRead, Scope: core.ScopeRepository, Selector: "*", Effect: permission.EffectAllow,
		})
		if opts.AllowWrite && preGrant {
			_, _ = pm.GrantPermission(ctx, permission.Grant{
				Permission: core.PermGitMutation, Scope: core.ScopeRepository, Selector: "*", Effect: permission.EffectAllow,
			})
		}
		tools = append(tools, builtin.NewGitExec(git.New("")))
		toolNames = append(toolNames, "git_exec")
	}
	if registerExec {
		if opts.AllowExec && preGrant {
			_, _ = pm.GrantPermission(ctx, permission.Grant{
				Permission: core.PermExecute, Scope: core.ScopeCommand, Selector: "*", Effect: permission.EffectAllow,
			})
			_, _ = pm.GrantPermission(ctx, permission.Grant{
				Permission: core.PermProcessSpawn, Scope: core.ScopeHost, Selector: "*", Effect: permission.EffectAllow,
			})
		}
		// One engine shared by both tools so process_control supervises what terminal_run starts.
		termEngine := terminal.New()
		tools = append(tools, builtin.NewTerminalRun(termEngine), builtin.NewProcessControl(termEngine))
		toolNames = append(toolNames, "terminal_run", "process_control")
	}
	if registerNet {
		if opts.AllowNetwork && preGrant {
			_, _ = pm.GrantPermission(ctx, permission.Grant{
				Permission: core.PermNetwork, Scope: core.ScopeDomain, Selector: "*", Effect: permission.EffectAllow,
			})
		}
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

	// AGENTS.md is active: its content is folded into the system prompt on every run so project
	// guidance steers the agent (the file is read here, in the composition layer).
	system := composeSystem(opts.System, projectInstructions(root))

	eng := agent.New(opts.Provider, rt, sessions, nil)
	return eng.Run(ctx, agent.RunInput{
		SessionID:     sessionID,
		Goal:          opts.Goal,
		System:        system,
		Model:         opts.Model,
		Effort:        opts.Effort,
		History:       opts.History,
		ToolNames:     toolNames,
		MaxIterations: maxIter,
		Sink:          opts.Sink,
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
