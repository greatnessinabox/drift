package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/greatnessinabox/drift/internal/config"
)

type GoAnalyzer struct{}

func (g *GoAnalyzer) Language() Language { return LangGo }

func (g *GoAnalyzer) Extensions() []string { return []string{".go"} }

func (g *GoAnalyzer) FindFiles(root string, exclude []string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			for _, ex := range exclude {
				if info.Name() == ex {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (g *GoAnalyzer) AnalyzeComplexity(files []string) ([]FunctionComplexity, int) {
	var results []FunctionComplexity
	fset := token.NewFileSet()
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		funcs := analyzeComplexity(fset, f, path)
		results = append(results, funcs...)
	}
	return results, len(results)
}

func (g *GoAnalyzer) AnalyzeDeps(root string) ([]DepStatus, error) {
	return analyzeDeps(root)
}

func (g *GoAnalyzer) AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation {
	if len(rules) == 0 {
		return nil
	}
	fset := token.NewFileSet()
	var allFiles []*ast.File
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, f)
	}
	return analyzeImports(fset, allFiles, rules, root)
}

func (g *GoAnalyzer) AnalyzeDeadCode(files []string) []DeadFunction {
	fset := token.NewFileSet()
	var allFiles []*ast.File
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, f)
	}
	return analyzeDeadCode(fset, allFiles)
}
