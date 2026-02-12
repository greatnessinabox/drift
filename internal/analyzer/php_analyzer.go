package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/greatnessinabox/drift/internal/config"
)

type PHPAnalyzer struct{}

func (p *PHPAnalyzer) Language() Language { return LangPHP }

func (p *PHPAnalyzer) Extensions() []string { return []string{".php"} }

func (p *PHPAnalyzer) FindFiles(root string, exclude []string) ([]string, error) {
	exclude = append(exclude, "vendor", "cache", ".phpunit.cache", "storage")
	skip := []string{"Test.php", "Tests.php", "test/", "tests/"}
	return walkFiles(root, exclude, p.Extensions(), skip)
}

var phpFuncPattern = regexp.MustCompile(
	`(?:public|private|protected|static|\s)*function\s+(\w+)\s*\(`,
)

var phpComplexityPatterns = []complexityPattern{
	{regexp.MustCompile(`\bif\s*\(`), 1},
	{regexp.MustCompile(`\belse\s*if\b`), 1},
	{regexp.MustCompile(`\belseif\s*\(`), 1},
	{regexp.MustCompile(`\bfor\s*\(`), 1},
	{regexp.MustCompile(`\bforeach\s*\(`), 1},
	{regexp.MustCompile(`\bwhile\s*\(`), 1},
	{regexp.MustCompile(`\bdo\s*\{`), 1},
	{regexp.MustCompile(`\bcase\s+[^:]+:`), 1},
	{regexp.MustCompile(`\bcatch\s*\(`), 1},
	{regexp.MustCompile(`&&`), 1},
	{regexp.MustCompile(`\|\|`), 1},
	{regexp.MustCompile(`\?\s*[^:]+\s*:`), 1},
}

func (p *PHPAnalyzer) AnalyzeComplexity(files []string) ([]FunctionComplexity, int) {
	var results []FunctionComplexity
	for _, path := range files {
		funcs := scanHeuristicComplexity(path, phpFuncPattern, 1, phpComplexityPatterns)
		results = append(results, funcs...)
	}
	return results, len(results)
}

func (p *PHPAnalyzer) AnalyzeDeps(root string) ([]DepStatus, error) {
	composerPath := filepath.Join(root, "composer.json")
	data, err := os.ReadFile(composerPath)
	if err != nil {
		return nil, fmt.Errorf("reading composer.json: %w", err)
	}

	var composer struct {
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, fmt.Errorf("parsing composer.json: %w", err)
	}

	var results []DepStatus
	for name, version := range composer.Require {
		// Skip PHP version and extensions
		if name == "php" || strings.HasPrefix(name, "ext-") {
			continue
		}

		cleanVersion := strings.TrimLeft(version, "^~>=<! ")

		dep := DepStatus{
			Module:         name,
			CurrentVersion: cleanVersion,
		}

		latest, err := fetchPackagistLatest(name)
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

type packagistResponse struct {
	Packages map[string][]struct {
		Version string `json:"version"`
	} `json:"packages"`
}

func fetchPackagistLatest(name string) (string, error) {
	var resp packagistResponse
	url := fmt.Sprintf("https://repo.packagist.org/p2/%s.json", name)
	if err := fetchJSON(url, &resp, ""); err != nil {
		return "", err
	}

	versions := resp.Packages[name]
	for _, v := range versions {
		// Skip dev versions
		if strings.Contains(v.Version, "dev") {
			continue
		}
		// Return first stable version (they're sorted newest first)
		return strings.TrimPrefix(v.Version, "v"), nil
	}

	if len(versions) > 0 {
		return strings.TrimPrefix(versions[0].Version, "v"), nil
	}
	return "", fmt.Errorf("no versions found")
}

var phpImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^use\s+(\S+);`),
	regexp.MustCompile(`require_once\s+['"](\S+)['"]`),
	regexp.MustCompile(`include\s+['"](\S+)['"]`),
}

func (p *PHPAnalyzer) AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation {
	var violations []BoundaryViolation
	for _, path := range files {
		imports := extractImports(path, phpImportPatterns, 1)
		violations = append(violations, checkBoundaryViolations(path, imports, rules, root)...)
	}
	return violations
}

var phpExportPattern = regexp.MustCompile(`public\s+(?:static\s+)?function\s+(\w+)`)

func (p *PHPAnalyzer) AnalyzeDeadCode(files []string) []DeadFunction {
	callPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\w+\(`),
		regexp.MustCompile(`->\w+\(`),
		regexp.MustCompile(`::\w+\(`),
	}
	return detectExportsAndCalls(files, phpExportPattern, 1, callPatterns)
}
