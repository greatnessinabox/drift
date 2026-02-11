package ai

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/greatnessinabox/drift/internal/analyzer"
	"github.com/greatnessinabox/drift/internal/config"
	"github.com/greatnessinabox/drift/internal/health"
)

func BuildDiagnosisPrompt(cfg *config.Config, score health.Score, results *analyzer.Results) string {
	var sb strings.Builder

	lang := string(results.Language)
	if lang == "" {
		lang = "Go"
	}
	sb.WriteString(fmt.Sprintf("Analyze this %s codebase health report and provide actionable recommendations:\n\n", lang))

	sb.WriteString(fmt.Sprintf("Overall Health Score: %.0f/100\n", score.Total))
	sb.WriteString(fmt.Sprintf("  Complexity Score: %.0f/100\n", score.Complexity))
	sb.WriteString(fmt.Sprintf("  Dependencies Score: %.0f/100\n", score.Deps))
	sb.WriteString(fmt.Sprintf("  Boundaries Score: %.0f/100\n\n", score.Boundaries))

	sb.WriteString(fmt.Sprintf("Codebase: %d files, %d functions\n\n", results.FileCount, results.FuncCount))

	if len(results.Complexity) > 0 {
		sb.WriteString("Top Complex Functions:\n")
		count := 3
		if len(results.Complexity) < count {
			count = len(results.Complexity)
		}
		for i := 0; i < count; i++ {
			fc := results.Complexity[i]
			sb.WriteString(fmt.Sprintf("  - %s() in %s:%d — cyclomatic complexity %d\n",
				fc.Name, fc.File, fc.Line, fc.Complexity))
			
			// Include code snippet for worst functions
			if snippet := getCodeSnippet(cfg.Root, fc.File, fc.Line, 10); snippet != "" {
				sb.WriteString(fmt.Sprintf("\n```%s\n%s\n```\n\n", lang, snippet))
			}
		}
		sb.WriteString("\n")
	}

	staleCount := 0
	for _, dep := range results.Dependencies {
		if dep.Status != "current" {
			staleCount++
		}
	}
	if staleCount > 0 {
		sb.WriteString(fmt.Sprintf("Stale Dependencies (%d):\n", staleCount))
		for _, dep := range results.Dependencies {
			if dep.Status != "current" {
				sb.WriteString(fmt.Sprintf("  - %s: current %s, latest %s (%d days behind)\n",
					dep.Module, dep.CurrentVersion, dep.LatestVersion, dep.StaleDays))
			}
		}
		sb.WriteString("\n")
	}

	if len(results.Violations) > 0 {
		sb.WriteString(fmt.Sprintf("Boundary Violations (%d):\n", len(results.Violations)))
		for _, v := range results.Violations {
			sb.WriteString(fmt.Sprintf("  - %s imports %s (%s:%d) — violates %s → %s boundary\n",
				v.File, v.Import, v.File, v.Line, v.From, v.To))
		}
	}

	return sb.String()
}

// getCodeSnippet reads a code snippet from a file starting at the given line
func getCodeSnippet(root, filename string, startLine, numLines int) string {
	// Try to find the file
	fullPath := filepath.Join(root, filename)
	
	// If not found, search for it
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Search for the file in the root directory
		matches, _ := filepath.Glob(filepath.Join(root, "**", filename))
		if len(matches) > 0 {
			fullPath = matches[0]
		} else {
			return ""
		}
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	var lines []string

	// Read up to the start line
	for scanner.Scan() && lineNum < startLine {
		lineNum++
	}

	// Read the snippet
	for scanner.Scan() && len(lines) < numLines {
		lines = append(lines, scanner.Text())
		lineNum++
	}

	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n")
}

func RunDiagnosis(cfg *config.Config, score health.Score, results *analyzer.Results) (string, error) {
	provider, err := NewProvider(cfg.AI)
	if err != nil {
		return "", err
	}

	prompt := BuildDiagnosisPrompt(cfg, score, results)
	return provider.Diagnose(context.Background(), prompt)
}
