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

type RubyAnalyzer struct{}

func (r *RubyAnalyzer) Language() Language { return LangRuby }

func (r *RubyAnalyzer) Extensions() []string { return []string{".rb"} }

func (r *RubyAnalyzer) FindFiles(root string, exclude []string) ([]string, error) {
	exclude = append(exclude, ".bundle", "vendor", "tmp", ".ruby-lsp")
	skip := []string{"_test.rb", "_spec.rb", "spec/", "test/"}
	return walkFiles(root, exclude, r.Extensions(), skip)
}

var rbFuncPattern = regexp.MustCompile(`^(\s*)def\s+(self\.)?(\w+[?!=]?)`)

var rbComplexityPatterns = []complexityPattern{
	{regexp.MustCompile(`\bif\b`), 1},
	{regexp.MustCompile(`\belsif\b`), 1},
	{regexp.MustCompile(`\bunless\b`), 1},
	{regexp.MustCompile(`\bfor\b`), 1},
	{regexp.MustCompile(`\bwhile\b`), 1},
	{regexp.MustCompile(`\buntil\b`), 1},
	{regexp.MustCompile(`\bwhen\b`), 1},
	{regexp.MustCompile(`\brescue\b`), 1},
	{regexp.MustCompile(`&&`), 1},
	{regexp.MustCompile(`\|\|`), 1},
}

func (r *RubyAnalyzer) AnalyzeComplexity(files []string) ([]FunctionComplexity, int) {
	var results []FunctionComplexity
	for _, path := range files {
		funcs := analyzeRubyComplexity(path)
		results = append(results, funcs...)
	}
	return results, len(results)
}

func analyzeRubyComplexity(path string) []FunctionComplexity {
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
		matches := rbFuncPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		indent := len(matches[1])
		prefix := matches[2] // "self." or ""
		name := matches[3]

		if strings.HasPrefix(name, "_") {
			continue
		}

		if prefix != "" {
			name = "self." + name
		}

		// Track def/end depth to find function boundary
		complexity := 1
		depth := 1
		for j := i + 1; j < len(lines); j++ {
			bodyLine := lines[j]
			trimmed := strings.TrimSpace(bodyLine)

			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			// Count nested def/class/module/do/begin/if/unless/while/for/case/until as depth openers
			bodyIndent := len(bodyLine) - len(strings.TrimLeft(bodyLine, " \t"))
			if bodyIndent > indent {
				// Check for block openers
				if regexp.MustCompile(`\b(def|class|module|do|begin)\b`).MatchString(trimmed) &&
					!strings.Contains(trimmed, " end") {
					depth++
				}
				if regexp.MustCompile(`\b(if|unless|while|for|until|case)\b`).MatchString(trimmed) &&
					!strings.HasSuffix(trimmed, "end") &&
					!strings.Contains(trimmed, " then ") {
					// Only count standalone if/unless blocks, not inline modifiers
					if regexp.MustCompile(`^\s*(if|unless|while|for|until|case)\b`).MatchString(bodyLine) {
						depth++
					}
				}
			}

			if trimmed == "end" || strings.HasPrefix(trimmed, "end ") || strings.HasPrefix(trimmed, "end#") {
				endIndent := len(bodyLine) - len(strings.TrimLeft(bodyLine, " \t"))
				if endIndent <= indent {
					depth--
					if depth <= 0 {
						break
					}
				} else {
					depth--
				}
			}

			for _, p := range rbComplexityPatterns {
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

func (r *RubyAnalyzer) AnalyzeDeps(root string) ([]DepStatus, error) {
	gemfilePath := filepath.Join(root, "Gemfile")
	f, err := os.Open(gemfilePath)
	if err != nil {
		return nil, fmt.Errorf("reading Gemfile: %w", err)
	}
	defer f.Close()

	var results []DepStatus
	scanner := bufio.NewScanner(f)
	gemPattern := regexp.MustCompile(`gem\s+['"]([^'"]+)['"](?:\s*,\s*['"]([^'"]+)['"])?`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		m := gemPattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		name := m[1]
		version := ""
		if len(m) > 2 {
			version = strings.TrimLeft(m[2], "~>= ")
		}

		dep := DepStatus{
			Module:         name,
			CurrentVersion: version,
		}

		latest, err := fetchRubyGemsLatest(name)
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

type rubyGemsResponse struct {
	Version string `json:"version"`
}

func fetchRubyGemsLatest(name string) (string, error) {
	var resp rubyGemsResponse
	url := fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", name)
	if err := fetchJSON(url, &resp, ""); err != nil {
		return "", err
	}
	return resp.Version, nil
}

var rbImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`require\s+['"](\S+)['"]`),
	regexp.MustCompile(`require_relative\s+['"](\S+)['"]`),
}

func (r *RubyAnalyzer) AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation {
	var violations []BoundaryViolation
	for _, path := range files {
		imports := extractImports(path, rbImportPatterns, 1)
		violations = append(violations, checkBoundaryViolations(path, imports, rules, root)...)
	}
	return violations
}

var rbExportPattern = regexp.MustCompile(`^\s*def\s+(?:self\.)?(\w+[?!=]?)`)

func (r *RubyAnalyzer) AnalyzeDeadCode(files []string) []DeadFunction {
	callPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\w+[?!]?\(`),
		regexp.MustCompile(`\.\w+[?!]?`),
	}
	return detectExportsAndCalls(files, rbExportPattern, 1, callPatterns)
}
