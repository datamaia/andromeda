package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/buildinfo"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/updater"
)

// updateRepo is the GitHub repository whose releases the updater checks.
const updateRepo = "datamaia/andromeda"

// githubReleaseSource is a read-only updater.ReleaseSource backed by the GitHub Releases API. It
// answers "what is the latest published version?" so `andromeda update` can tell the user whether a
// newer release exists. It does not download or apply anything — real upgrades go through brew or the
// install script — so Fetch is intentionally unsupported.
type githubReleaseSource struct {
	repo string
	hc   *http.Client
}

func newGitHubReleaseSource() githubReleaseSource {
	return githubReleaseSource{repo: updateRepo, hc: &http.Client{Timeout: 6 * time.Second}}
}

// Latest returns the newest published (non-draft, non-prerelease) release tag, normalized without a
// leading "v" to match buildinfo's version string. channel is ignored: only stable releases are
// published to the tap/installer today.
func (g githubReleaseSource) Latest(ctx context.Context, _ string) (version, checksum string, err error) {
	url := "https://api.github.com/repos/" + g.repo + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := g.hc.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github releases: HTTP %d", resp.StatusCode)
	}
	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", "", err
	}
	return strings.TrimPrefix(body.TagName, "v"), "", nil
}

// Fetch is unsupported: upgrades happen via brew or the install script, not in-place self-replacement.
func (githubReleaseSource) Fetch(context.Context, string) (string, string, error) {
	return "", "", errors.New("in-place update is not supported; use brew or the install script")
}

// checkForUpdate runs a descriptive update check and returns human-readable text. It never returns an
// error: an unreachable feed or a dev build is reported as part of the message, not as a failure.
func checkForUpdate(ctx context.Context, channel, targetPath string) string {
	u := updater.New(buildinfo.Get().Version, channel, targetPath, newGitHubReleaseSource())
	res, err := u.Check(ctx)
	return describeUpdate(res, buildinfo.Get(), err)
}

// describeUpdate renders an actionable update report: what you're running, and what to do next.
func describeUpdate(res ports.UpdateCheckResult, i buildinfo.Info, checkErr error) string {
	channel := res.Channel
	if channel == "" {
		channel = "stable"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "andromeda %s · channel %s · %s/%s\n", i.Version, channel, i.GoOS, i.GoArch)
	fmt.Fprintf(&b, "  build %s (%s)", shortCommit(i.Commit), orDash(i.Date))

	switch {
	case isDevVersion(i.Version):
		b.WriteString("\n  development build — the auto-updater only tracks published releases.")
		if res.Latest != "" && res.Latest != i.Version {
			fmt.Fprintf(&b, "\n  latest published release: v%s", res.Latest)
		}
		b.WriteString("\n" + installHint())
	case checkErr != nil:
		fmt.Fprintf(&b, "\n  couldn't reach the release feed (%v)\n  you're on v%s", compactErr(checkErr), i.Version)
	case res.Status == "update_available":
		fmt.Fprintf(&b, "\n  ✨ update available: v%s → v%s\n%s", i.Version, res.Latest, installHint())
	default:
		fmt.Fprintf(&b, "\n  ✓ you're on the latest release (v%s)", i.Version)
	}
	return b.String()
}

// installHint lists the sanctioned upgrade paths.
func installHint() string {
	return "  update with:  brew upgrade andromeda\n" +
		"           or:  curl -fsSL https://andromedacli.com/install | bash"
}

// isDevVersion reports whether v is a from-source development build (not a published release).
func isDevVersion(v string) bool {
	return v == "" || v == "dev" || strings.Contains(v, "dev")
}

func shortCommit(c string) string {
	if len(c) > 10 {
		return c[:10]
	}
	if c == "" {
		return "unknown"
	}
	return c
}

func orDash(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

// compactErr trims a wrapped error to its leaf message for a one-line report.
func compactErr(err error) string {
	msg := err.Error()
	if i := strings.LastIndex(msg, ": "); i >= 0 {
		return msg[i+2:]
	}
	return msg
}
