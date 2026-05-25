# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
make run          # go run ./cmd/gooser
make build        # produces bin/gooser (injects version/commit/date via ldflags)
make test         # go test ./...
make race         # go test -race ./...
make coverage     # coverage report
make lint         # golangci-lint run
make fmt          # gofmt -w
make fmt-check    # fails if any files are unformatted
make vet          # go vet ./...
make ci           # fmt-check → vet → lint → test → build (full pipeline, matches GitHub Actions)
make tidy         # go mod tidy
```

Run a single test:
```sh
go test ./... -run TestFunctionName
```

## Architecture

This is a single-binary CLI tool (`cmd/gooser/main.go`) with no sub-packages yet. The entry point uses `client-go` to load kubeconfig (`~/.kube/config` by default, overridable with `-kubeconfig`) and then connects to the ArgoCD API via the ArgoCD typed client (`github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned`).

Current functionality:
- **List applications**: queries the `argocd` namespace and prints each app's name, health status, and sync status.
- **`patchApp`** (unexported, not yet wired to CLI): triggers a hard refresh on an ArgoCD app by patching the `argocd.argoproj.io/refresh: hard` annotation via a merge-patch.

## Linting

golangci-lint v2 config (`.golangci.yml`) enables: `gosec`, `godot`, `misspell`, `errorlint`, `revive`, plus `staticcheck` (all checks). Formatters: `gofmt` + `goimports`.

## Known Issues

`go_vulncheck` and `govulncheck ./...` fail to load packages due to a transitive dependency conflict: `argoproj/gitops-engine` mixes `sigs.k8s.io/structured-merge-diff` v4 and v6 types, and `k8s.io/kubernetes v1.31.0` references removed `v1alpha1` API packages. The main binary itself (`./cmd/gooser`) builds cleanly despite this.

## Scaffolding

The project was generated from `go-template` (copier). To update the scaffold: re-run copier from `/Users/ian/projects/scaffolding/go-template`. The `.copier-answers.yml` file records the answers used at generation time.
