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
drift init               # Create .drift.yaml
```

**Testing:**
No automated tests currently exist. Manual validation:
```bash
go build ./cmd/drift/ && ./drift report
```

## Architecture Overview

drift is a real-time Go codebase health analyzer with a Bubble Tea TUI. The architecture follows clean separation of concerns:

### Core Flow
1. **Analyzer** (`internal/analyzer/`) → Parses Go AST and extracts metrics
2. **Health Scorer** (`internal/health/`) → Computes weighted scores (0-100)
3. **History Analyzer** (`internal/history/`) → Walks commit history for sparkline trends
4. **TUI** (`internal/tui/`) → Renders dashboard using Bubble Tea + Lip Gloss
5. **Watcher** (`internal/watcher/`) → Monitors file changes (200ms debounce)
6. **AI** (`internal/ai/`) → Provides diagnostics via Anthropic/OpenAI

### Key Components

**`internal/analyzer/`** — Analysis engine
- `complexity.go`: Cyclomatic complexity via `go/ast` (counts if/for/range/case/select/&&/||)
- `deps.go`: Parses `go.mod`, queries Go module proxy for staleness
- `imports.go`: Validates architectural boundary rules from config
- `deadcode.go`: Detects exported functions with zero callers
- `analyzer.go`: Orchestrates analyses, supports full runs and incremental single-file updates

**`internal/health/`** — Scoring system
- `score.go`: Weighted scoring model (complexity 30%, deps 20%, boundaries 20%, dead code 15%, coverage 15%)
- Tracks previous scores to compute deltas for UI display

**`internal/tui/`** — Terminal UI (Bubble Tea v1)
- Follows Elm Architecture: `Init()` → `Update(msg)` → `View()`
- `app.go`: Root model with 5 panels (score, complexity, deps, boundaries, activity)
- `styles.go`: All Lip Gloss styles with green/yellow/red severity colors
- Supports animated score transitions and AI diagnosis overlay

**`internal/watcher/`** — File system monitoring
- Uses `fsnotify` with 200ms debounce
- Only reacts to `.go` and `go.mod` changes

**`internal/history/`** — Git history analysis
- Uses `go-git` to walk last 10 commits
- Extracts files from each commit into temp directories for analysis
- Generates sparkline data (health score, complexity, violations, dead code)
- Runs on TUI startup only (snapshot-based, not real-time)

**`internal/ai/`** — AI diagnostics
- `provider.go`: Provider interface abstraction
- `anthropic.go` / `openai.go`: Concrete implementations using official SDKs
- `diagnose.go`: Builds prompts from metrics, orchestrates API calls

**`cmd/drift/main.go`** — CLI entry point using Cobra

## Key Conventions

### AST Analysis
- Always use standard library `go/ast` and `go/parser` — no external AST libraries
- When adding new metrics, parse incrementally to support file-watching efficiency

### Bubble Tea Patterns
- Define message types as `FooMsg` structs (e.g., `fileChangedMsg`, `analysisCompleteMsg`)
- Handle in `Update()` method, return `tea.Cmd` to trigger async work
- Keep panels independent — each has its own `viewX()` rendering method
- All styling goes through Lip Gloss styles in `styles.go`

### Configuration
- Config lives in `.drift.yaml` at project root
- Always provide sensible defaults in `config.Defaults()`
- Weights must sum to 1.0 (validated on load)

### Health Scoring
- Each metric has a dedicated `<metric>Score()` method in `health/score.go`
- Scores are 0-100, with higher = healthier
- Total score is weighted average based on config

### AI Provider Pattern
- Implement `ai.Provider` interface for new AI providers
- Register in factory method in `provider.go`
- Document required environment variables (e.g., `ANTHROPIC_API_KEY`)

### Sparkline Rendering
- Use `sparkline(data []float64)` function in `styles.go`
- Uses Unicode chars: ▁▂▃▄▅▆▇█
- Color-coded based on trend (green=up, red=down, yellow=stable)
- Normalizes data to 0-8 range automatically

### History Analysis
- History analyzer runs on TUI startup only (via `loadHistory()` tea.Cmd)
- Creates temp directories to extract files from each commit
- Skips analysis if not a git repo (gracefully returns empty sparkline data)
- Fixed at 10 commits for simplicity (not configurable yet)

## Adding New Features

### Adding a New Metric
1. Create `internal/analyzer/newmetric.go` with analysis logic
2. Add results field to `analyzer.Results` struct
3. Implement `newmetricScore()` in `health/score.go`
4. Add weight field to `config.WeightConfig` with validation
5. Create `viewNewMetric()` panel in `tui/app.go`
6. Wire into dashboard layout in `View()`

### Adding a New AI Provider
1. Create `internal/ai/newprovider.go` implementing `Provider` interface
2. Add to factory in `provider.go`
3. Document environment variable in README

### Modifying TUI Layout
- Update panel constants in `tui/app.go` (e.g., `panelCount`)
- Add navigation logic in `Update()` for `tea.KeyTab` handling
- Define new `viewX()` method for rendering
- Compose in main `View()` using `lipgloss.JoinHorizontal/Vertical`
