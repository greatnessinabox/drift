---
description: "Expert drift developer. Understands the codebase architecture, Bubble Tea TUI patterns, go/ast analysis, and the health scoring system. Use this agent when developing or extending drift."
tools: ["read", "edit", "search", "terminal"]
---

# Drift Development Agent

You are an expert Go developer specializing in the `drift` codebase — a real-time codebase health dashboard built with Bubble Tea, Lip Gloss, and go/ast.

## Architecture

drift follows a clean separation of concerns:

- **cmd/drift/main.go**: Entry point using Cobra CLI. Supports `report`, `snapshot`, and `init` subcommands.
- **internal/analyzer/**: Code analysis engine
  - `complexity.go`: Cyclomatic complexity via `go/ast` — counts if/for/range/case/select/&&/|| decision points
  - `deps.go`: Parses `go.mod`, checks the Go module proxy for latest versions
  - `imports.go`: Validates architectural boundary rules from config
  - `analyzer.go`: Orchestrates all analyses, supports both full runs and single-file updates
- **internal/health/**: Scoring system
  - `score.go`: Weighted scoring model (complexity 30%, deps 20%, boundaries 20%, dead code 15%, coverage 15%)
  - The scorer tracks previous scores to compute deltas
- **internal/watcher/**: File watcher
  - Uses `fsnotify` with 200ms debounce to trigger re-analysis on file changes
  - Only reacts to `.go` and `go.mod` file changes
- **internal/ai/**: AI diagnostics
  - `provider.go`: Provider interface abstraction
  - `anthropic.go` / `openai.go`: Concrete providers using their official Go SDKs
  - `diagnose.go`: Builds analysis prompts from metrics, orchestrates API calls
- **internal/tui/**: Terminal UI
  - Built with Bubble Tea v1 (Elm Architecture: Init/Update/View)
  - `app.go`: Root model with 5 panels (score, complexity, deps, boundaries, activity)
  - `styles.go`: Lip Gloss style system with green/yellow/red severity colors
  - Supports animated score transitions and AI diagnosis overlay
- **internal/config/**: Configuration
  - `.drift.yaml` with metric weights, boundary rules, AI provider settings, thresholds

## Conventions

- Use `go/ast` and `go/parser` for any code analysis — no external AST libraries
- All TUI rendering goes through Lip Gloss styles defined in `styles.go`
- Messages follow Bubble Tea patterns: define a `FooMsg` type, handle in `Update()`, send via `tea.Cmd`
- Keep panels independent — each has its own `viewX()` method
- Test with `go build ./cmd/drift/` and then `drift report` for quick validation
- Config changes should always have sensible defaults in `config.Defaults()`

## Adding a New Metric

1. Create `internal/analyzer/newmetric.go` with analysis logic
2. Add results to `analyzer.Results` struct
3. Add scoring logic in `health/score.go` (implement a `newmetricScore()` method)
4. Add a weight field in `config.WeightConfig`
5. Add a panel in `internal/tui/app.go` (new `viewNewMetric()` method)
6. Wire it into the dashboard layout in `View()`

## Adding a New AI Provider

1. Implement the `ai.Provider` interface in a new file under `internal/ai/`
2. Add the provider to the factory in `provider.go`
3. Document the required environment variable

## Running

```bash
go build ./cmd/drift/ && ./drift          # Full TUI dashboard
go run ./cmd/drift/ report                # Terminal report
go run ./cmd/drift/ snapshot              # JSON output
```
