package analyzer

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/greatnessinabox/drift/internal/config"
)

type CSharpAnalyzer struct{}

func (c *CSharpAnalyzer) Language() Language { return LangCSharp }

func (c *CSharpAnalyzer) Extensions() []string { return []string{".cs"} }

func (c *CSharpAnalyzer) FindFiles(root string, exclude []string) ([]string, error) {
	exclude = append(exclude, "bin", "obj", ".vs", "packages", "TestResults")
	skip := []string{"Tests.cs", "Test.cs", ".Tests/", ".Test/"}
	return walkFiles(root, exclude, c.Extensions(), skip)
}

var csFuncPattern = regexp.MustCompile(
	`(?:public|private|protected|internal|static|async|virtual|override|abstract|\s)+[\w<>\[\]?]+\s+(\w+)\s*\(`,
)

var csComplexityPatterns = []complexityPattern{
	{regexp.MustCompile(`\bif\s*\(`), 1},
	{regexp.MustCompile(`\belse\s+if\b`), 1},
	{regexp.MustCompile(`\bfor\s*\(`), 1},
	{regexp.MustCompile(`\bforeach\s*\(`), 1},
	{regexp.MustCompile(`\bwhile\s*\(`), 1},
	{regexp.MustCompile(`\bdo\s*\{`), 1},
	{regexp.MustCompile(`\bcase\s+[^:]+:`), 1},
	{regexp.MustCompile(`\bcatch\s*\(`), 1},
	{regexp.MustCompile(`\bswitch\s*\(`), 1},
	{regexp.MustCompile(`&&`), 1},
	{regexp.MustCompile(`\|\|`), 1},
	{regexp.MustCompile(`\?\?`), 1},
	{regexp.MustCompile(`\?\s*[^:]+\s*:`), 1},
}

func (c *CSharpAnalyzer) AnalyzeComplexity(files []string) ([]FunctionComplexity, int) {
	var results []FunctionComplexity
	for _, path := range files {
		funcs := scanHeuristicComplexity(path, csFuncPattern, 1, csComplexityPatterns)
		results = append(results, funcs...)
	}
	return results, len(results)
}

// .csproj XML parsing

type csprojProject struct {
	ItemGroups []csprojItemGroup `xml:"ItemGroup"`
}

type csprojItemGroup struct {
	PackageReferences []csprojPackageRef `xml:"PackageReference"`
}

type csprojPackageRef struct {
	Include string `xml:"Include,attr"`
	Version string `xml:"Version,attr"`
}

func (c *CSharpAnalyzer) AnalyzeDeps(root string) ([]DepStatus, error) {
	// Find .csproj file
	csprojFiles, err := filepath.Glob(filepath.Join(root, "*.csproj"))
	if err != nil || len(csprojFiles) == 0 {
		// Try one level deep (common in .NET solutions)
		csprojFiles, err = filepath.Glob(filepath.Join(root, "*", "*.csproj"))
		if err != nil || len(csprojFiles) == 0 {
			return nil, fmt.Errorf("no .csproj file found")
		}
	}

	var results []DepStatus
	for _, csprojPath := range csprojFiles {
		deps, err := parseCsprojDeps(csprojPath)
		if err != nil {
			continue
		}
		results = append(results, deps...)
	}
	return results, nil
}

func parseCsprojDeps(path string) ([]DepStatus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading .csproj: %w", err)
	}

	var proj csprojProject
	if err := xml.Unmarshal(data, &proj); err != nil {
		return nil, fmt.Errorf("parsing .csproj: %w", err)
	}

	var results []DepStatus
	for _, ig := range proj.ItemGroups {
		for _, ref := range ig.PackageReferences {
			if ref.Include == "" {
				continue
			}

			dep := DepStatus{
				Module:         ref.Include,
				CurrentVersion: ref.Version,
			}

			latest, err := fetchNuGetLatest(ref.Include)
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
	}
	return results, nil
}

type nugetIndexResponse struct {
	Versions []string `json:"versions"`
}

func fetchNuGetLatest(name string) (string, error) {
	var resp nugetIndexResponse
	url := fmt.Sprintf("https://api.nuget.org/v3-flatcontainer/%s/index.json", strings.ToLower(name))
	if err := fetchJSON(url, &resp, ""); err != nil {
		return "", err
	}
	if len(resp.Versions) == 0 {
		return "", fmt.Errorf("no versions found on NuGet")
	}
	// Return last version (newest)
	return resp.Versions[len(resp.Versions)-1], nil
}

var csImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^using\s+(\S+);`),
	regexp.MustCompile(`^using\s+static\s+(\S+);`),
}

func (c *CSharpAnalyzer) AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation {
	var violations []BoundaryViolation
	for _, path := range files {
		imports := extractImports(path, csImportPatterns, 1)
		violations = append(violations, checkBoundaryViolations(path, imports, rules, root)...)
	}
	return violations
}

var csExportPattern = regexp.MustCompile(
	`public\s+(?:static\s+)?(?:async\s+)?(?:virtual\s+)?(?:override\s+)?(?:[\w<>\[\]?]+\s+)?(\w+)\s*\(`,
)

func (c *CSharpAnalyzer) AnalyzeDeadCode(files []string) []DeadFunction {
	callPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\w+\(`),
		regexp.MustCompile(`\.\w+\(`),
	}
	return detectExportsAndCalls(files, csExportPattern, 1, callPatterns)
}
