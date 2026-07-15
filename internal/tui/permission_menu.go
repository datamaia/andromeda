package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// The /permission command manages the command allowlist: argv-prefix rules that let the agent run a
// vetted command without an approval prompt (allow) or that it must always refuse (deny). Rules are
// stored in the workspace's .andromeda/permissions.toml (so both lists are visible and committable)
// and merged with andromeda.toml's [permission] section at runtime. The driver supplies the current
// policy through Actions.Permissions and applies edits through Actions.Permission, keeping this
// package free of filesystem/config imports.

// PermRule is one allow/deny entry surfaced in the /permission menu.
type PermRule struct {
	Command string
	Managed bool // true when it lives in .andromeda (removable here); false when it comes from andromeda.toml
}

// PermissionView is the current command policy, as surfaced to the menu.
type PermissionView struct {
	Allow []PermRule
	Deny  []PermRule
	Path  string // where managed rules live (.andromeda/permissions.toml), shown to the user
}

// cmdPermission dispatches the text subcommands (/permission allow|deny <cmd>, rm <list> <cmd>,
// list) to the Permission action, and opens the interactive menu on a bare /permission.
func cmdPermission(m Model, args string) (tea.Model, tea.Cmd) {
	if strings.TrimSpace(args) != "" {
		if m.actions.Permission == nil {
			return m.unavailable("permission"), nil
		}
		return m.sys(m.actions.Permission(context.Background(), args)), nil
	}
	return m.openPermissionMenu()
}

// openPermissionMenu shows the allow and deny lists (each drilling into its entries) and offers to
// add a rule to either, with a friendly empty state.
func (m Model) openPermissionMenu() (tea.Model, tea.Cmd) {
	var v PermissionView
	if m.actions.Permissions != nil {
		v = m.actions.Permissions(context.Background())
	}
	lvl := menuLevel{title: "Permissions · command allowlist"}
	if len(v.Allow) == 0 && len(v.Deny) == 0 {
		lvl.hint = "No rules yet — pre-approve safe commands (allow) or block dangerous ones (deny)."
	}
	allow, deny := v.Allow, v.Deny // capture for the drill-in closures
	lvl.items = append(lvl.items,
		menuItem{
			label: fmt.Sprintf("Allow (%d)", len(allow)),
			desc:  "commands the agent may run without asking",
			run:   func(mm Model) (Model, tea.Cmd) { return mm.pushPermissionList("allow", "Allow", allow), nil },
		},
		menuItem{
			label: fmt.Sprintf("Deny (%d)", len(deny)),
			desc:  "commands the agent is always refused",
			run:   func(mm Model) (Model, tea.Cmd) { return mm.pushPermissionList("deny", "Deny", deny), nil },
		},
		menuItem{
			label: "＋ Allow a command",
			desc:  `pre-approve, e.g. "git status" — matched by argv prefix`,
			run: func(mm Model) (Model, tea.Cmd) {
				mm = mm.closeMenu()
				mm.input = "/permission allow "
				return mm.sys(`type the command to pre-approve (argv prefix), then enter — e.g. "go test ./..."`), nil
			},
		},
		menuItem{
			label: "＋ Deny a command",
			desc:  `always refuse, e.g. "git push --force"`,
			run: func(mm Model) (Model, tea.Cmd) {
				mm = mm.closeMenu()
				mm.input = "/permission deny "
				return mm.sys(`type the command to always refuse (argv prefix), then enter — e.g. "rm -rf"`), nil
			},
		},
	)
	if v.Path != "" {
		lvl.items = append(lvl.items, menuItem{label: "Stored in", desc: v.Path}) // info row (nil run)
	}
	return m.openMenu(lvl)
}

// pushPermissionList drills into one list. Managed rules open a detail view with a Remove action;
// rules inherited from andromeda.toml are shown read-only (edit them in that file).
func (m Model) pushPermissionList(list, title string, rules []PermRule) Model {
	lvl := menuLevel{title: title}
	for _, r := range rules {
		r := r
		item := menuItem{label: r.Command}
		if r.Managed {
			item.desc = "enter to remove"
			item.run = func(mm Model) (Model, tea.Cmd) { return mm.pushPermissionRule(list, r), nil }
		} else {
			item.desc = "from andromeda.toml (read-only here)"
		}
		lvl.items = append(lvl.items, item)
	}
	if len(rules) == 0 {
		lvl.hint = "Nothing here yet — add one from the previous menu."
	}
	return m.pushMenu(lvl)
}

// pushPermissionRule drills into a single managed rule, offering to remove it.
func (m Model) pushPermissionRule(list string, r PermRule) Model {
	lvl := menuLevel{title: r.Command, hint: list + " rule · matched by argv prefix"}
	lvl.items = append(lvl.items, menuItem{
		label: "Remove",
		desc:  "delete this rule from .andromeda/permissions.toml",
		run: func(mm Model) (Model, tea.Cmd) {
			if mm.actions.Permission != nil {
				mm = mm.sys(mm.actions.Permission(context.Background(), "rm "+list+" "+r.Command))
			}
			nm, cmd := mm.openPermissionMenu() // reopen the refreshed policy
			return nm.(Model), cmd
		},
	})
	return m.pushMenu(lvl)
}
