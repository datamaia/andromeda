#!/usr/bin/env bash
# Traceability validator (Volume 11 chapter 07, E-GH-001): verify that every non-merge commit
# in a range follows Conventional Commits and carries no attribution/advertising trailer. This
# is the CI mirror of the local .githooks/commit-msg hook; the two share this policy so a commit
# that bypassed the hook is still caught before merge.
#
# Usage: scripts/check_commits.sh <git-range>      e.g. scripts/check_commits.sh abc123..HEAD
set -euo pipefail

range="${1:?usage: check_commits.sh <git-range>}"

subject_re='^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9./-]+\))?!?: .{1,100}$'
attr_re='^[[:space:]]*(co-authored-by|generated[- ](with|by)|assisted[- ]by|signed-off-by:.*(claude|anthropic|codex|copilot|gemini|openai))'

fail=0
commits="$(git rev-list --no-merges "$range")"
if [ -z "$commits" ]; then
  echo "check-commits: no non-merge commits in ${range}; nothing to validate"
  exit 0
fi

for sha in $commits; do
  subject="$(git log -1 --format=%s "$sha")"
  body="$(git log -1 --format=%B "$sha")"
  short="$(git log -1 --format=%h "$sha")"

  case "$subject" in
    "fixup! "* | "squash! "* | "Revert "* | "Merge "*) ;; # tolerated prefixes
    *)
      if ! printf '%s' "$subject" | grep -qE "$subject_re"; then
        echo "::error::${short} subject is not a Conventional Commit: ${subject}"
        fail=1
      fi
      ;;
  esac

  if printf '%s' "$body" | grep -qiE "$attr_re" || printf '%s' "$body" | grep -q '🤖'; then
    echo "::error::${short} carries an attribution/advertising trailer (change information only)"
    fail=1
  fi
done

if [ "$fail" -ne 0 ]; then
  echo "check-commits: FAILED"
  exit 1
fi
echo "check-commits: all commits in ${range} conform"
