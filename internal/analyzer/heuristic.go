package analyzer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/greatnessinabox/drift/internal/config"
)

func walkFiles(root string, exclude, extensions, skipPatterns []string) ([]string, error) {
	var files []string
	extSet := make(map[string]bool)
	for _, e := range extensions {
		extSet[e] = true
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			for _, ex := range exclude {
				if name == ex {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := filepath.Ext(path)
		if !extSet[ext] {
			return nil
		}

		for _, skip := range skipPatterns {
			if strings.Contains(path, skip) {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})
	return files, err
}

type complexityPattern struct {
	pattern *regexp.Regexp
	weight  int
}

type funcBoundary struct {
	name  string
	file  string
	line  int
	start int
	end   int
}

func scanHeuristicComplexity(
	path string,
	funcPattern *regexp.Regexp,
	nameGroup int,
	patterns []complexityPattern,
) []FunctionComplexity {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var allLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	funcs := detectFunctions(allLines, funcPattern, nameGroup, path)

	var results []FunctionComplexity
	for _, fn := range funcs {
		complexity := 1
		for i := fn.start; i < fn.end && i < len(allLines); i++ {
			line := allLines[i]
			for _, p := range patterns {
				if p.pattern.MatchString(line) {
					complexity += p.weight
				}
			}
		}
		results = append(results, FunctionComplexity{
			File:       filepath.Base(fn.file),
			Name:       fn.name,
			Line:       fn.line,
			Complexity: complexity,
		})
	}
	return results
}

func detectFunctions(lines []string, funcPattern *regexp.Regexp, nameGroup int, path string) []funcBoundary {
	var funcs []funcBoundary
	depth := 0

	for i, line := range lines {
		if matches := funcPattern.FindStringSubmatch(line); matches != nil && nameGroup < len(matches) {
			name := matches[nameGroup]
			if name == "" {
				name = "anonymous"
			}
			funcs = append(funcs, funcBoundary{
				name:  name,
				file:  path,
				line:  i + 1,
				start: i,
				end:   len(lines),
			})
		}
	}

	// Set end boundaries using brace depth tracking
	for idx := range funcs {
		startLine := funcs[idx].start
		depth = 0
		started := false
		for i := startLine; i < len(lines); i++ {
			line := lines[i]
			for _, ch := range line {
				if ch == '{' {
					depth++
					started = true
				} else if ch == '}' {
					depth--
				}
			}
			if started && depth <= 0 {
				funcs[idx].end = i + 1
				break
			}
		}
	}

	return funcs
}

type heuristicImportMatch struct {
	path string
	line int
}

func extractImports(filePath string, patterns []*regexp.Regexp, groupIndex int) []heuristicImportMatch {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var imports []heuristicImportMatch
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		for _, p := range patterns {
			if matches := p.FindStringSubmatch(line); matches != nil && groupIndex < len(matches) {
				imports = append(imports, heuristicImportMatch{
					path: matches[groupIndex],
					line: lineNum,
				})
			}
		}
	}
	return imports
}

func checkBoundaryViolations(filePath string, imports []heuristicImportMatch, rules []config.BoundaryRule, root string) []BoundaryViolation {
	if len(rules) == 0 {
		return nil
	}

	relPath, err := filepath.Rel(root, filePath)
	if err != nil {
		return nil
	}
	fileDir := filepath.Dir(relPath)

	var violations []BoundaryViolation
	for _, imp := range imports {
		for _, rule := range rules {
			from, to := parseBoundaryRule(rule.Deny)
			if from == "" || to == "" {
				continue
			}
			if matchesPath(fileDir, from) && matchesImport(imp.path, to) {
				violations = append(violations, BoundaryViolation{
					File:   filepath.Base(filePath),
					Line:   imp.line,
					From:   from,
					To:     to,
					Import: imp.path,
				})
			}
		}
	}
	return violations
}

func detectExportsAndCalls(
	files []string,
	exportPattern *regexp.Regexp,
	exportNameGroup int,
	callPatterns []*regexp.Regexp,
) []DeadFunction {
	type exportInfo struct {
		file string
		name string
		line int
	}

	exported := make(map[string]exportInfo)
	called := make(map[string]bool)

	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			if matches := exportPattern.FindStringSubmatch(line); matches != nil && exportNameGroup < len(matches) {
				name := matches[exportNameGroup]
				if name != "" {
					exported[name] = exportInfo{
						file: filepath.Base(path),
						name: name,
						line: lineNum,
					}
				}
			}

			for _, cp := range callPatterns {
				for _, m := range cp.FindAllString(line, -1) {
					called[m] = true
				}
			}
		}
		f.Close()
	}

	// Also scan all files for any reference to exported names
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(content)
		for name := range exported {
			if strings.Count(text, name) > 1 {
				called[name] = true
			}
		}
	}

	var dead []DeadFunction
	for name, info := range exported {
		if !called[name] {
			dead = append(dead, DeadFunction{
				File: info.file,
				Name: info.name,
				Line: info.line,
			})
		}
	}
	return dead
}

func fetchJSON(url string, target interface{}, userAgent string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
