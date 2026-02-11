package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/greatnessinabox/drift/internal/config"
)

type Results struct {
	Complexity   []FunctionComplexity
	Dependencies []DepStatus
	Violations   []BoundaryViolation
	DeadCode     []DeadFunction
	FileCount    int
	FuncCount    int
}

type Analyzer struct {
	cfg *config.Config
	mu  sync.Mutex
}

func New(cfg *config.Config) *Analyzer {
	return &Analyzer{cfg: cfg}
}

func (a *Analyzer) Run() (*Results, error) {
	results := &Results{}

	goFiles, err := a.findGoFiles()
	if err != nil {
		return nil, fmt.Errorf("finding Go files: %w", err)
	}
	results.FileCount = len(goFiles)

	fset := token.NewFileSet()
	var allFiles []*ast.File

	for _, path := range goFiles {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, f)

		funcs := analyzeComplexity(fset, f, path)
		results.Complexity = append(results.Complexity, funcs...)
		results.FuncCount += len(funcs)
	}

	deps, err := analyzeDeps(a.cfg.Root)
	if err == nil {
		results.Dependencies = deps
	}

	violations := analyzeImports(fset, allFiles, a.cfg.Boundaries, a.cfg.Root)
	results.Violations = violations

	results.DeadCode = analyzeDeadCode(fset, allFiles)

	sortComplexityDesc(results.Complexity)

	return results, nil
}

func (a *Analyzer) RunSingle(path string) (*Results, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !strings.HasSuffix(path, ".go") {
		return &Results{}, nil
	}

	if strings.HasSuffix(path, "_test.go") {
		return &Results{}, nil
	}

	results := &Results{}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return results, nil
	}

	funcs := analyzeComplexity(fset, f, path)
	results.Complexity = funcs
	results.FuncCount = len(funcs)
	results.FileCount = 1

	return results, nil
}

func (a *Analyzer) findGoFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(a.cfg.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			for _, ex := range a.cfg.Exclude {
				if name == ex {
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

func sortComplexityDesc(funcs []FunctionComplexity) {
	for i := 0; i < len(funcs); i++ {
		for j := i + 1; j < len(funcs); j++ {
			if funcs[j].Complexity > funcs[i].Complexity {
				funcs[i], funcs[j] = funcs[j], funcs[i]
			}
		}
	}
}
