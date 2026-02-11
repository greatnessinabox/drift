# Contributing to drift

Thanks for your interest in contributing to drift! Here's how to get started.

## Development Setup

```bash
# Clone the repo
git clone https://github.com/greatnessinabox/drift.git
cd drift

# Build
go build ./cmd/drift/

# Run tests
go test ./...

# Run with race detector
go test -race ./...
```

## Project Structure

```
cmd/drift/          CLI entrypoint
internal/
  analyzer/         Language analyzers (Go, TS, Python, Rust, Java)
  ai/               AI-powered diagnostics (Anthropic, OpenAI)
  config/           YAML configuration
  health/           Weighted health scoring
  history/          Git-based trend tracking
  tui/              Bubble Tea live dashboard
  watcher/          File system watcher
```

## Adding a New Language Analyzer

1. Create `internal/analyzer/<lang>_analyzer.go`
2. Implement the `LanguageAnalyzer` interface:
   - `Language()` - return the language constant
   - `Extensions()` - file extensions to scan
   - `FindFiles()` - walk the file tree
   - `AnalyzeComplexity()` - count decision points per function
   - `AnalyzeDeps()` - parse manifest and check registry
   - `AnalyzeImports()` - detect boundary violations
   - `AnalyzeDeadCode()` - find unused exports
3. Register in `language.go` (`NewLanguageAnalyzer` and `DetectLanguage`)
4. Add tests

## Pull Requests

1. Fork the repo and create a feature branch
2. Make your changes
3. Run `go test ./...` and ensure all tests pass
4. Run `go vet ./...` with no warnings
5. Keep commits focused and write clear commit messages
6. Open a PR against `main`

## Reporting Bugs

Open an issue with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- `drift` version and OS

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions focused and small
- Add tests for new functionality
- No external parser dependencies for language analyzers (use heuristic/regex)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
