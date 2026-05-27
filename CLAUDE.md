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
make clean        # removes bin/ and coverage.out
```

Run a single test:
```sh
go test ./... -run TestFunctionName
```

## Architecture

CLI tool (`cmd/gooser/main.go`) backed by two internal packages:

- **`internal/gooser`** — Kubernetes dynamic client wrapper; `Client` exposes `List`, `Goose`, and `Twiddle`.
- **`internal/tui`** — Bubbletea v2 TUI (`charmed.go` model, `theme.go` Catppuccin Frappé styles).

Entry point flow:

1. Reads a positional `[appname]` argument from `os.Args` (optional). Special values `version`, `--version`, `-v` print the build stamp and exit.
2. Loads kubeconfig via `flag.String("kubeconfig", ...)` defaulting to `~/.kube/config`.
3. Builds a `k8s.io/client-go/dynamic` client — intentionally generic, no ArgoCD SDK dependency.
4. Lists all ArgoCD `Application` CRs in the `argocd` namespace via the `argoproj.io/v1alpha1/applications` GVR.
5. If no `appname` was provided, runs the TUI so the user can pick an app and choose an action (`goose` or `twiddle`). If `appname` was provided, defaults to `goose`.
6. Executes the chosen action:
   - **Goose** — patches the app with `argocd.argoproj.io/refresh: hard` to trigger a hard refresh.
   - **Twiddle** — reads the current `spec.syncPolicy.automated` field and toggles it on (selfHeal + prune) or off (null-patches the field away).

The dynamic client approach (rather than the ArgoCD typed client) keeps the dependency tree light and avoids the transitive conflicts that come with `github.com/argoproj/argo-cd/v2`.

## TUI key bindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `enter` / `g` | Goose selected app (hard refresh) |
| `t` | Twiddle auto-sync for selected app |
| `q` / `ctrl+c` | Quit |

## Releases

Releases are automated via GoReleaser v2 (`.goreleaser.yaml`). Artifacts: `tar.gz` archives for linux/darwin × amd64/arm64, SHA-256 checksums, SBOM (syft), and cosign keyless signatures. The GitHub Actions release workflow gates on the full `make ci` pipeline before publishing.

```sh
goreleaser release --snapshot --clean   # local dry run
goreleaser check                        # validate config
```

Version, commit, and date are stamped into `main.version`, `main.commit`, `main.date` via `-ldflags` by both the Makefile and GoReleaser.

## Linting

golangci-lint v2 config (`.golangci.yml`) enables: `gosec`, `godot`, `misspell`, `errorlint`, `revive`, plus `staticcheck` (all checks). Formatters: `gofmt` + `goimports`.

## Scaffolding

The project was generated from `go-template` (copier). To update the scaffold: re-run copier from `/Users/ian/projects/scaffolding/go-template`. The `.copier-answers.yml` file records the answers used at generation time.
