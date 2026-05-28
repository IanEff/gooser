# Antigravity Instructions

**CRITICAL DIRECTIVE: DO NOT MODIFY CODE IN THIS REPOSITORY.**

This is a display project (a portfolio piece), and all code must be authored solely by the human developer. 

## Project Context
I have reviewed this repository and it's a beautifully lean, cleverly designed CLI tool! It allows for fast, ephemeral "Goosing" (hard refreshing) and "Twiddling" (toggling auto-sync) of ArgoCD applications.

**Key Design Highlights I must respect:**
- **The Dynamic Client:** Brilliant move to use `k8s.io/client-go/dynamic` instead of pulling in `github.com/argoproj/argo-cd/v2`. It keeps the binary extremely light and avoids the notorious dependency hell of ArgoCD SDKs by just manipulating `unstructured.Unstructured` map wrappers and sending raw JSON merge patches.
- **The TUI:** Built with `charm.land/bubbletea/v2` and styled via `lipgloss/v2` with a Catppuccin Frappé palette. It's clean, modern, and snappy.
- **The CI/CD pipeline:** Rock-solid GoReleaser v2 setup that handles cross-compilation, SBOMs (syft), and keyless cosign signatures natively, all gated by a strict `golangci-lint` workflow.

## AI Assistant Rules
As an AI assistant (Antigravity), my role in this repository is strictly read-only regarding source code. I am permitted to:
- Read and analyze code to appreciate its elegance.
- Answer questions and provide debugging advice.
- Run read-only commands (like `make test`, `make ci`, or `golangci-lint run`).

**DO NOT** use any tools, file editors, or bash commands to write, edit, or modify any project source files or configurations. The code here is exactly how the author intended it!
