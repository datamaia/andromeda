# Andromeda developer gate.
#
# `make ci` is the authoritative quality gate: it runs locally with no dependency on
# GitHub Actions, so development is never blocked by CI-minute limits (the repository is
# private; Actions mirrors a lean subset of these targets — see .github/workflows/ci.yml
# and STATUS.md). Every CI job invokes one of these targets.

SHELL := /bin/bash
GO ?= go
PKGS := ./...
# SM-14 MVP overall floor; ramps to 80 at v1 (Volume 13)
COVERAGE_MIN ?= 70
COVER_PROFILE := coverage.out

.DEFAULT_GOAL := help

.PHONY: help
help: ## List available targets
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Sync go.mod/go.sum
	$(GO) mod tidy

.PHONY: fmt
fmt: ## Format sources with gofmt
	@gofmt -w $$(git ls-files '*.go')

.PHONY: fmt-check
fmt-check: ## Fail if any Go file is not gofmt-clean
	@diff=$$(gofmt -l $$(git ls-files '*.go')); \
	if [ -n "$$diff" ]; then echo "gofmt needed:"; echo "$$diff"; exit 1; fi

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKGS)

.PHONY: lint
lint: fmt-check vet ## Static analysis (golangci-lint if installed, always gofmt+vet)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "note: golangci-lint not installed; ran gofmt+vet only (see .golangci.yml)"; \
	fi

.PHONY: build
build: ## Build all packages
	$(GO) build $(PKGS)

.PHONY: test
test: ## Run all tests with the race detector
	$(GO) test -race -timeout 120s $(PKGS)

.PHONY: test-unit
test-unit: ## Run unit tests with the race detector
	$(GO) test -race -timeout 120s $(PKGS)

.PHONY: test-integration
test-integration: ## Run integration-tagged tests
	$(GO) test -race -tags=integration $(PKGS)

.PHONY: coverage
coverage: ## Produce a coverage profile
	$(GO) test -covermode=atomic -coverprofile=$(COVER_PROFILE) $(PKGS)

.PHONY: coverage-gate
coverage-gate: coverage ## Fail if total coverage is below COVERAGE_MIN (SM-14)
	@total=$$($(GO) tool cover -func=$(COVER_PROFILE) | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "total coverage: $$total% (min $(COVERAGE_MIN)%)"; \
	awk -v t="$$total" -v m="$(COVERAGE_MIN)" 'BEGIN{exit !(t+0 >= m+0)}' \
		|| { echo "coverage below threshold"; exit 1; }

.PHONY: spec-lint
spec-lint: ## Run the specification linter over docs/spec
	python3 scripts/spec_lint.py

.PHONY: structure-check
structure-check: ## Verify the mandatory repository layout (FR-GH-002)
	@bash scripts/structure_check.sh

.PHONY: ci
ci: tidy-check lint build test coverage-gate spec-lint structure-check ## Full local gate (mirrors CI)
	@echo "ci: all gates passed"

.PHONY: tidy-check
tidy-check: ## Fail if go.mod/go.sum are not tidy
	@cp go.mod go.mod.bak; cp go.sum go.sum.bak 2>/dev/null || true; \
	$(GO) mod tidy; \
	if ! diff -q go.mod go.mod.bak >/dev/null 2>&1; then \
		echo "go.mod not tidy; run 'make tidy'"; mv go.mod.bak go.mod; mv go.sum.bak go.sum 2>/dev/null || true; exit 1; \
	fi; \
	rm -f go.mod.bak go.sum.bak

.PHONY: clean
clean: ## Remove build and coverage artifacts
	@rm -f $(COVER_PROFILE); $(GO) clean
