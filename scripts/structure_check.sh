#!/usr/bin/env bash
# Structure check (FR-GH-002): fail when a mandatory repository entry is missing or a
# mandatory CODEOWNERS rule is absent. Runs in `make ci` and mirrors the CI structure job.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

fail=0
note() { echo "structure-check: MISSING $1"; fail=1; }

# Mandatory top-level entries (Volume 11 chapter 03 tree).
required_paths=(
  go.mod
  Makefile
  LICENSE
  README.md
  CHANGELOG.md
  CONTRIBUTING.md
  CODE_OF_CONDUCT.md
  SECURITY.md
  GOVERNANCE.md
  MAINTAINERS.md
  .golangci.yml
  cmd/andromeda
  internal
  docs/spec
  scripts
  .github/CODEOWNERS
  .github/PULL_REQUEST_TEMPLATE.md
  .github/dependabot.yml
  .github/labels.yml
  .github/workflows
  .github/ISSUE_TEMPLATE/config.yml
)

for p in "${required_paths[@]}"; do
  [ -e "$p" ] || note "$p"
done

# One issue form per chapter-05 issue type.
issue_forms=(
  bug feature security documentation performance refactor architecture research
  tech-debt release provider-integration tool-integration mcp plugin platform-compat
)
for f in "${issue_forms[@]}"; do
  [ -f ".github/ISSUE_TEMPLATE/${f}.yml" ] || note ".github/ISSUE_TEMPLATE/${f}.yml"
done

# Mandatory CODEOWNERS path rules.
co=".github/CODEOWNERS"
if [ -f "$co" ]; then
  for rule in "/internal/ports/" "/sdk/" "/docs/spec/" "/.github/workflows/" \
              "/internal/permission/" "/internal/sandbox/" "/internal/secret/" \
              "/.goreleaser.yaml" "/packaging/"; do
    grep -qF "$rule" "$co" || { echo "structure-check: MISSING CODEOWNERS rule $rule"; fail=1; }
  done
else
  note "$co"
fi

# Module count pinned to two (ADR-031): root + sdk. A third go.mod needs an ADR.
mod_count=$(git ls-files '**/go.mod' 'go.mod' | wc -l | tr -d ' ')
if [ "$mod_count" -gt 2 ]; then
  echo "structure-check: too many go.mod files ($mod_count); ADR-031 pins the module count to 2"
  fail=1
fi

if [ "$fail" -ne 0 ]; then
  echo "structure-check: FAILED"
  exit 1
fi
echo "structure-check: OK"
