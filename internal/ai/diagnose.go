package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/greatnessinabox/drift/internal/analyzer"
	"github.com/greatnessinabox/drift/internal/config"
	"github.com/greatnessinabox/drift/internal/health"
)

func BuildDiagnosisPrompt(score health.Score, results *analyzer.Results) string {
	var sb strings.Builder

	sb.WriteString("Analyze this Go codebase health report and provide actionable recommendations:\n\n")

	sb.WriteString(fmt.Sprintf("Overall Health Score: %.0f/100\n", score.Total))
	sb.WriteString(fmt.Sprintf("  Complexity Score: %.0f/100\n", score.Complexity))
	sb.WriteString(fmt.Sprintf("  Dependencies Score: %.0f/100\n", score.Deps))
	sb.WriteString(fmt.Sprintf("  Boundaries Score: %.0f/100\n\n", score.Boundaries))

	sb.WriteString(fmt.Sprintf("Codebase: %d files, %d functions\n\n", results.FileCount, results.FuncCount))

	if len(results.Complexity) > 0 {
		sb.WriteString("Top Complex Functions:\n")
		count := 5
		if len(results.Complexity) < count {
			count = len(results.Complexity)
		}
		for i := 0; i < count; i++ {
			fc := results.Complexity[i]
			sb.WriteString(fmt.Sprintf("  - %s() in %s:%d — cyclomatic complexity %d\n",
				fc.Name, fc.File, fc.Line, fc.Complexity))
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

func RunDiagnosis(cfg *config.Config, score health.Score, results *analyzer.Results) (string, error) {
	provider, err := NewProvider(cfg.AI)
	if err != nil {
		return "", err
	}

	prompt := BuildDiagnosisPrompt(score, results)
	return provider.Diagnose(context.Background(), prompt)
}
