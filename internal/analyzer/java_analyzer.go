package analyzer

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/greatnessinabox/drift/internal/config"
)

type JavaAnalyzer struct{}

func (j *JavaAnalyzer) Language() Language { return LangJava }

func (j *JavaAnalyzer) Extensions() []string { return []string{".java"} }

func (j *JavaAnalyzer) FindFiles(root string, exclude []string) ([]string, error) {
	exclude = append(exclude, "target", "build", ".gradle", ".idea", "bin", "out")
	skip := []string{"Test.java", "Tests.java", "IT.java"}
	return walkFiles(root, exclude, j.Extensions(), skip)
}

var javaFuncPattern = regexp.MustCompile(
	`(?:public|private|protected|static|\s)+[\w<>\[\]]+\s+(\w+)\s*\(`,
)

var javaComplexityPatterns = []complexityPattern{
	{regexp.MustCompile(`\bif\s*\(`), 1},
	{regexp.MustCompile(`\belse\s+if\b`), 1},
	{regexp.MustCompile(`\bfor\s*\(`), 1},
	{regexp.MustCompile(`\bwhile\s*\(`), 1},
	{regexp.MustCompile(`\bdo\s*\{`), 1},
	{regexp.MustCompile(`\bcase\s+[^:]+:`), 1},
	{regexp.MustCompile(`\bcatch\s*\(`), 1},
	{regexp.MustCompile(`&&`), 1},
	{regexp.MustCompile(`\|\|`), 1},
	{regexp.MustCompile(`\?\s*[^:]+\s*:`), 1}, // ternary
}

func (j *JavaAnalyzer) AnalyzeComplexity(files []string) ([]FunctionComplexity, int) {
	var results []FunctionComplexity
	for _, path := range files {
		funcs := scanHeuristicComplexity(path, javaFuncPattern, 1, javaComplexityPatterns)
		results = append(results, funcs...)
	}
	return results, len(results)
}

// Maven POM parsing

type pomProject struct {
	Dependencies struct {
		Dependency []pomDep `xml:"dependency"`
	} `xml:"dependencies"`
}

type pomDep struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

func (j *JavaAnalyzer) AnalyzeDeps(root string) ([]DepStatus, error) {
	// Try pom.xml first
	pomPath := filepath.Join(root, "pom.xml")
	if _, err := os.Stat(pomPath); err == nil {
		return parsePomDeps(pomPath)
	}

	// Try build.gradle
	gradlePath := filepath.Join(root, "build.gradle")
	if _, err := os.Stat(gradlePath); err == nil {
		return parseGradleDeps(gradlePath)
	}

	return nil, fmt.Errorf("no Java build file found")
}

func parsePomDeps(path string) ([]DepStatus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading pom.xml: %w", err)
	}

	var pom pomProject
	if err := xml.Unmarshal(data, &pom); err != nil {
		return nil, fmt.Errorf("parsing pom.xml: %w", err)
	}

	var results []DepStatus
	for _, dep := range pom.Dependencies.Dependency {
		if dep.Version == "" || strings.HasPrefix(dep.Version, "${") {
			continue
		}

		ds := DepStatus{
			Module:         dep.GroupID + ":" + dep.ArtifactID,
			CurrentVersion: dep.Version,
		}

		latest, err := fetchMavenLatest(dep.GroupID, dep.ArtifactID)
		if err != nil {
			ds.Status = "unknown"
			ds.LatestVersion = "?"
		} else {
			ds.LatestVersion = latest
			if ds.CurrentVersion == latest {
				ds.Status = "current"
			} else {
				ds.StaleDays = 30
				ds.Status = "stale"
			}
		}

		results = append(results, ds)
	}
	return results, nil
}

var gradleDepPattern = regexp.MustCompile(
	`(?:implementation|api|compile|testImplementation)\s+['"]([^:]+):([^:]+):([^'"]+)['"]`,
)

func parseGradleDeps(path string) ([]DepStatus, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading build.gradle: %w", err)
	}
	defer f.Close()

	var results []DepStatus
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if m := gradleDepPattern.FindStringSubmatch(line); m != nil {
			groupID, artifactID, version := m[1], m[2], m[3]

			ds := DepStatus{
				Module:         groupID + ":" + artifactID,
				CurrentVersion: version,
			}

			latest, err := fetchMavenLatest(groupID, artifactID)
			if err != nil {
				ds.Status = "unknown"
				ds.LatestVersion = "?"
			} else {
				ds.LatestVersion = latest
				if ds.CurrentVersion == latest {
					ds.Status = "current"
				} else {
					ds.StaleDays = 30
					ds.Status = "stale"
				}
			}

			results = append(results, ds)
		}
	}
	return results, nil
}

type mavenSearchResponse struct {
	Response struct {
		Docs []struct {
			LatestVersion string `json:"latestVersion"`
		} `json:"docs"`
	} `json:"response"`
}

func fetchMavenLatest(groupID, artifactID string) (string, error) {
	var resp mavenSearchResponse
	url := fmt.Sprintf(
		"https://search.maven.org/solrsearch/select?q=g:%%22%s%%22+AND+a:%%22%s%%22&rows=1&wt=json",
		groupID, artifactID,
	)
	if err := fetchJSON(url, &resp, ""); err != nil {
		return "", err
	}
	if len(resp.Response.Docs) == 0 {
		return "", fmt.Errorf("not found on Maven Central")
	}
	return resp.Response.Docs[0].LatestVersion, nil
}

var javaImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^import\s+(?:static\s+)?(\S+);`),
}

func (j *JavaAnalyzer) AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation {
	var violations []BoundaryViolation
	for _, path := range files {
		imports := extractImports(path, javaImportPatterns, 1)
		violations = append(violations, checkBoundaryViolations(path, imports, rules, root)...)
	}
	return violations
}

var javaExportPattern = regexp.MustCompile(`public\s+(?:static\s+)?(?:final\s+)?(?:[\w<>\[\]]+\s+)?(\w+)\s*\(`)

func (j *JavaAnalyzer) AnalyzeDeadCode(files []string) []DeadFunction {
	callPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\w+\(`),
	}
	return detectExportsAndCalls(files, javaExportPattern, 1, callPatterns)
}
