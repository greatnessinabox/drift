---
name: codebase-health-analysis
description: "Analyze codebase health across Go, TypeScript/JS, Python, Rust, Java, Ruby, PHP, and C# — including cyclomatic complexity, dependency freshness, dead code detection, and architectural boundary violations. Use this skill when asked about code quality, tech debt, code health, complexity analysis, or dependency management."
---

# Codebase Health Analysis

When analyzing codebase health, drift auto-detects the language from manifest files and applies the appropriate analysis strategy.

## Language Detection

Check for manifest files in this order:
1. `go.mod` → Go (full AST analysis)
2. `package.json` → TypeScript/JavaScript (heuristic regex)
3. `pyproject.toml` or `requirements.txt` → Python (indentation-aware heuristic)
4. `Cargo.toml` → Rust (heuristic regex)
5. `pom.xml` or `build.gradle` → Java (heuristic regex)
6. `Gemfile` → Ruby (def/end tracking heuristic)
7. `composer.json` → PHP (heuristic regex)
8. `*.csproj` → C# (heuristic regex)

## 1. Cyclomatic Complexity Analysis

Calculate cyclomatic complexity for each function by counting decision points:

**Base complexity:** 1 per function, then +1 for each decision point.

### Go (AST-based)
- +1 for each: `if`, `for`, `range`, `select`, `switch` (type switch)
- +1 for each `case` clause (except default)
- +1 for each `&&` or `||` binary operator
- Do NOT recurse into function literals

### TypeScript/JavaScript (regex-based)
- +1 for: `if(`, `else if`, `for(`, `while(`, `do{`, `case X:`, `catch(`
- +1 for: `&&`, `||`, `??`, `?.`

### Python (indentation-aware)
- +1 for: `if`, `elif`, `for`, `while`, `except`, `with`
- +1 for: `and`, `or`, inline `if...else`
- Function boundaries determined by indentation level

### Rust
- +1 for: `if`, `else if`, `for`, `while`, `loop`, `match`, `=> {`
- +1 for: `&&`, `||`, `?`

### Java
- +1 for: `if(`, `else if`, `for(`, `while(`, `do{`, `case X:`, `catch(`
- +1 for: `&&`, `||`, ternary `? :`

### Ruby (def/end tracking)
- +1 for: `if`, `elsif`, `unless`, `for`, `while`, `until`, `when`, `rescue`
- +1 for: `&&`, `||`
- Function boundaries determined by `def`/`end` keyword depth

### PHP
- +1 for: `if(`, `elseif(`, `for(`, `foreach(`, `while(`, `do{`, `case X:`, `catch(`
- +1 for: `&&`, `||`, ternary `? :`

### C#
- +1 for: `if(`, `else if`, `for(`, `foreach(`, `while(`, `do{`, `case X:`, `catch(`, `switch(`
- +1 for: `&&`, `||`, `??`, ternary `? :`

**Severity thresholds:**
- 1-10: Good (green) — easy to understand and test
- 11-20: Warning (yellow) — consider refactoring
- 21+: Critical (red) — should be refactored immediately

## 2. Dependency Freshness

Check the language-specific manifest against its registry:

| Language | Manifest | Registry |
|----------|----------|----------|
| Go | `go.mod` | `https://proxy.golang.org/{mod}/@latest` |
| TypeScript/JS | `package.json` | `https://registry.npmjs.org/{pkg}/latest` |
| Python | `requirements.txt` / `pyproject.toml` | `https://pypi.org/pypi/{pkg}/json` |
| Rust | `Cargo.toml` | `https://crates.io/api/v1/crates/{crate}` |
| Java | `pom.xml` / `build.gradle` | `https://search.maven.org/solrsearch/select` |
| Ruby | `Gemfile` | `https://rubygems.org/api/v1/gems/{name}.json` |
| PHP | `composer.json` | `https://repo.packagist.org/p2/{vendor}/{package}.json` |
| C# | `*.csproj` | `https://api.nuget.org/v3-flatcontainer/{name}/index.json` |

**Staleness classification:**
- Current: version matches latest
- Stale (30-90 days): behind latest
- Outdated (90+ days): significantly behind, may have security patches

## 3. Architectural Boundary Violations

Define import rules that enforce module boundaries. A boundary rule like `pkg/api -> internal/db` means code in `pkg/api/` should NOT import packages containing `internal/db`.

Import detection patterns per language:
- **Go**: Full AST import parsing
- **TypeScript/JS**: `import...from`, `require()`, dynamic `import()`
- **Python**: `import X`, `from X import`
- **Rust**: `use`, `pub use`, `extern crate`
- **Java**: `import` (including static)
- **Ruby**: `require`, `require_relative`
- **PHP**: `use`, `require_once`, `include`
- **C#**: `using`, `using static`

## 4. Dead Code Detection

Find exported/public functions that have zero callers within the project:

- **Go**: Full AST — tracks all `CallExpr` and `SelectorExpr` nodes
- **Other languages**: Scans for export patterns and cross-references with all call sites

## 5. Health Score Calculation

Combine metrics into an overall health score (0-100):

```
total = complexity_score * 0.30
      + deps_score       * 0.20
      + boundaries_score * 0.20
      + dead_code_score  * 0.15
      + coverage_score   * 0.15
```

## Example Usage

```bash
# Install drift
go install github.com/greatnessinabox/drift/cmd/drift@latest

# Run in any supported project
cd /path/to/project
drift report          # Terminal-formatted report
drift snapshot        # JSON output for CI (includes "language" field)
drift check --fail-under 70  # CI health gate
drift                 # Full interactive TUI dashboard
```
