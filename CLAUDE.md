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

Single-binary CLI tool (`cmd/gooser/main.go`) with no sub-packages. The entry point:

1. Reads a positional `[appname]` argument from `os.Args` (optional).
2. Loads kubeconfig via `flag.String("kubeconfig", ...)` defaulting to `~/.kube/config`.
3. Builds a `k8s.io/client-go/dynamic` client — intentionally generic, no ArgoCD SDK dependency.
4. Lists all ArgoCD `Application` CRs in the `argocd` namespace via the `argoproj.io/v1alpha1/applications` GVR, printing name, sync status, and health status.
5. If `appname` was provided, patches that application with the `argocd.argoproj.io/refresh: hard` annotation to trigger a hard refresh.

The dynamic client approach (rather than the ArgoCD typed client) keeps the dependency tree light and avoids the transitive conflicts that come with `github.com/argoproj/argo-cd/v2`.

## Linting

golangci-lint v2 config (`.golangci.yml`) enables: `gosec`, `godot`, `misspell`, `errorlint`, `revive`, plus `staticcheck` (all checks). Formatters: `gofmt` + `goimports`.

## Scaffolding

The project was generated from `go-template` (copier). To update the scaffold: re-run copier from `/Users/ian/projects/scaffolding/go-template`. The `.copier-answers.yml` file records the answers used at generation time.
