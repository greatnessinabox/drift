# drift

**Real-time codebase health dashboard with AI diagnostics.**

drift watches your Go codebase in real-time, detects code health degradation, and uses AI to diagnose problems and suggest fixes. Think Datadog for your codebase, in your terminal.

```
┌─────────────────────── DRIFT ────────────────────────┐
│                                                       │
│            ████████████░░░░░░  87/100                 │
│            ▲ +3 from last commit                      │
│                                                       │
├────────────────────────┬──────────────────────────────┤
│  COMPLEXITY            │  DEPENDENCIES                │
│  ──────────            │  ────────────                 │
│  ⚠ parse.go:47   32   │  ✓ bubbletea    v1.3  current│
│  ⚠ analyze.go:12 28   │  ⚠ cobra        v1.8  34d   │
│    main.go:8     12   │  ✗ go-git       v5.8  180d  │
├────────────────────────┼──────────────────────────────┤
│  BOUNDARIES            │  ACTIVITY                    │
│  ──────────            │  ────────                    │
│  ✗ api → db (3 hits)   │  14:32:05  parse.go modified │
│  ✓ cmd → pkg           │  14:31:52  go.mod updated    │
├────────────────────────┴──────────────────────────────┤
│ [tab] navigate  [d] diagnose  [r] refresh  [q] quit  │
└───────────────────────────────────────────────────────┘
```

## Features

- **Live Dashboard** — Full-screen TUI that updates in real-time as you edit code
- **Sparkline Trends** — Visualize health metrics over the last 10 commits with inline charts
- **Cyclomatic Complexity** — Identifies your most complex functions using Go AST analysis
- **Dependency Freshness** — Checks every dependency against the Go module proxy
- **Architecture Boundaries** — Define import rules and catch violations instantly
- **Dead Code Detection** — Finds exported functions with zero callers
- **AI Diagnostics** — Press `d` to get AI-powered analysis via Claude or GPT-4o
- **Health Score** — Weighted 0-100 score with animated transitions
- **CI-Friendly** — `drift snapshot` outputs JSON for pipeline integration

## Install

```bash
go install github.com/greatnessinabox/drift@latest
```

Or build from source:

```bash
git clone https://github.com/greatnessinabox/drift.git
cd drift
go build ./cmd/drift/
```

## Quick Start

```bash
# Run the live dashboard in any Go project
cd your-go-project
drift

# Generate a terminal-formatted report
drift report

# Output JSON for CI pipelines
drift snapshot

# Check health score and fail if below threshold (for CI)
drift check --fail-under 70

# Create a config file
drift init
```

## Configuration

Create a `.drift.yaml` in your project root:

```yaml
# Directories to exclude
exclude:
  - vendor
  - .git

# Metric weights (must sum to 1.0)
weights:
  complexity: 0.30
  deps: 0.20
  boundaries: 0.20
  dead_code: 0.15
  coverage: 0.15

# Architecture boundary rules
boundaries:
  - deny: "pkg/api -> internal/db"
  - deny: "cmd -> internal/tui"

# AI diagnostics (optional)
ai:
  provider: anthropic  # or "openai"
  model: ""            # uses sensible defaults

# Thresholds
thresholds:
  max_complexity: 15
  max_stale_days: 90
  min_score: 70
```

## AI Diagnostics

Press `d` in the dashboard to trigger an AI diagnosis. Supports:

- **Anthropic Claude** — Set `ANTHROPIC_API_KEY` env var
- **OpenAI GPT-4o** — Set `OPENAI_API_KEY` env var

Configure your provider in `.drift.yaml` under `ai.provider`.

The AI analyzes your worst-scoring metrics and provides specific, actionable recommendations with file and function names.

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `tab` | Navigate between panels |
| `shift+tab` | Navigate backwards |
| `d` | Run AI diagnosis |
| `r` | Force full re-analysis |
| `q` / `ctrl+c` | Quit |
| `esc` | Close diagnosis overlay |

## How It Works

1. **Analysis Engine** — Parses all `.go` files using `go/ast` to calculate cyclomatic complexity, detect dead code, and map import graphs
2. **Dependency Checker** — Reads `go.mod` and queries the Go module proxy for latest versions
3. **File Watcher** — Uses `fsnotify` with 200ms debounce for instant feedback on file changes
4. **History Analyzer** — Uses `go-git` to walk commit history and generate sparkline trends
5. **Health Score** — Weighted average of all metrics, with configurable thresholds
6. **TUI** — Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) for a beautiful terminal experience

## Copilot CLI Integration

drift ships with a custom GitHub Copilot agent profile and an agent skill:

- **`.github/agents/drift-dev.agent.md`** — Turns Copilot CLI into a drift development expert
- **`.github/skills/go-health-analysis/`** — Reusable skill for Go code health analysis

These work with GitHub Copilot CLI, VS Code Copilot, and the Copilot coding agent.

## CI Integration

Use the `drift check` command with `--fail-under` flag for easy CI integration:

```yaml
# GitHub Actions example
- name: Check codebase health
  run: |
    go install github.com/greatnessinabox/drift@latest
    drift check --fail-under 70
```

Or use the `snapshot` command for more advanced workflows:

```yaml
# Advanced CI integration with JSON output
- name: Check codebase health
  run: |
    go install github.com/greatnessinabox/drift@latest
    SCORE=$(drift snapshot | jq '.score.total')
    if (( $(echo "$SCORE < 70" | bc -l) )); then
      echo "Health score $SCORE is below threshold"
      exit 1
    fi
```

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [fsnotify](https://github.com/fsnotify/fsnotify) — File system notifications
- [go-git](https://github.com/go-git/go-git) — Git repository access for history trends
- [Anthropic SDK](https://github.com/anthropics/anthropic-sdk-go) — Claude AI integration
- [OpenAI SDK](https://github.com/openai/openai-go) — GPT integration

## License

MIT
