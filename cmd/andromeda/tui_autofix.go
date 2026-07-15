package main

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

// autofixPRAction backs /autofix-pr: it reads a pull request's failing CI checks via gh and, if any
// are red, returns a fix goal for the TUI to dispatch as an agent turn (so the fix runs interactively
// with approval prompts). It never pushes or comments — it only diagnoses and hands the agent a
// concrete goal. Returns (goal, status): an empty goal means nothing to do (status explains why).
func (s *tuiSession) autofixPRAction(_ context.Context, args string) (goal, status string) {
	pr := strings.TrimSpace(args)
	if pr == "" {
		out, err := ghRunner(s.ctx, "", "pr", "view", "--json", "number", "-q", ".number")
		if err != nil {
			if errors.Is(err, errNoGH) {
				return "", "autofix-pr needs the GitHub CLI — install gh and run `gh auth login`"
			}
			return "", "autofix-pr: no PR for this branch — pass a number, e.g. /autofix-pr 42"
		}
		pr = strings.TrimSpace(out)
	}
	if pr == "" {
		return "", "autofix-pr: no PR found — pass a number, e.g. /autofix-pr 42"
	}
	// gh pr view exits 0 even when checks fail (unlike gh pr checks), so it captures cleanly.
	failing, err := ghRunner(s.ctx, "",
		"pr", "view", pr, "--json", "statusCheckRollup",
		"-q", `.statusCheckRollup[] | select((.conclusion // .state) == "FAILURE") | .name`)
	if err != nil {
		if errors.Is(err, errNoGH) {
			return "", "autofix-pr needs the GitHub CLI — install gh and run `gh auth login`"
		}
		return "", "autofix-pr: " + err.Error()
	}
	names := nonEmptyLines(failing)
	if len(names) == 0 {
		return "", "CI is green on PR #" + pr + " — nothing to fix"
	}
	goal = "CI is failing on pull request #" + pr + ". These checks are red:\n" +
		"- " + strings.Join(names, "\n- ") + "\n\n" +
		"Investigate why they fail — read the relevant workflow configuration and reproduce the " +
		"failures locally (run the tests, linters, or build they invoke) — then fix the code so they " +
		"would pass. Do not commit or push; I'll review the changes."
	return goal, "found " + plural(len(names), "failing check") + " on PR #" + pr + " — starting a fix"
}

// nonEmptyLines splits output into trimmed non-empty lines.
func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// plural renders "1 thing" / "N things".
func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}
