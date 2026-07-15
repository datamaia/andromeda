package main

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	"github.com/datamaia/andromeda/internal/settingstore"
)

// errNoGH is returned when the GitHub CLI is not installed.
var errNoGH = errors.New("the GitHub CLI (gh) is not installed")

// ghRunner runs a gh subcommand with optional stdin and returns trimmed stdout. It is a package var
// so tests can stub it — /share and /unshare must never reach GitHub during a test run.
var ghRunner = defaultGHRunner

func defaultGHRunner(ctx context.Context, stdin string, args ...string) (string, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return "", errNoGH
	}
	cmd := exec.CommandContext(ctx, "gh", args...) //nolint:gosec // args are fixed literals + stored gist id
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.Output()
	if err != nil {
		return "", ghError(err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ghError surfaces gh's stderr (from exec.ExitError) so the user sees why it failed.
func ghError(err error) error {
	var ee *exec.ExitError
	if errors.As(err, &ee) && len(ee.Stderr) > 0 {
		return errors.New(strings.TrimSpace(string(ee.Stderr)))
	}
	return err
}

// shareAction backs /share: it uploads the transcript as a secret GitHub gist and remembers its id
// so /unshare can remove it. Secret (not public) by default, since a transcript may hold sensitive
// content; the returned message states plainly that it was uploaded.
func (s *tuiSession) shareAction(lines []string) string {
	var b strings.Builder
	b.WriteString("# Andromeda transcript\n\n")
	for _, l := range lines {
		b.WriteString(l + "\n\n")
	}
	url, err := ghRunner(s.ctx, b.String(),
		"gist", "create", "--filename", "andromeda-transcript.md",
		"--desc", "Andromeda session transcript", "-")
	if err != nil {
		if errors.Is(err, errNoGH) {
			return "share needs the GitHub CLI — install gh and run `gh auth login`"
		}
		return "share failed: " + err.Error()
	}
	if id := gistID(url); id != "" {
		st, _ := settingstore.Load(s.wd)
		st.LastGist = id
		_ = settingstore.Save(s.wd, st)
	}
	return "shared this transcript as a secret gist:\n  " + url + "\n\nremove it any time with /unshare"
}

// unshareAction backs /unshare: it deletes the gist last created by /share from this workspace.
func (s *tuiSession) unshareAction(_ context.Context) string {
	st, _ := settingstore.Load(s.wd)
	if st.LastGist == "" {
		return "nothing shared from this workspace yet — use /share first"
	}
	if _, err := ghRunner(s.ctx, "", "gist", "delete", st.LastGist); err != nil {
		if errors.Is(err, errNoGH) {
			return "unshare needs the GitHub CLI — install gh and run `gh auth login`"
		}
		return "unshare failed: " + err.Error()
	}
	id := st.LastGist
	st.LastGist = ""
	_ = settingstore.Save(s.wd, st)
	return "deleted shared gist " + id
}

// gistID extracts the trailing id from a gist URL (…/gists or gist.github.com/<user>/<id>).
func gistID(url string) string {
	url = strings.TrimSpace(url)
	if i := strings.LastIndex(url, "/"); i >= 0 && i < len(url)-1 {
		return url[i+1:]
	}
	return ""
}
