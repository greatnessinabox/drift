package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/greatnessinabox/drift/internal/config"
)

type RustAnalyzer struct{}

func (r *RustAnalyzer) Language() Language { return LangRust }

func (r *RustAnalyzer) Extensions() []string { return []string{".rs"} }

func (r *RustAnalyzer) FindFiles(root string, exclude []string) ([]string, error) {
	exclude = append(exclude, "target")
	return walkFiles(root, exclude, r.Extensions(), nil)
}

var rsFuncPattern = regexp.MustCompile(`(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?fn\s+(\w+)`)

var rsComplexityPatterns = []complexityPattern{
	{regexp.MustCompile(`\bif\b`), 1},
	{regexp.MustCompile(`\belse\s+if\b`), 1},
	{regexp.MustCompile(`\bfor\b`), 1},
	{regexp.MustCompile(`\bwhile\b`), 1},
	{regexp.MustCompile(`\bloop\b`), 1},
	{regexp.MustCompile(`\bmatch\b`), 1},
	{regexp.MustCompile(`=>\s*\{`), 1},
	{regexp.MustCompile(`&&`), 1},
	{regexp.MustCompile(`\|\|`), 1},
	{regexp.MustCompile(`\?;`), 1},
}

func (r *RustAnalyzer) AnalyzeComplexity(files []string) ([]FunctionComplexity, int) {
	var results []FunctionComplexity
	for _, path := range files {
		funcs := scanHeuristicComplexity(path, rsFuncPattern, 1, rsComplexityPatterns)
		results = append(results, funcs...)
	}
	return results, len(results)
}

func (r *RustAnalyzer) AnalyzeDeps(root string) ([]DepStatus, error) {
	cargoPath := filepath.Join(root, "Cargo.toml")
	f, err := os.Open(cargoPath)
	if err != nil {
		return nil, fmt.Errorf("reading Cargo.toml: %w", err)
	}
	defer f.Close()

	var results []DepStatus
	inDeps := false
	scanner := bufio.NewScanner(f)

	depLineSimple := regexp.MustCompile(`^(\w[\w-]*)\s*=\s*"([^"]+)"`)
	depLineTable := regexp.MustCompile(`^(\w[\w-]*)\s*=\s*\{.*version\s*=\s*"([^"]+)"`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "[dependencies]" || line == "[dev-dependencies]" {
			inDeps = line == "[dependencies]"
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}

		if !inDeps || line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var name, version string
		if m := depLineSimple.FindStringSubmatch(line); m != nil {
			name, version = m[1], m[2]
		} else if m := depLineTable.FindStringSubmatch(line); m != nil {
			name, version = m[1], m[2]
		} else {
			continue
		}

		dep := DepStatus{
			Module:         name,
			CurrentVersion: version,
		}

		latest, err := fetchCratesIOLatest(name)
		if err != nil {
			dep.Status = "unknown"
			dep.LatestVersion = "?"
		} else {
			dep.LatestVersion = latest
			if dep.CurrentVersion == latest {
				dep.Status = "current"
			} else {
				dep.StaleDays = 30
				dep.Status = "stale"
			}
		}

		results = append(results, dep)
	}
	return results, nil
}

type cratesIOResponse struct {
	Crate struct {
		MaxStableVersion string `json:"max_stable_version"`
	} `json:"crate"`
}

func fetchCratesIOLatest(name string) (string, error) {
	var resp cratesIOResponse
	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s", name)
	if err := fetchJSON(url, &resp, "drift/1.0 (https://github.com/greatnessinabox/drift)"); err != nil {
		return "", err
	}
	return resp.Crate.MaxStableVersion, nil
}

var rsImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^use\s+(\S+);`),
	regexp.MustCompile(`^pub\s+use\s+(\S+);`),
	regexp.MustCompile(`^extern\s+crate\s+(\w+);`),
}

func (r *RustAnalyzer) AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation {
	var violations []BoundaryViolation
	for _, path := range files {
		imports := extractImports(path, rsImportPatterns, 1)
		violations = append(violations, checkBoundaryViolations(path, imports, rules, root)...)
	}
	return violations
}

var rsExportPattern = regexp.MustCompile(`^pub\s+(?:async\s+)?fn\s+(\w+)`)

func (r *RustAnalyzer) AnalyzeDeadCode(files []string) []DeadFunction {
	callPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\w+\(`),
	}
	return detectExportsAndCalls(files, rsExportPattern, 1, callPatterns)
}
