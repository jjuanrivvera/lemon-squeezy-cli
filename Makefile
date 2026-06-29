# Makefile — lsqueezy CLI (built with cliwright). `make verify` is the acceptance gate.
BINARY      := lsqueezy
MODULE      := github.com/jjuanrivvera/lemon-squeezy-cli
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE        ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w \
  -X $(MODULE)/internal/version.Version=$(VERSION) \
  -X $(MODULE)/internal/version.Commit=$(COMMIT) \
  -X $(MODULE)/internal/version.Date=$(DATE)
COVERAGE_MIN ?= 80

.DEFAULT_GOAL := build

## --- build & run ---
# NOTE: cliwright's template assumes ./cmd/$(BINARY); this project keeps main.go at the
# repo root (per the build brief), so the build target points at "." instead.
build: ## build to bin/$(BINARY)
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o bin/$(BINARY) .
install: ## go install the binary
	CGO_ENABLED=0 go install -ldflags '$(LDFLAGS)' .
uninstall: ; rm -f "$$(go env GOPATH)/bin/$(BINARY)"
run: build ; ./bin/$(BINARY) $(ARGS)
dev: fmt vet build ## fast local cycle

## --- quality gate ---
fmt: ; gofmt -s -w .
vet: ; go vet ./...
lint: ; golangci-lint run ./... || (echo "golangci-lint missing or failed" >&2; exit 1)
tidy: ; go mod tidy
test: ; go test ./...
test-race: ; go test -race ./...
# -coverpkg=./... credits the cross-package integration tests (commands_test exercises the
# resources + internal/api code paths end-to-end), so coverage reflects what the suite
# actually drives, not just same-package unit tests.
test-coverage: ; go test ./... -coverpkg=./... -coverprofile=coverage.out
cover-check: test-coverage ; ./scripts/cover-check.sh $(COVERAGE_MIN)
check: fmt vet lint test ## the local quality gate

## --- the acceptance gate (cliwright) ---
# verify == "done and high". The /goal completion promise binds to this exiting 0.
verify: check spec-check cover-check ## full acceptance gate; exit 0 == done
	./scripts/dod-check.sh $(BINARY)
	./scripts/judge.sh
spec-check: ## built CLI surface must match the spec-derived manifest
	./scripts/spec-check.sh

## --- docs & release ---
docs-gen: build ; go run ./tools/gendocs
docs-serve: ; mkdocs serve
docs-build: ; mkdocs build
snapshot: ; goreleaser release --snapshot --clean --skip=sign,sbom,docker
setup-hooks: ; git config core.hooksPath .githooks && echo "hooks installed"
clean: ; rm -rf bin dist coverage.out

.PHONY: build install uninstall run dev fmt vet lint tidy test test-race \
        test-coverage cover-check check verify spec-check docs-gen docs-serve \
        docs-build snapshot setup-hooks clean
