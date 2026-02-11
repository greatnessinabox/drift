package analyzer

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/greatnessinabox/drift/internal/config"
)

type BoundaryViolation struct {
	File   string
	Line   int
	From   string
	To     string
	Import string
}

func analyzeImports(fset *token.FileSet, files []*ast.File, rules []config.BoundaryRule, root string) []BoundaryViolation {
	if len(rules) == 0 {
		return nil
	}

	var violations []BoundaryViolation

	for _, f := range files {
		pos := fset.Position(f.Pos())
		filePath := pos.Filename

		relPath, err := filepath.Rel(root, filePath)
		if err != nil {
			continue
		}

		fileDir := filepath.Dir(relPath)

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)

			for _, rule := range rules {
				from, to := parseBoundaryRule(rule.Deny)
				if from == "" || to == "" {
					continue
				}

				if matchesPath(fileDir, from) && matchesImport(importPath, to) {
					impPos := fset.Position(imp.Pos())
					violations = append(violations, BoundaryViolation{
						File:   filepath.Base(filePath),
						Line:   impPos.Line,
						From:   from,
						To:     to,
						Import: importPath,
					})
				}
			}
		}
	}

	return violations
}

func parseBoundaryRule(deny string) (string, string) {
	parts := strings.Split(deny, "->")
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func matchesPath(dir, pattern string) bool {
	dir = filepath.ToSlash(dir)
	pattern = filepath.ToSlash(pattern)
	return strings.HasPrefix(dir, pattern) || dir == pattern
}

func matchesImport(importPath, pattern string) bool {
	return strings.Contains(importPath, pattern)
}
