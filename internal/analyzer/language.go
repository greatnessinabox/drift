package analyzer

import (
	"os"
	"path/filepath"

	"github.com/greatnessinabox/drift/internal/config"
)

type Language string

const (
	LangGo         Language = "go"
	LangTypeScript Language = "typescript"
	LangPython     Language = "python"
	LangRust       Language = "rust"
	LangJava       Language = "java"
	LangRuby       Language = "ruby"
	LangPHP        Language = "php"
	LangCSharp     Language = "csharp"
	LangUnknown    Language = "unknown"
)

type LanguageAnalyzer interface {
	Language() Language
	Extensions() []string
	FindFiles(root string, exclude []string) ([]string, error)
	AnalyzeComplexity(files []string) ([]FunctionComplexity, int)
	AnalyzeDeps(root string) ([]DepStatus, error)
	AnalyzeImports(files []string, rules []config.BoundaryRule, root string) []BoundaryViolation
	AnalyzeDeadCode(files []string) []DeadFunction
}

func DetectLanguage(root string) Language {
	checks := []struct {
		file string
		lang Language
	}{
		{"go.mod", LangGo},
		{"package.json", LangTypeScript},
		{"pyproject.toml", LangPython},
		{"requirements.txt", LangPython},
		{"Cargo.toml", LangRust},
		{"pom.xml", LangJava},
		{"build.gradle", LangJava},
		{"Gemfile", LangRuby},
		{"composer.json", LangPHP},
	}

	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(root, c.file)); err == nil {
			return c.lang
		}
	}

	// Check for .csproj files (C# projects)
	csprojFiles, _ := filepath.Glob(filepath.Join(root, "*.csproj"))
	if len(csprojFiles) > 0 {
		return LangCSharp
	}
	// Also check one level deep for .NET solutions
	csprojFiles, _ = filepath.Glob(filepath.Join(root, "*", "*.csproj"))
	if len(csprojFiles) > 0 {
		return LangCSharp
	}

	return LangUnknown
}

func NewLanguageAnalyzer(lang Language) LanguageAnalyzer {
	switch lang {
	case LangGo:
		return &GoAnalyzer{}
	case LangTypeScript:
		return &TypeScriptAnalyzer{}
	case LangPython:
		return &PythonAnalyzer{}
	case LangRust:
		return &RustAnalyzer{}
	case LangJava:
		return &JavaAnalyzer{}
	case LangRuby:
		return &RubyAnalyzer{}
	case LangPHP:
		return &PHPAnalyzer{}
	case LangCSharp:
		return &CSharpAnalyzer{}
	default:
		return &GoAnalyzer{}
	}
}
