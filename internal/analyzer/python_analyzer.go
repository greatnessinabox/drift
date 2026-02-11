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

type PythonAnalyzer struct{}

func (p *PythonAnalyzer) Language() Language { return LangPython }

func (p *PythonAnalyzer) Extensions() []string { return []string{".py"} }

func (p *PythonAnalyzer) FindFiles(root string, exclude []string) ([]string, error) {
	exclude = append(exclude, "__pycache__", ".venv", "venv", "env", ".tox", ".eggs", ".mypy_cache")
	skip := []string{"test_", "_test.py", "conftest.py"}
	return walkFiles(root, exclude, p.Extensions(), skip)
}

var pyFuncPattern = regexp.MustCompile(`^(\s*)(?:async\s+)?def\s+(\w+)\s*\(`)

var pyComplexityPatterns = []complexityPattern{
	{regexp.MustCompile(`^\s*if\b`), 1},
	{regexp.MustCompile(`^\s*elif\b`), 1},
	{regexp.MustCompile(`^\s*for\b`), 1},
	{regexp.MustCompile(`^\s*while\b`), 1},
	{regexp.MustCompile(`^\s*except\b`), 1},
	{regexp.MustCompile(`^\s*with\b`), 1},
	{regexp.MustCompile(`\band\b`), 1},
	{regexp.MustCompile(`\bor\b`), 1},
	{regexp.MustCompile(`\bif\b.+\belse\b`), 1},
}

func (p *PythonAnalyzer) AnalyzeComplexity(files []string) ([]FunctionComplexity, int) {
	var results []FunctionComplexity
	for _, path := range files {
		funcs := analyzePythonComplexity(path)
		results = append(results, funcs...)
	}
	return results, len(results)
}

func analyzePythonComplexity(path string) []FunctionComplexity {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var results []FunctionComplexity

	for i, line := range lines {
		matches := pyFuncPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		indent := len(matches[1])
		name := matches[2]

		if strings.HasPrefix(name, "_") && !strings.HasPrefix(name, "__") {
			continue
		}

		complexity := 1
		for j := i + 1; j < len(lines); j++ {
			bodyLine := lines[j]
			if strings.TrimSpace(bodyLine) == "" {
				continue
			}

			bodyIndent := len(bodyLine) - len(strings.TrimLeft(bodyLine, " \t"))
			if bodyIndent <= indent && strings.TrimSpace(bodyLine) != "" {
				break
			}

			for _, p := range pyComplexityPatterns {
				if p.pattern.MatchString(bodyLine) {
					complexity += p.weight
				}
			}
		}

		results = append(results, FunctionComplexity{
			File:       filepath.Base(path),
			Name:       name,
			Line:       i + 1,
			Complexity: complexity,
		})
	}

	return results
}

func (p *PythonAnalyzer) AnalyzeDeps(root string) ([]DepStatus, error) {
	// Try requirements.txt first
	reqPath := filepath.Join(root, "requirements.txt")
	if _, err := os.Stat(reqPath); err == nil {
		return parsePythonRequirements(reqPath)
	}

	// Try pyproject.toml
	pyprojectPath := filepath.Join(root, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); err == nil {
		return parsePyproject(pyprojectPath)
	}

	return nil, fmt.Errorf("no Python dependency file found")
}

func parsePythonRequirements(path string) ([]DepStatus, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var results []DepStatus
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}

		var name, version string
		for _, sep := range []string{"==", ">=", "<=", "~=", "!="} {
			if idx := strings.Index(line, sep); idx > 0 {
				name = strings.TrimSpace(line[:idx])
				version = strings.TrimSpace(line[idx+len(sep):])
				break
			}
		}
		if name == "" {
			name = line
		}

		dep := DepStatus{
			Module:         name,
			CurrentVersion: version,
		}

		latest, err := fetchPyPILatest(name)
		if err != nil {
			dep.Status = "unknown"
			dep.LatestVersion = "?"
		} else {
			dep.LatestVersion = latest
			if dep.CurrentVersion == latest || dep.CurrentVersion == "" {
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

func parsePyproject(path string) ([]DepStatus, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var results []DepStatus
	inDeps := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "dependencies = [" || line == "[tool.poetry.dependencies]" {
			inDeps = true
			continue
		}
		if inDeps && (strings.HasPrefix(line, "[") || line == "]") {
			inDeps = false
			continue
		}

		if !inDeps {
			continue
		}

		line = strings.Trim(line, `"',`)
		if line == "" {
			continue
		}

		name := line
		for _, sep := range []string{">=", "==", "~=", "^"} {
			if idx := strings.Index(line, sep); idx > 0 {
				name = strings.TrimSpace(line[:idx])
				break
			}
		}

		if name == "python" {
			continue
		}

		dep := DepStatus{Module: name, CurrentVersion: ""}
		latest, err := fetchPyPILatest(name)
		if err != nil {
			dep.Status = "unknown"
			dep.LatestVersion = "?"
		} else {
			dep.LatestVersion = latest
			dep.Status = "current"
		}
		results = append(results, dep)
	}
	return results, nil
}

type pypiInfo struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
}

func fetchPyPILatest(pkg string) (string, error) {
	var info pypiInfo
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", pkg)
	if err := fetchJSON(url, &info, ""); err != nil {
		return "", err
	}
	return info.Info.Version, nil
}

var pyImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^import\s+(\S+)`),
	regexp.MustCompile(`^from\s+(\S+)\s+import`),
}

func (p *PythonAnalyzer) AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation {
	var violations []BoundaryViolation
	for _, path := range files {
		imports := extractImports(path, pyImportPatterns, 1)
		violations = append(violations, checkBoundaryViolations(path, imports, rules, root)...)
	}
	return violations
}

var pyExportPattern = regexp.MustCompile(`^def\s+(\w+)\s*\(`)

func (p *PythonAnalyzer) AnalyzeDeadCode(files []string) []DeadFunction {
	callPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\w+\(`),
	}
	return detectExportsAndCalls(files, pyExportPattern, 1, callPatterns)
}
