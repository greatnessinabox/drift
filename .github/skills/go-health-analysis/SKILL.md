---
name: go-health-analysis
description: "Analyze Go codebase health including cyclomatic complexity, dependency freshness, dead code detection, and architectural boundary violations. Use this skill when asked about code quality, tech debt, code health, complexity analysis, or dependency management in Go projects."
---

# Go Codebase Health Analysis

When analyzing Go codebase health, follow this systematic approach:

## 1. Cyclomatic Complexity Analysis

Calculate cyclomatic complexity for each function by counting decision points in the AST:

- Start with a base complexity of 1
- +1 for each: `if`, `for`, `range`, `select`, `switch` (type switch)
- +1 for each `case` clause (except the default case)
- +1 for each `&&` or `||` binary operator
- Do NOT recurse into function literals (closures count separately)

**Severity thresholds:**
- 1-10: Good (green) — function is easy to understand and test
- 11-20: Warning (yellow) — consider refactoring, may be hard to test thoroughly
- 21+: Critical (red) — should be refactored immediately, high bug risk

**Common refactoring strategies for high complexity:**
- Extract nested conditions into helper functions
- Replace switch/case with a lookup map or table-driven approach
- Split large functions along logical boundaries (validation, processing, response)
- Use early returns to reduce nesting depth

## 2. Dependency Freshness

Check `go.mod` for direct dependencies and compare versions against the Go module proxy:

```
GET https://proxy.golang.org/{module}/@latest
```

**Staleness classification:**
- Current: version matches latest
- Stale (30-90 days): the latest version was released 30-90 days ago and you're not on it
- Outdated (90+ days): significantly behind, may have security patches

**Key concerns for outdated deps:**
- Check the Go vulnerability database for known CVEs
- Review changelogs for breaking changes before updating
- Update one dependency at a time, run tests between updates

## 3. Architectural Boundary Violations

Define import rules that enforce module boundaries. A boundary rule like `pkg/api -> internal/db` means code in `pkg/api/` should NOT import packages containing `internal/db`.

To check violations:
1. Parse all `.go` files in the project
2. For each file, determine its directory relative to the project root
3. For each import in the file, check against all boundary rules
4. A violation occurs when the file's directory matches a rule's "from" pattern AND its import path matches the "to" pattern

**Common boundary patterns for Go projects:**
- `cmd/ -> internal/` (only through defined interfaces)
- `pkg/api/ -> internal/db/` (API should use repository interfaces)
- `internal/service/ -> internal/handler/` (services shouldn't know about HTTP)

## 4. Dead Code Detection

Find exported functions that have zero callers within the project:

1. Build a complete map of all function declarations across all files
2. Walk all AST nodes to find function call expressions
3. Track which functions are called at least once
4. Exported functions with zero callers are candidates for dead code
5. Exclude `main()`, `init()`, test functions, and interface implementations

**Note:** False positives can occur for functions used via reflection or called from external packages.

## 5. Health Score Calculation

Combine metrics into an overall health score (0-100):

```
total = complexity_score * 0.30
      + deps_score       * 0.20
      + boundaries_score * 0.20
      + dead_code_score  * 0.15
      + coverage_score   * 0.15
```

Each component score ranges from 0-100:
- Complexity: 100 minus penalties for functions exceeding the threshold
- Dependencies: 100 minus penalties based on staleness ratio
- Boundaries: 100 minus 10 points per violation
- Dead code: 100 minus penalties per dead exported function
- Coverage: directly from test coverage percentage

## Example Usage

To analyze a Go project at `/path/to/project`:

```bash
# Install drift
go install github.com/greatnessinabox/drift@latest

# Run in the project directory
cd /path/to/project
drift report          # Terminal-formatted report
drift snapshot        # JSON output for CI
drift                 # Full interactive TUI dashboard
```
