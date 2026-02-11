package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/greatnessinabox/drift/internal/config"
)

type TypeScriptAnalyzer struct{}

func (t *TypeScriptAnalyzer) Language() Language { return LangTypeScript }

func (t *TypeScriptAnalyzer) Extensions() []string {
	return []string{".ts", ".tsx", ".js", ".jsx"}
}

func (t *TypeScriptAnalyzer) FindFiles(root string, exclude []string) ([]string, error) {
	exclude = append(exclude, "node_modules", "dist", "build", ".next", "coverage")
	skip := []string{".test.", ".spec.", "__tests__", "__mocks__", ".d.ts"}
	return walkFiles(root, exclude, t.Extensions(), skip)
}

var tsFuncPattern = regexp.MustCompile(
	`(?:^|\s)(?:export\s+)?(?:async\s+)?(?:function\s+(\w+)|(\w+)\s*(?::\s*\w+)?\s*=\s*(?:async\s*)?\(|(\w+)\s*\([^)]*\)\s*(?::\s*\w+)?\s*\{)`,
)

var tsComplexityPatterns = []complexityPattern{
	{regexp.MustCompile(`\bif\s*\(`), 1},
	{regexp.MustCompile(`\belse\s+if\b`), 1},
	{regexp.MustCompile(`\bfor\s*\(`), 1},
	{regexp.MustCompile(`\bwhile\s*\(`), 1},
	{regexp.MustCompile(`\bdo\s*\{`), 1},
	{regexp.MustCompile(`\bcase\s+[^:]+:`), 1},
	{regexp.MustCompile(`\bcatch\s*\(`), 1},
	{regexp.MustCompile(`&&`), 1},
	{regexp.MustCompile(`\|\|`), 1},
	{regexp.MustCompile(`\?\?`), 1},
	{regexp.MustCompile(`\?\.\w`), 1},
}

func (t *TypeScriptAnalyzer) AnalyzeComplexity(files []string) ([]FunctionComplexity, int) {
	var results []FunctionComplexity
	for _, path := range files {
		funcs := scanHeuristicComplexity(path, tsFuncPattern, 1, tsComplexityPatterns)
		// Fix name extraction: try groups 1, 2, 3
		for i := range funcs {
			if funcs[i].Name == "" || funcs[i].Name == "anonymous" {
				// Re-parse to try other groups
				content, err := os.ReadFile(path)
				if err == nil {
					lines := strings.Split(string(content), "\n")
					if funcs[i].Line-1 < len(lines) {
						line := lines[funcs[i].Line-1]
						if m := tsFuncPattern.FindStringSubmatch(line); m != nil {
							for _, g := range m[1:] {
								if g != "" {
									funcs[i].Name = g
									break
								}
							}
						}
					}
				}
			}
		}
		results = append(results, funcs...)
	}
	return results, len(results)
}

type packageJSON struct {
	Dependencies map[string]string `json:"dependencies"`
}

type npmPackageInfo struct {
	Version string `json:"version"`
	Time    string `json:"time"`
}

func (t *TypeScriptAnalyzer) AnalyzeDeps(root string) ([]DepStatus, error) {
	pkgPath := filepath.Join(root, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("reading package.json: %w", err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parsing package.json: %w", err)
	}

	var results []DepStatus
	for name, version := range pkg.Dependencies {
		dep := DepStatus{
			Module:         name,
			CurrentVersion: cleanVersion(version),
		}

		var info npmPackageInfo
		url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", name)
		if err := fetchJSON(url, &info, ""); err != nil {
			dep.Status = "unknown"
			dep.LatestVersion = "?"
		} else {
			dep.LatestVersion = info.Version
			if dep.CurrentVersion == info.Version {
				dep.Status = "current"
			} else {
				dep.StaleDays = estimateStaleDays(name, info.Version)
				if dep.StaleDays > 90 {
					dep.Status = "outdated"
				} else {
					dep.Status = "stale"
				}
			}
		}
		results = append(results, dep)
	}
	return results, nil
}

var tsImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`import\s+.*\s+from\s+['"]([^'"]+)['"]`),
	regexp.MustCompile(`import\s+['"]([^'"]+)['"]`),
	regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`),
}

func (t *TypeScriptAnalyzer) AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation {
	var violations []BoundaryViolation
	for _, path := range files {
		imports := extractImports(path, tsImportPatterns, 1)
		violations = append(violations, checkBoundaryViolations(path, imports, rules, root)...)
	}
	return violations
}

var tsExportPattern = regexp.MustCompile(`export\s+(?:async\s+)?(?:function|const|let|var|class)\s+(\w+)`)

func (t *TypeScriptAnalyzer) AnalyzeDeadCode(files []string) []DeadFunction {
	callPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\w+\(`),
	}
	return detectExportsAndCalls(files, tsExportPattern, 1, callPatterns)
}

func cleanVersion(v string) string {
	v = strings.TrimPrefix(v, "^")
	v = strings.TrimPrefix(v, "~")
	v = strings.TrimPrefix(v, ">=")
	v = strings.TrimPrefix(v, "<=")
	v = strings.TrimPrefix(v, ">")
	v = strings.TrimPrefix(v, "<")
	v = strings.TrimPrefix(v, "=")
	return strings.TrimSpace(v)
}

func estimateStaleDays(pkg, latestVersion string) int {
	type npmFullInfo struct {
		Time map[string]string `json:"time"`
	}
	var info npmFullInfo
	url := fmt.Sprintf("https://registry.npmjs.org/%s", pkg)
	if err := fetchJSON(url, &info, ""); err != nil {
		return 30
	}
	if timeStr, ok := info.Time[latestVersion]; ok {
		t, err := time.Parse(time.RFC3339, timeStr)
		if err == nil {
			return int(time.Since(t).Hours() / 24)
		}
	}
	return 30
}
