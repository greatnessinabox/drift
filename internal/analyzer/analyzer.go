package analyzer

import (
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
	Language     Language
}

type Analyzer struct {
	cfg  *config.Config
	lang LanguageAnalyzer
	mu   sync.Mutex
}

func New(cfg *config.Config) *Analyzer {
	lang := detectOrConfiguredLanguage(cfg)
	return &Analyzer{cfg: cfg, lang: lang}
}

func (a *Analyzer) Extensions() []string {
	return a.lang.Extensions()
}

func (a *Analyzer) DetectedLanguage() Language {
	return a.lang.Language()
}

func (a *Analyzer) Run() (*Results, error) {
	results := &Results{
		Language: a.lang.Language(),
	}

	files, err := a.lang.FindFiles(a.cfg.Root, a.cfg.Exclude)
	if err != nil {
		return nil, err
	}
	results.FileCount = len(files)

	complexity, funcCount := a.lang.AnalyzeComplexity(files)
	results.Complexity = complexity
	results.FuncCount = funcCount

	deps, err := a.lang.AnalyzeDeps(a.cfg.Root)
	if err == nil {
		results.Dependencies = deps
	}

	results.Violations = a.lang.AnalyzeImports(files, a.cfg.Boundaries, a.cfg.Root)
	results.DeadCode = a.lang.AnalyzeDeadCode(files)

	sortComplexityDesc(results.Complexity)

	return results, nil
}

func (a *Analyzer) RunSingle(path string) (*Results, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	results := &Results{
		Language: a.lang.Language(),
	}

	matched := false
	for _, ext := range a.lang.Extensions() {
		if len(path) > len(ext) && path[len(path)-len(ext):] == ext {
			matched = true
			break
		}
	}
	if !matched {
		return results, nil
	}

	files := []string{path}
	complexity, funcCount := a.lang.AnalyzeComplexity(files)
	results.Complexity = complexity
	results.FuncCount = funcCount
	results.FileCount = 1

	return results, nil
}

func detectOrConfiguredLanguage(cfg *config.Config) LanguageAnalyzer {
	if cfg.Language != "" {
		return NewLanguageAnalyzer(Language(cfg.Language))
	}
	detected := DetectLanguage(cfg.Root)
	return NewLanguageAnalyzer(detected)
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
