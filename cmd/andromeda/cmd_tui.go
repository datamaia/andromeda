package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/buildinfo"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/settingstore"
	"github.com/datamaia/andromeda/internal/tui"
	"github.com/spf13/cobra"
)

// tuiConfig holds the provider wiring for a TUI session. It is shared by the explicit
// `tui` subcommand and the bare `andromeda` invocation (FR-CLI-003) so both launch the
// same shell; the defaults match the `run` command (local Ollama, no granted capabilities).
type tuiConfig struct {
	provider, baseURL, apiKeyEnv, model string
	effort                              string // reasoning effort chosen via /effort
	allowWrite, allowExec               bool
}

func defaultTUIConfig() tuiConfig {
	return tuiConfig{provider: "ollama", model: "llama3"}
}

// tuiSession owns the live provider for a TUI session so the in-app provider menu can rebuild it
// mid-session. The agent responder and the menu's select callback both read/mutate it.
type tuiSession struct {
	ctx  context.Context // program lifetime; cancelling it tears a run and its approver down
	wd   string
	cfg  tuiConfig
	prov ports.ProviderPort

	// persistent session: the conversation so far (fed back as each run's history for cross-turn
	// memory) and the id under which it is saved to disk after every turn.
	sessionID string
	history   []ports.Message

	// graphURL is the address of the graph viewer once /graph open has started it; empty until then.
	// The server runs for the program lifetime (bound to s.ctx), so a second open just reopens it.
	graphURL string

	// extraDirs are additional working directories added with /add-dir; their files join @-mention
	// completion alongside the primary workspace (s.wd).
	extraDirs []string

	// pendingNotes are out-of-band context lines queued with /btw; they are folded into the next
	// message to the agent (and then cleared) rather than triggering a reply on their own.
	pendingNotes []string

	// autoCompact mirrors the workspace setting (.andromeda/settings.toml): when true, the history is
	// summarized before a turn once it grows past autoCompactTurns. Toggled by /autocompact.
	autoCompact bool

	// undoStack/redoStack hold git tree SHAs of working-tree checkpoints for /undo and /redo. A
	// pre-turn snapshot is pushed to undoStack before each agent turn; only touched on the UI thread.
	undoStack []string
	redoStack []string
}

func (s *tuiSession) build() error {
	apiKey := ""
	if s.cfg.apiKeyEnv != "" {
		apiKey = os.Getenv(s.cfg.apiKeyEnv)
	}
	prov, err := app.BuildProvider(app.ProviderSpec{Name: s.cfg.provider, BaseURL: s.cfg.baseURL, APIKey: apiKey})
	if err != nil {
		return err
	}
	s.prov = prov
	if st, err := settingstore.Load(s.wd); err == nil {
		s.autoCompact = st.AutoCompact
	}
	return nil
}

// planModeSystem constrains the agent to proposing a plan without touching anything.
const planModeSystem = "You are in PLAN MODE. Analyze the request and propose a concise, numbered " +
	"plan of the steps you would take. You have read-only access: do NOT create, edit, delete, or " +
	"run anything. End by asking the user to switch to agent mode (/agent) to execute."

// agentModeSystem makes the agent take action: it must call the available tools to accomplish the
// goal rather than describing what it would do. Destructive actions are gated by the permission
// prompt, so the model is told to act directly and let the user approve or deny at that point.
const agentModeSystem = "You are Andromeda, an autonomous software engineering agent working inside " +
	"the user's real workspace. You have tools to read, search, write, and edit files, run shell " +
	"commands, and use git. When the user asks you to do something, DO IT by calling the appropriate " +
	"tools — do not merely describe the steps, and do not ask for permission you don't need (a " +
	"separate approval prompt guards state-changing actions, so just call the tool and let the user " +
	"approve or deny). Only ask the user a question when a genuinely required detail is missing. " +
	"Use absolute or workspace-relative paths. After acting, briefly report what you did."

// respond handles a submitted line according to the active interaction mode: shell runs it as a
// command, plan drives the agent read-only with a planning prompt, agent runs the full loop with
// whatever capabilities the session was granted.
func (s *tuiSession) respond(goal, mode string) string {
	switch mode {
	case "shell":
		return s.runShell(goal)
	case "plan":
		return s.runAgent(goal, planModeSystem, false, false)
	default:
		return s.runAgent(goal, "", s.cfg.allowWrite, s.cfg.allowExec)
	}
}

// runAgent drives the real agent for a goal with explicit capability grants.
func (s *tuiSession) runAgent(goal, system string, allowWrite, allowExec bool) string {
	res, err := app.RunAgent(context.Background(), app.RunAgentOptions{
		WorkspaceRoot: s.wd, Goal: goal, System: system, Model: s.cfg.model, Effort: s.cfg.effort,
		Provider: s.prov, AllowWrite: allowWrite, AllowExec: allowExec,
	})
	if err != nil {
		return "error: " + err.Error()
	}
	return res.FinalText
}

// selectEffort records the reasoning effort chosen via /effort so the agent runs with it.
func (s *tuiSession) selectEffort(effort string) { s.cfg.effort = effort }

// gitBranch returns the current branch name, or "" when the workspace is not a git repo.
func gitBranch(ctx context.Context, wd string) string {
	out, err := exec.CommandContext(ctx, "git", "-C", wd, "rev-parse", "--abbrev-ref", "HEAD").Output() //nolint:gosec // fixed 'git' command; wd is the workspace path
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// contextAction supplies the /status Context tab: workspace root, branch, and uncommitted-change
// count. Best-effort — git failures simply omit those lines.
func (s *tuiSession) contextAction(ctx context.Context) []string {
	lines := []string{fmt.Sprintf("%-12s%s", "workspace", s.wd)}
	for _, d := range s.extraDirs {
		lines = append(lines, fmt.Sprintf("%-12s%s", "+dir", d))
	}
	if b := gitBranch(ctx, s.wd); b != "" {
		lines = append(lines, fmt.Sprintf("%-12s%s", "branch", b))
	}
	if out, err := exec.CommandContext(ctx, "git", "-C", s.wd, "status", "--porcelain").Output(); err == nil { //nolint:gosec // fixed 'git' command; s.wd is the workspace path
		n := 0
		for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if strings.TrimSpace(l) != "" {
				n++
			}
		}
		lines = append(lines, fmt.Sprintf("%-12s%d file(s)", "changes", n))
	}
	return lines
}

// runShell runs the line as a command in the workspace directory (shell mode is the user's own
// command, not the agent's, so it is not gated by the agent permission model).
func (s *tuiSession) runShell(line string) string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.CommandContext(context.Background(), shell, "-c", line) //nolint:gosec // shell mode runs the user's own typed command in their shell
	cmd.Dir = s.wd
	out, err := cmd.CombinedOutput()
	text := strings.TrimRight(string(out), "\n")
	if err != nil {
		if text != "" {
			return text + "\n[" + err.Error() + "]"
		}
		return "[" + err.Error() + "]"
	}
	if text == "" {
		return "[no output]"
	}
	return text
}

// selectProvider rebuilds the live provider from a catalog ID (menu selection), adopting the
// catalog's default model, and returns the model now in use. It reads the key from the catalog's
// environment variable; an unset required key surfaces as an error the menu shows.
func (s *tuiSession) selectProvider(id string) (string, error) {
	info, ok := app.LookupProvider(id)
	if !ok {
		return "", fmt.Errorf("unknown provider %q", id)
	}
	prov, err := app.BuildProvider(app.ProviderSpec{Name: id})
	if err != nil {
		return "", err
	}
	s.prov = prov
	s.cfg.provider = id
	if info.DefaultModel != "" {
		s.cfg.model = info.DefaultModel
	}
	s.rememberChoice()
	return s.cfg.model, nil
}

// selectModel records the model the user chose so the agent runs on it (not the provider default).
func (s *tuiSession) selectModel(id string) {
	s.cfg.model = id
	s.rememberChoice()
}

// rememberChoice persists the current provider/model as the global default for the next launch
// (best-effort — a returning user is dropped back into their last setup instead of re-onboarding).
func (s *tuiSession) rememberChoice() {
	_ = app.SavePrefs(app.Prefs{Provider: s.cfg.provider, Model: s.cfg.model})
}

// persistSession saves the current conversation to disk (best-effort). It no-ops on an empty history
// or when the store is unavailable, so a run never fails because persistence did.
func (s *tuiSession) persistSession(mode string) {
	if len(s.history) == 0 {
		return
	}
	_ = app.SaveSession(app.StoredSession{
		ID:        s.sessionID,
		Title:     app.SessionTitle(s.history),
		Provider:  s.cfg.provider,
		Model:     s.cfg.model,
		Mode:      mode,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Messages:  s.history,
	})
}

// historyEntries flattens stored provider messages into displayable transcript lines (user prompts
// and assistant replies; tool and system turns are omitted from the visible transcript).
func historyEntries(msgs []ports.Message) []tui.HistoryEntry {
	var out []tui.HistoryEntry
	for _, m := range msgs {
		text := messageText(m)
		if strings.TrimSpace(text) == "" {
			continue
		}
		switch m.Role {
		case "user":
			out = append(out, tui.HistoryEntry{Role: "user", Text: text})
		case "assistant":
			out = append(out, tui.HistoryEntry{Role: "agent", Text: text})
		}
	}
	return out
}

// messageText concatenates the text parts of a message.
func messageText(m ports.Message) string {
	var b strings.Builder
	for _, p := range m.Parts {
		if p.Type == "" || p.Type == "text" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

// listFiles returns workspace-relative file paths for @-mention completion. It prefers `git
// ls-files` (fast, honors .gitignore) and falls back to a bounded directory walk that skips VCS and
// dependency directories.
func (s *tuiSession) listFiles(ctx context.Context) []string {
	files := filesIn(ctx, s.wd, "")
	// Files from each /add-dir directory join the list, prefixed by the directory's base name so they
	// are distinguishable and usable as @-mentions.
	for _, dir := range s.extraDirs {
		files = append(files, filesIn(ctx, dir, filepath.Base(dir)+"/")...)
	}
	return files
}

// filesIn lists a directory's files (git ls-files when it is a repo, else a bounded walk), each
// prefixed with prefix. Directories that only add noise are skipped in the walk fallback.
func filesIn(ctx context.Context, dir, prefix string) []string {
	if out, err := exec.CommandContext(ctx, "git", "-C", dir, "ls-files").Output(); err == nil { //nolint:gosec // fixed 'git' command; dir is a workspace path
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			for i := range lines {
				lines[i] = prefix + lines[i]
			}
			return lines
		}
	}
	const maxFiles = 20000
	var files []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor", ".venv", "target", "dist", "build":
				return filepath.SkipDir
			}
			return nil
		}
		if rel, err := filepath.Rel(dir, p); err == nil {
			files = append(files, prefix+rel)
		}
		if len(files) >= maxFiles {
			return filepath.SkipAll
		}
		return nil
	})
	return files
}

// resolvePath resolves a possibly ~- or relative path against the current working directory.
func (s *tuiSession) resolvePath(p string) string {
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, p[2:])
		}
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(s.wd, p)
}

// addDirAction backs /add-dir: it validates the path is a directory and adds it to the session's
// extra working directories (de-duplicated).
func (s *tuiSession) addDirAction(_ context.Context, path string) string {
	abs := s.resolvePath(path)
	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		return "add-dir: not a directory: " + path
	}
	if abs == s.wd {
		return "add-dir: that is already the working directory"
	}
	for _, d := range s.extraDirs {
		if d == abs {
			return "add-dir: already added: " + abs
		}
	}
	s.extraDirs = append(s.extraDirs, abs)
	return "added working directory: " + abs
}

// cdAction backs /cd: it validates the path is a directory and moves the session there, so later runs
// and completions use the new root. Returns (resolvedDir, gitBranch, status); dir is empty on error.
func (s *tuiSession) cdAction(ctx context.Context, path string) (string, string, string) {
	abs := s.resolvePath(path)
	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		return "", "", "cd: not a directory: " + path
	}
	s.wd = abs
	return abs, gitBranch(ctx, abs), "working directory → " + abs
}

// providerChoices adapts the app catalog into the TUI's menu entries.
func providerChoices() []tui.ProviderChoice {
	infos := app.Providers()
	choices := make([]tui.ProviderChoice, 0, len(infos))
	for _, p := range infos {
		auth := "no key"
		if p.KeyEnv != "" {
			auth = p.KeyEnv
		}
		choices = append(choices, tui.ProviderChoice{ID: p.ID, Display: p.Display, Auth: auth, Note: p.Note})
	}
	return choices
}

// launchTUIResume builds the session and hands control to the TUI shell with the provider menu,
// agent runner, interactive sign-in, and API-key entry wired. When onboard is true the session
// opens in first-run mode: a provider (with sign-in/key) and a model must be chosen before chatting.
// When resumeID is set, the saved conversation is loaded so the agent remembers it and the
// transcript view is re-seeded; onboarding is skipped for a resumed session.
func launchTUIResume(ctx context.Context, cfg tuiConfig, onboard bool, resumeID string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	sess := &tuiSession{ctx: ctx, wd: wd, cfg: cfg, sessionID: app.NewSessionID()}

	var restored []tui.HistoryEntry
	if resumeID != "" {
		st, err := app.LoadSession(resumeID)
		if err != nil {
			return fmt.Errorf("resume %s: %w", resumeID, err)
		}
		sess.sessionID = st.ID
		sess.history = st.Messages
		if st.Provider != "" {
			sess.cfg.provider = st.Provider
		}
		if st.Model != "" {
			sess.cfg.model = st.Model
		}
		restored = historyEntries(st.Messages)
		onboard = false // a resumed session already has a provider and model
	}

	// Remember the last provider/model the user actually used: on a bare onboarding launch, default
	// to it and skip the picker when it builds cleanly. A missing key / signed-out provider falls
	// back to onboarding from the safe defaults so the user can fix it.
	if onboard {
		if p, err := app.LoadPrefs(); err == nil && p.Provider != "" {
			saved := sess.cfg
			sess.cfg.provider = p.Provider
			if p.Model != "" {
				sess.cfg.model = p.Model
			}
			if sess.build() == nil {
				onboard = false
			} else {
				sess.cfg = saved // couldn't use the remembered provider; onboard from defaults
			}
		}
	}
	if sess.prov == nil {
		if err := sess.build(); err != nil {
			return err
		}
	}
	m := tui.New(sess.cfg.provider, sess.cfg.model, sess.respond).
		WithVersion(buildinfo.Get().Version).
		WithProviderMenu(providerChoices(), sess.selectProvider).
		WithModelSelect(sess.selectModel).
		WithEffortSelect(sess.selectEffort).
		WithCustomCommands(discoverCustomCommands(wd)).
		WithWorkspace(wd, gitBranch(ctx, wd)).
		WithActions(sess.sessionActions()).
		WithAgentRunner(sess.startAgentRun).
		WithProviderAuth(sess.startProviderAuth).
		WithProviderKeyEntry(providerKeyEnvFor, setProviderKey).
		WithHistory(restored)
	if onboard {
		m = m.WithOnboarding()
	}
	return tui.RunModel(ctx, m)
}

func newTUICommand() *cobra.Command {
	cfg := defaultTUIConfig()
	var resumeID string
	var continueLast bool
	c := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive terminal UI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resume := resumeID
			if continueLast && resume == "" {
				resume = app.LatestSessionID()
			}
			// Onboard (pick provider + model) unless the caller pinned a provider or is resuming.
			onboard := !cmd.Flags().Changed("provider") && resume == ""
			return launchTUIResume(cmd.Context(), cfg, onboard, resume)
		},
	}
	c.Flags().StringVar(&cfg.provider, "provider", cfg.provider, "provider name")
	c.Flags().StringVar(&cfg.baseURL, "base-url", "", "provider base URL")
	c.Flags().StringVar(&cfg.apiKeyEnv, "api-key-env", "", "environment variable holding the API key")
	c.Flags().StringVar(&cfg.model, "model", cfg.model, "model identifier")
	c.Flags().BoolVar(&cfg.allowWrite, "allow-write", false, "grant the agent write access")
	c.Flags().BoolVar(&cfg.allowExec, "allow-exec", false, "grant the agent command execution")
	c.Flags().StringVar(&resumeID, "resume", "", "resume a saved session by id")
	c.Flags().BoolVar(&continueLast, "continue", false, "resume the most recent session")
	return c
}
