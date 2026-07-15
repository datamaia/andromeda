package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/auth"
	"github.com/datamaia/andromeda/internal/graph"
	"github.com/datamaia/andromeda/internal/memnote"
	"github.com/datamaia/andromeda/internal/ontology"
	"github.com/datamaia/andromeda/internal/permstore"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/skill"
	"github.com/datamaia/andromeda/internal/tui"
)

// sessionActions wires the TUI slash commands to the real app operations. Each handler returns the
// text the palette shows in the transcript; errors are formatted, not thrown, so a failing command
// never tears down the session.
func (s *tuiSession) sessionActions() tui.Actions {
	return tui.Actions{
		Doctor:      s.doctorAction,
		Update:      s.updateAction,
		Memory:      s.memoryAction,
		MemoryList:  s.memoryListAction,
		Collection:  s.collectionAction,
		Models:      s.modelsAction,
		Config:      s.configAction,
		Logout:      s.logoutAction,
		Export:      s.exportAction,
		Init:        s.initAction,
		Files:       s.listFiles,
		Context:     s.contextAction,
		Ontology:    s.ontologyAction,
		Graph:       s.graphAction,
		Skills:      s.skillListAction,
		Permission:  s.permissionAction,
		Permissions: s.permissionView,
	}
}

// permissionView backs the /permission menu: the workspace-managed allow/deny rules (from
// .andromeda/permissions.toml, removable in the menu) plus any inherited from andromeda.toml (shown
// read-only), so the user sees the full effective command policy in one place.
func (s *tuiSession) permissionView(ctx context.Context) tui.PermissionView {
	v := tui.PermissionView{Path: relOr(s.wd, permstore.Path(s.wd))}
	if r, err := permstore.Load(s.wd); err == nil {
		for _, c := range r.Allow {
			v.Allow = append(v.Allow, tui.PermRule{Command: c, Managed: true})
		}
		for _, c := range r.Deny {
			v.Deny = append(v.Deny, tui.PermRule{Command: c, Managed: true})
		}
	}
	if cfg, err := app.LoadedConfig(ctx, s.wd); err == nil {
		for _, c := range configStrings(cfg.Values["permission.allow"]) {
			v.Allow = append(v.Allow, tui.PermRule{Command: c})
		}
		for _, c := range configStrings(cfg.Values["permission.deny"]) {
			v.Deny = append(v.Deny, tui.PermRule{Command: c})
		}
	}
	return v
}

// permissionAction runs the /permission text subcommands, editing the .andromeda store and returning
// a status line. Grammar: allow <cmd> · deny <cmd> · rm allow|deny <cmd> · list.
func (s *tuiSession) permissionAction(ctx context.Context, args string) string {
	fields := strings.Fields(args)
	if len(fields) == 0 {
		return s.permissionList(ctx)
	}
	switch fields[0] {
	case "allow", "deny":
		cmd := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(args), fields[0]))
		if cmd == "" {
			return "usage: /permission " + fields[0] + " <command>   (e.g. " + fields[0] + " git status)"
		}
		if _, err := permstore.Add(s.wd, fields[0], cmd); err != nil {
			return "permission: " + err.Error()
		}
		return "added to " + fields[0] + ": " + cmd + "  ·  saved to " + relOr(s.wd, permstore.Path(s.wd))
	case "rm", "remove", "delete":
		rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(args), fields[0]))
		rf := strings.Fields(rest)
		if len(rf) < 2 || (rf[0] != "allow" && rf[0] != "deny") {
			return "usage: /permission rm allow|deny <command>"
		}
		cmd := strings.TrimSpace(strings.TrimPrefix(rest, rf[0]))
		if _, err := permstore.Remove(s.wd, rf[0], cmd); err != nil {
			return "permission: " + err.Error()
		}
		return "removed from " + rf[0] + ": " + cmd
	case "list", "show":
		return s.permissionList(ctx)
	default:
		return "permission subcommands: allow <cmd> · deny <cmd> · rm allow|deny <cmd> · list"
	}
}

// permissionList renders the effective command policy: managed rules from .andromeda plus any
// inherited from andromeda.toml, with ✓ for allow and ✗ for deny.
func (s *tuiSession) permissionList(ctx context.Context) string {
	r, _ := permstore.Load(s.wd)
	var b strings.Builder
	fmt.Fprintf(&b, "permission · allow %d · deny %d  (%s)", len(r.Allow), len(r.Deny), relOr(s.wd, permstore.Path(s.wd)))
	for _, c := range r.Allow {
		fmt.Fprintf(&b, "\n  ✓ %s", c)
	}
	for _, c := range r.Deny {
		fmt.Fprintf(&b, "\n  ✗ %s", c)
	}
	if cfg, err := app.LoadedConfig(ctx, s.wd); err == nil {
		for _, c := range configStrings(cfg.Values["permission.allow"]) {
			fmt.Fprintf(&b, "\n  ✓ %s  (andromeda.toml)", c)
		}
		for _, c := range configStrings(cfg.Values["permission.deny"]) {
			fmt.Fprintf(&b, "\n  ✗ %s  (andromeda.toml)", c)
		}
	}
	return b.String()
}

// configStrings coerces a resolved config value (a TOML string array or a single string) to a slice
// of trimmed, non-empty strings — the driver-side mirror of app.configStringSlice.
func configStrings(v any) []string {
	var out []string
	switch a := v.(type) {
	case []string:
		for _, sv := range a {
			if sv = strings.TrimSpace(sv); sv != "" {
				out = append(out, sv)
			}
		}
	case []any:
		for _, e := range a {
			if sv, ok := e.(string); ok {
				if sv = strings.TrimSpace(sv); sv != "" {
					out = append(out, sv)
				}
			}
		}
	case string:
		if sv := strings.TrimSpace(a); sv != "" {
			out = append(out, sv)
		}
	}
	return out
}

// skillListAction backs the $-mention skill invocation: it discovers the workspace's skills (across
// .agents/.claude/.codex/.agent) and returns them with their instruction bodies so the TUI can both
// complete "$name" and fold the selected skill's instructions into the run.
func (s *tuiSession) skillListAction(_ context.Context) []tui.SkillNote {
	ds := skill.Discover(s.wd)
	out := make([]tui.SkillNote, 0, len(ds))
	for _, d := range ds {
		out = append(out, tui.SkillNote{
			Name:        d.Manifest.Name,
			Description: d.Manifest.Description,
			Path:        d.Path,
			Body:        d.Prompt,
		})
	}
	return out
}

// ontologyAction backs the /ontology slash command: it scans the workspace and manages the
// deterministic Turtle map under .andromeda/ontology/. op is build | show | rm.
func (s *tuiSession) ontologyAction(ctx context.Context, op string) string {
	switch op {
	case "build":
		m, err := ontology.Scan(ctx, s.wd)
		if err != nil {
			return "ontology: " + err.Error()
		}
		path, err := ontology.Write(s.wd, m)
		if err != nil {
			return "ontology: " + err.Error()
		}
		return "ontology · " + m.Stats() + "\n  written to " + relOr(s.wd, path)
	case "show":
		data, err := os.ReadFile(filepath.Join(ontology.Dir(s.wd), "project.ttl")) //nolint:gosec // fixed path under the workspace marker dir
		if err != nil {
			if os.IsNotExist(err) {
				return "no ontology yet — run /ontology build"
			}
			return "ontology: " + err.Error()
		}
		return string(data)
	case "rm":
		if err := ontology.Remove(s.wd); err != nil {
			return "ontology: " + err.Error()
		}
		return "removed .andromeda/ontology"
	default:
		return "ontology subcommands: build · show · rm"
	}
}

// graphAction backs the /graph slash command: it scans the workspace, writes the visual graph model
// (graph.json + Markdown notes), and can serve the interactive viewer. op is build | open | show | rm.
func (s *tuiSession) graphAction(ctx context.Context, op string) string {
	switch op {
	case "build":
		m, err := ontology.Scan(ctx, s.wd)
		if err != nil {
			return "graph: " + err.Error()
		}
		g, dir, err := graph.Write(s.wd, m)
		if err != nil {
			return "graph: " + err.Error()
		}
		return "graph · " + g.Stats() + "\n  written to " + relOr(s.wd, dir)
	case "open":
		return s.graphOpen(ctx)
	case "show":
		data, err := os.ReadFile(filepath.Join(graph.Dir(s.wd), "index.md")) //nolint:gosec // fixed path under the workspace marker dir
		if err != nil {
			if os.IsNotExist(err) {
				return "no graph yet — run /graph build"
			}
			return "graph: " + err.Error()
		}
		return string(data)
	case "rm":
		if err := graph.Remove(s.wd); err != nil {
			return "graph: " + err.Error()
		}
		return "removed .andromeda/graph"
	default:
		return "graph subcommands: build · open · show · rm"
	}
}

// graphOpen rebuilds the graph (so the viewer reflects the current tree) and serves the interactive
// viewer on localhost, opening the system browser. The server is bound to the program lifetime, so a
// second /graph open just reopens the already-running viewer.
func (s *tuiSession) graphOpen(ctx context.Context) string {
	m, err := ontology.Scan(ctx, s.wd)
	if err != nil {
		return "graph: " + err.Error()
	}
	if _, _, err := graph.Write(s.wd, m); err != nil {
		return "graph: " + err.Error()
	}
	if s.graphURL != "" {
		_ = openBrowser(s.graphURL)
		return "graph viewer already running at " + s.graphURL + " — reopened in your browser"
	}
	ready := make(chan string, 1)
	go func() {
		_ = graph.Serve(s.ctx, s.wd, 0, func(url string) { ready <- url })
	}()
	select {
	case url := <-ready:
		s.graphURL = url
		_ = openBrowser(url)
		return "graph viewer serving at " + url + "\n  opened in your browser · stays up for this session"
	case <-time.After(3 * time.Second):
		return "graph: viewer did not start in time"
	}
}

// relOr renders path relative to base for display, falling back to the absolute path.
func relOr(base, path string) string {
	if rel, err := filepath.Rel(base, path); err == nil {
		return rel
	}
	return path
}

// initAction scaffolds the project layout — AGENTS.md, andromeda.toml, .agents/, and .andromeda/ —
// in the workspace root, seeding andromeda.toml with the session's current provider and model. It is
// idempotent: missing pieces are created, an existing andromeda.toml is augmented with any config
// sections it lacks, and existing files are never overwritten. See scaffoldProject.
func (s *tuiSession) initAction(_ context.Context, provider, model string) string {
	return scaffoldProject(s.wd, provider, model)
}

func (s *tuiSession) configAction(ctx context.Context) string {
	cfg, err := app.LoadedConfig(ctx, s.wd)
	if err != nil {
		return "config: " + err.Error()
	}
	keys := make([]string, 0, len(cfg.Values))
	for k := range cfg.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	fmt.Fprintf(&b, "config · %d keys resolved", len(keys))
	for _, k := range keys {
		src := cfg.Sources[k]
		if src == "" {
			src = "default"
		}
		fmt.Fprintf(&b, "\n  %-30s %v  (%s)", k, cfg.Values[k], src)
	}
	return b.String()
}

func (s *tuiSession) logoutAction(ctx context.Context, provider string) string {
	if provider == auth.OpenAIChatGPTProvider {
		mgr, err := newAuthManager()
		if err != nil {
			return "logout: " + err.Error()
		}
		if err := mgr.Revoke(ctx, ports.AuthenticationHandle{Provider: provider, Profile: "default"}); err != nil {
			return "logout: " + err.Error()
		}
		return "signed out of ChatGPT — pick a provider to sign in again"
	}
	if info, ok := app.LookupProvider(provider); ok && info.KeyEnv != "" {
		_ = os.Unsetenv(info.KeyEnv)
		return "cleared the API key for " + provider + " (this session)"
	}
	return "nothing to sign out of for " + provider
}

func (s *tuiSession) exportAction(lines []string) string {
	name := "andromeda-transcript-" + time.Now().Format("20060102-150405") + ".md"
	path := filepath.Join(s.wd, name)
	var b strings.Builder
	b.WriteString("# Andromeda transcript\n\n")
	for _, l := range lines {
		b.WriteString(l + "\n\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		return "export failed: " + err.Error()
	}
	return "saved transcript to " + path
}

func (s *tuiSession) doctorAction(ctx context.Context) string {
	rep, err := app.Doctor(ctx, s.wd)
	if err != nil {
		return "doctor: " + err.Error()
	}
	var b strings.Builder
	status := "all checks passed"
	if !rep.OK() {
		status = "some checks failed"
	}
	b.WriteString("doctor · " + status)
	for _, c := range rep.Checks {
		mark := "✓"
		if !c.OK {
			mark = "✗"
		}
		fmt.Fprintf(&b, "\n  %s %-13s %s", mark, c.Name, c.Detail)
	}
	return b.String()
}

func (s *tuiSession) updateAction(ctx context.Context) string {
	self, _ := os.Executable()
	return checkForUpdate(ctx, "stable", self)
}

// memoryAction backs the /memory text subcommands over the file-based note store (memnote): a folder
// of Markdown notes under .andromeda/memory/ with a generated MEMORY.md index, alongside AGENTS.md.
func (s *tuiSession) memoryAction(_ context.Context, args string) string {
	sub, rest, _ := strings.Cut(strings.TrimSpace(args), " ")
	rest = strings.TrimSpace(rest)
	switch sub {
	case "add":
		if rest == "" {
			return "usage: /memory add <title> [#tag …]"
		}
		title, tags := extractTags(rest)
		n, err := memnote.Add(s.wd, title, tags, "")
		if err != nil {
			return "memory: " + err.Error()
		}
		return "remembered " + n.ID + " · " + n.Title
	case "search":
		if rest == "" {
			return "usage: /memory search <query>"
		}
		hits, err := memnote.Search(s.wd, rest)
		return formatNotes(fmt.Sprintf("search %q", rest), hits, err)
	case "rm", "delete":
		if rest == "" {
			return "usage: /memory rm <id>"
		}
		if err := memnote.Delete(s.wd, rest); err != nil {
			return "memory: " + err.Error()
		}
		return "deleted memory " + rest
	case "", "list":
		notes, err := memnote.List(s.wd)
		return formatNotes("memory", notes, err)
	default:
		return "memory subcommands: list · add <title> [#tag] · search <query> · rm <id>"
	}
}

// memoryListAction returns the notes for the interactive /memory menu.
func (s *tuiSession) memoryListAction(_ context.Context) []tui.MemoryNote {
	notes, _ := memnote.List(s.wd)
	out := make([]tui.MemoryNote, 0, len(notes))
	for _, n := range notes {
		out = append(out, tui.MemoryNote{
			ID: n.ID, Title: n.Title, Tags: n.Tags, Created: n.Created,
			Preview: firstLine(n.Body), Path: relOr(s.wd, n.Path(s.wd)),
		})
	}
	return out
}

func formatNotes(label string, notes []memnote.Note, err error) string {
	if err != nil {
		return "memory: " + err.Error()
	}
	if len(notes) == 0 {
		return "no memories yet — add one with /memory add <title>"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s · %d note(s)", label, len(notes))
	for _, n := range notes {
		tags := ""
		if len(n.Tags) > 0 {
			tags = " [" + strings.Join(n.Tags, ", ") + "]"
		}
		fmt.Fprintf(&b, "\n  %s  %s%s", n.ID, n.Title, tags)
	}
	return b.String()
}

// extractTags pulls inline #tags out of an add line, returning the remaining title and the tags.
func extractTags(s string) (title string, tags []string) {
	var words []string
	for _, w := range strings.Fields(s) {
		if strings.HasPrefix(w, "#") && len(w) > 1 {
			tags = append(tags, strings.TrimPrefix(w, "#"))
			continue
		}
		words = append(words, w)
	}
	return strings.Join(words, " "), tags
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

// collectionAction backs the interactive /skills, /mcp, /workflows and /plugins menus, returning the
// entries of each manageable capability set (with a friendly empty state when there are none).
func (s *tuiSession) collectionAction(_ context.Context, kind string) tui.CollectionView {
	switch kind {
	case "skills":
		return s.skillCollection()
	case "mcp":
		return s.mcpCollection()
	case "workflows":
		return s.workflowCollection()
	case "plugins":
		return s.pluginCollection()
	}
	return tui.CollectionView{}
}

func (s *tuiSession) skillCollection() tui.CollectionView {
	v := tui.CollectionView{
		Empty:  "No skills yet.",
		Create: "add .agents/skills/<name>/SKILL.md (also scanned: .claude/.codex/.agent)",
	}
	for _, d := range skill.Discover(s.wd) {
		title := d.Manifest.Name
		if d.Manifest.Version != "" {
			title = fmt.Sprintf("%s@%s", d.Manifest.Name, d.Manifest.Version)
		}
		detail := d.Manifest.Description
		if d.Source != ".agents" { // note non-default sources so the origin is clear
			detail = strings.TrimSpace(detail + "  ·  " + d.Source)
		}
		v.Entries = append(v.Entries, tui.CollectionEntry{Title: title, Detail: detail, Path: d.Path})
	}
	sort.Slice(v.Entries, func(i, j int) bool { return v.Entries[i].Title < v.Entries[j].Title })
	return v
}

func (s *tuiSession) mcpCollection() tui.CollectionView {
	v := tui.CollectionView{
		Empty:  "No MCP servers configured.",
		Create: "declare one under [mcp.servers] in andromeda.toml, or add a file in .agents/mcp/",
	}
	dir := filepath.Join(s.wd, ".agents", "mcp")
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		v.Entries = append(v.Entries, tui.CollectionEntry{
			Title: e.Name(), Detail: "MCP server config", Path: filepath.Join(dir, e.Name()),
		})
	}
	return v
}

// workflowDirs are the workspace-relative directories scanned for step-by-step workflow recipes, in
// precedence order (first name wins). A workflow is a Markdown file — optional YAML front matter
// (description) plus numbered instructions — following the .windsurf/.cursor convention; .agents and
// .andromeda are Andromeda's own homes.
var workflowDirs = []string{".agents/workflows", ".andromeda/workflows", ".windsurf/workflows", ".cursor/workflows"}

func (s *tuiSession) workflowCollection() tui.CollectionView {
	v := tui.CollectionView{
		Empty:  "No workflows yet.",
		Create: "add a recipe at .agents/workflows/<name>.md (also scanned: .andromeda/.windsurf/.cursor)",
	}
	seen := map[string]bool{}
	for _, d := range workflowDirs {
		dir := filepath.Join(s.wd, d)
		ents, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range ents {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".md")
			if name == "" || seen[name] {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, e.Name())) //nolint:gosec // reads a discovered workflow file under the workspace
			if err != nil {
				continue
			}
			seen[name] = true
			desc, body := parseFrontMatter(string(data))
			body = strings.TrimSpace(body)
			if strings.TrimSpace(desc) == "" {
				desc = firstLine(body)
			}
			base := strings.SplitN(d, "/", 2)[0]
			if base != ".agents" { // note non-default sources so the origin is clear
				desc = strings.TrimSpace(desc + "  ·  " + base)
			}
			v.Entries = append(v.Entries, tui.CollectionEntry{
				Title: name, Detail: desc, Path: filepath.Join(dir, e.Name()), Body: body,
			})
		}
	}
	sort.Slice(v.Entries, func(i, j int) bool { return v.Entries[i].Title < v.Entries[j].Title })
	return v
}

func (s *tuiSession) pluginCollection() tui.CollectionView {
	v := tui.CollectionView{
		Empty:  "No plugins installed.",
		Create: "declare one under [plugins] in andromeda.toml",
	}
	dir := filepath.Join(s.wd, ".agents", "plugins")
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		v.Entries = append(v.Entries, tui.CollectionEntry{
			Title: e.Name(), Detail: "plugin", Path: filepath.Join(dir, e.Name()),
		})
	}
	return v
}

func (s *tuiSession) modelsAction(ctx context.Context) []string {
	if s.prov == nil {
		return nil
	}
	// Bound discovery so a slow or unresponsive provider can't stall model selection.
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	descs, err := s.prov.DiscoverModels(ctx)
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(descs))
	for _, d := range descs {
		if chatModelUsable(d.ID) {
			ids = append(ids, d.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

// chatModelUsable filters out models the agent cannot drive as a chat/completions model
// (embeddings, image/audio/video, moderation, reranking), so the model picker lists only usable
// ones. It errs toward inclusion — anything not clearly special-purpose is kept.
func chatModelUsable(id string) bool {
	l := strings.ToLower(id)
	for _, bad := range []string{
		"embed", "imagen", "-image", "image-", "veo", "-tts", "tts-", "whisper",
		"aqa", "rerank", "guard", "moderat", "-vision", "dall-e", "sora", "-audio", "audio-",
	} {
		if strings.Contains(l, bad) {
			return false
		}
	}
	return true
}
