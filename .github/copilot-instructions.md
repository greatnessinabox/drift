# Copilot Instructions for drift

## Build and Test

**Build:**
```bash
go build ./cmd/drift/
```

**Run the TUI:**
```bash
./drift                  # or: go run ./cmd/drift/
drift report             # Terminal report
drift snapshot           # JSON output for CI
drift check --fail-under 70  # CI health gate
drift fix                # Interactive Copilot-powered fixing
drift init               # Create .drift.yaml
```

**Testing:**
```bash
go test ./internal/analyzer/...   # Unit tests (19 tests)
go test ./...                     # All tests
```

## Architecture Overview

drift is a real-time **multi-language** codebase health analyzer with a Bubble Tea TUI. It supports Go, TypeScript/JavaScript, Python, Rust, and Java through a pluggable `LanguageAnalyzer` interface.

### Core Flow
1. **Language Detection** (`internal/analyzer/language.go`) → Auto-detects from manifest files or reads `language` from config
2. **Analyzer** (`internal/analyzer/`) → Delegates to the right language analyzer via the interface
3. **Health Scorer** (`internal/health/`) → Computes weighted scores (0-100)
4. **History Analyzer** (`internal/history/`) → Walks commit history for sparkline trends
5. **TUI** (`internal/tui/`) → Renders dashboard using Bubble Tea + Lip Gloss
6. **Watcher** (`internal/watcher/`) → Monitors file changes by detected extensions (200ms debounce)
7. **AI** (`internal/ai/`) → Provides language-aware diagnostics via Anthropic/OpenAI

### The LanguageAnalyzer Interface

```go
type LanguageAnalyzer interface {
    Language() Language
    Extensions() []string
    FindFiles(root string, exclude []string) ([]string, error)
    AnalyzeComplexity(files []string) ([]FunctionComplexity, int)
    AnalyzeDeps(root string) ([]DepStatus, error)
    AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation
    AnalyzeDeadCode(files []string) []DeadFunction
}
```

Implementations: `GoAnalyzer`, `TypeScriptAnalyzer`, `PythonAnalyzer`, `RustAnalyzer`, `JavaAnalyzer`

### Key Components

**`internal/analyzer/`** — Multi-language analysis engine
- `language.go`: Language type, interface definition, auto-detection, factory
- `analyzer.go`: Orchestrator — delegates to the detected language analyzer
- `heuristic.go`: Shared utilities for non-Go languages (file walking, regex complexity, import extraction, dead code)
- `go_analyzer.go` + `complexity.go`, `deps.go`, `imports.go`, `deadcode.go`: Go-specific AST analysis
- `typescript_analyzer.go`: TS/JS via regex + npm registry
- `python_analyzer.go`: Python via indentation-aware regex + PyPI
- `rust_analyzer.go`: Rust via regex + crates.io
- `java_analyzer.go`: Java via regex + Maven Central

**`internal/health/`** — Scoring system
- `score.go`: Weighted scoring model (complexity 30%, deps 20%, boundaries 20%, dead code 15%, coverage 15%)
- Tracks previous scores to compute deltas for UI display

**`internal/tui/`** — Terminal UI (Bubble Tea v1)
- Follows Elm Architecture: `Init()` → `Update(msg)` → `View()`
- `app.go`: Root model with 5 panels (score, complexity, deps, boundaries, activity)
- `styles.go`: All Lip Gloss styles with green/yellow/red severity colors
- Displays detected language in header, supports animated score transitions and AI diagnosis overlay

**`internal/watcher/`** — File system monitoring
- Uses `fsnotify` with 200ms debounce
- Watches only files matching the detected language's extensions (passed from analyzer)

**`internal/history/`** — Git history analysis
- Uses `go-git` to walk last 10 commits
- Extracts files by detected language extensions into temp directories for analysis
- Generates sparkline data (health score, complexity, violations, dead code)

**`internal/ai/`** — AI diagnostics
- `provider.go`: Provider interface abstraction
- `anthropic.go` / `openai.go`: Concrete implementations using official SDKs
- `diagnose.go`: Builds language-aware prompts from metrics, includes code snippets

**`cmd/drift/main.go`** — CLI entry point using Cobra

## Key Conventions

### Language Analysis
- Go uses standard library `go/ast` and `go/parser` — no external AST libraries
- All other languages use heuristic regex patterns via shared utilities in `heuristic.go`
- `scanHeuristicComplexity()` handles brace-depth function boundary detection
- Python has custom indentation-based boundary detection in `python_analyzer.go`

### Bubble Tea Patterns
- Define message types as `FooMsg` structs (e.g., `fileChangedMsg`, `analysisCompleteMsg`)
- Handle in `Update()` method, return `tea.Cmd` to trigger async work
- Keep panels independent — each has its own `viewX()` rendering method
- All styling goes through Lip Gloss styles in `styles.go`

### Configuration
- Config lives in `.drift.yaml` at project root
- `language` field: empty = auto-detect, or set explicitly ("go", "typescript", "python", "rust", "java")
- Always provide sensible defaults in `config.Defaults()`
- Weights must sum to 1.0

### Health Scoring
- Each metric has a dedicated `<metric>Score()` method in `health/score.go`
- Scores are 0-100, with higher = healthier
- Total score is weighted average based on config

### Adding a New Language
1. Create `internal/analyzer/<lang>_analyzer.go` implementing `LanguageAnalyzer`
2. Add `Lang<Name>` constant to `language.go`
3. Add manifest file to `DetectLanguage()` checks
4. Add case to `NewLanguageAnalyzer()` factory
5. Test: `go build ./cmd/drift/ && drift report` in a project of that language

### Adding a New Metric
1. Create `internal/analyzer/newmetric.go` with analysis logic
2. Add results field to `analyzer.Results` struct
3. Implement `newmetricScore()` in `health/score.go`
4. Add weight field to `config.WeightConfig` with validation
5. Create `viewNewMetric()` panel in `tui/app.go`
6. Wire into dashboard layout in `View()`
