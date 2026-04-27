# =============================================================================
# AngelaMos | 2026
# Justfile
# =============================================================================
# jsonpls — CSV / XLSX -> AI-friendly JSON
# =============================================================================

set export
set shell := ["bash", "-uc"]

project    := file_name(justfile_directory())
version    := `git describe --tags --always 2>/dev/null || echo "dev"`
ldflags    := "-s -w -X github.com/CarterPerez-dev/jsonpls/internal/app.version=" + version

# =============================================================================
# Default
# =============================================================================

default:
    @just --list --unsorted

# =============================================================================
# Linting and Formatting
# =============================================================================

[group('lint')]
lint *ARGS:
    golangci-lint run --timeout=5m {{ARGS}}

[group('lint')]
lint-fix:
    golangci-lint run --timeout=5m --fix

[group('lint')]
fmt:
    gofumpt -w .
    golines -w --max-len=100 .

[group('lint')]
tidy:
    go mod tidy

[group('lint')]
vet:
    go vet ./...

# =============================================================================
# Testing
# =============================================================================

[group('test')]
test *ARGS:
    go test -race ./... {{ARGS}}

[group('test')]
test-v *ARGS:
    go test -race -v ./... {{ARGS}}

[group('test')]
cover:
    go test -race -cover ./...

[group('test')]
cover-html:
    go test -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Wrote coverage.html"

# =============================================================================
# CI / Quality
# =============================================================================

[group('ci')]
ci: lint test
    @echo "All checks passed."

[group('ci')]
check: vet test
    @echo "Quick check passed."

# =============================================================================
# Development
# =============================================================================

[group('dev')]
run *ARGS:
    go run . {{ARGS}}

[group('dev')]
dev FILE:
    go run . {{FILE}} --stdout

[group('dev')]
dev-jsonl FILE:
    go run . {{FILE}} --jsonl --stdout

[group('dev')]
dev-raw FILE:
    go run . {{FILE}} --raw --stdout

# =============================================================================
# Build
# =============================================================================

[group('build')]
build:
    go build -ldflags="{{ldflags}}" -o bin/jsonpls .
    @echo "Built: bin/jsonpls ($(du -h bin/jsonpls | cut -f1))"

[group('build')]
build-debug:
    go build -o bin/jsonpls .

[group('build')]
install:
    go install -ldflags="{{ldflags}}" .

[group('build')]
build-all:
    @mkdir -p bin
    GOOS=linux  GOARCH=amd64 go build -ldflags="{{ldflags}}" -o bin/jsonpls_linux_amd64  .
    GOOS=linux  GOARCH=arm64 go build -ldflags="{{ldflags}}" -o bin/jsonpls_linux_arm64  .
    GOOS=darwin GOARCH=amd64 go build -ldflags="{{ldflags}}" -o bin/jsonpls_darwin_amd64 .
    GOOS=darwin GOARCH=arm64 go build -ldflags="{{ldflags}}" -o bin/jsonpls_darwin_arm64 .
    @ls -lh bin/

# =============================================================================
# Release
# =============================================================================
# `just release v0.2.0`           cut a release at the version you give
# `just release-patch`            auto-bump the patch number  (0.1.2 -> 0.1.3)
# `just release-minor`            auto-bump the minor number  (0.1.2 -> 0.2.0)
# `just release-major`            auto-bump the major number  (0.1.2 -> 1.0.0)
# Each pushes the tag, which triggers the GitHub Actions release workflow.
# =============================================================================

[group('release')]
release VERSION:
    #!/usr/bin/env bash
    set -euo pipefail
    VERSION="{{VERSION}}"
    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]]; then
        echo "error: version must look like v1.2.3 (got: $VERSION)" >&2
        exit 1
    fi
    if [[ -n "$(git status --porcelain)" ]]; then
        echo "error: working tree is dirty — commit or stash first" >&2
        git status --short >&2
        exit 1
    fi
    if git rev-parse "$VERSION" >/dev/null 2>&1; then
        echo "error: tag $VERSION already exists" >&2
        exit 1
    fi
    BRANCH=$(git rev-parse --abbrev-ref HEAD)
    if [[ "$BRANCH" != "main" ]]; then
        echo "warning: you are on branch '$BRANCH', not 'main'"
        read -rp "continue anyway? [y/N] " ok
        [[ "$ok" =~ ^[Yy]$ ]] || exit 1
    fi
    echo ""
    echo "  releasing $VERSION from $BRANCH"
    echo ""
    git push origin "$BRANCH"
    git tag -a "$VERSION" -m "$VERSION"
    git push origin "$VERSION"
    echo ""
    echo "  tag pushed. workflow now building binaries..."
    if command -v gh &>/dev/null; then
        sleep 3
        RUN_ID=$(gh run list --workflow=release.yml --limit=1 --json databaseId -q '.[0].databaseId')
        gh run watch "$RUN_ID" --exit-status
        echo ""
        gh release view "$VERSION" 2>/dev/null | head -10 || true
    else
        echo "  install 'gh' to auto-watch the workflow."
        echo "  or check: https://github.com/CarterPerez-dev/jsonpls/actions"
    fi

[group('release')]
release-patch:
    @just release "$(just _next-version patch)"

[group('release')]
release-minor:
    @just release "$(just _next-version minor)"

[group('release')]
release-major:
    @just release "$(just _next-version major)"

[group('release')]
release-dry:
    @echo "would release: $(just _next-version patch)"
    @git log --oneline "$(git describe --tags --abbrev=0 2>/dev/null || git rev-list --max-parents=0 HEAD)..HEAD"

# Internal helper — compute the next semver from the latest tag.
_next-version BUMP:
    #!/usr/bin/env bash
    set -euo pipefail
    LAST=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    SEMVER=${LAST#v}
    IFS=. read -r MAJ MIN PAT <<< "$SEMVER"
    case "{{BUMP}}" in
        major) echo "v$((MAJ+1)).0.0" ;;
        minor) echo "v${MAJ}.$((MIN+1)).0" ;;
        patch) echo "v${MAJ}.${MIN}.$((PAT+1))" ;;
        *) echo "error: unknown bump {{BUMP}}" >&2; exit 1 ;;
    esac

# =============================================================================
# Utilities
# =============================================================================

[group('util')]
info:
    @echo "Project:  {{project}}"
    @echo "Version:  {{version}}"
    @echo "Go:       $(go version | cut -d' ' -f3)"
    @echo "OS:       {{os()}} ({{arch()}})"
    @echo "Module:   $(grep '^module ' go.mod | cut -d' ' -f2)"

[group('util')]
clean:
    -rm -rf bin/ coverage.out coverage.html
    @echo "Cleaned build artifacts."

[group('util')]
deps:
    @echo "Required tools (install if missing):"
    @echo "  go               https://go.dev/dl/"
    @echo "  just             https://github.com/casey/just"
    @echo "  golangci-lint    https://golangci-lint.run/welcome/install/"
    @echo "  gofumpt          go install mvdan.cc/gofumpt@latest"
    @echo "  golines          go install github.com/segmentio/golines@latest"
    @echo "  gh               https://cli.github.com/  (for `just release`)"
    @echo "  goreleaser       https://goreleaser.com/install/  (optional, only for local release dry-runs)"
