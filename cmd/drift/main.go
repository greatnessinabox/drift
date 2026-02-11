package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/greatnessinabox/drift/internal/analyzer"
	"github.com/greatnessinabox/drift/internal/config"
	"github.com/greatnessinabox/drift/internal/health"
	"github.com/greatnessinabox/drift/internal/tui"
	"github.com/greatnessinabox/drift/internal/watcher"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	cfgFile string
)

func main() {
	root := &cobra.Command{
		Use:     "drift",
		Short:   "Real-time codebase health dashboard",
		Long:    "drift watches your codebase in real-time, detects code health degradation, and uses AI to diagnose problems.",
		Version: version,
		RunE:    runDashboard,
	}

	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: .drift.yaml)")

	root.AddCommand(newReportCmd())
	root.AddCommand(newSnapshotCmd())
	root.AddCommand(newInitCmd())
	root.AddCommand(newCheckCmd())
	root.AddCommand(newFixCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runDashboard(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	a := analyzer.New(cfg)

	results, err := a.Run()
	if err != nil {
		return fmt.Errorf("initial analysis: %w", err)
	}

	scorer := health.NewScorer(cfg)
	score := scorer.Calculate(results)

	w, err := watcher.New(cfg.Root, cfg.Exclude, a.Extensions())
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}

	app := tui.New(cfg, a, scorer, score, results, w)
	return app.Run()
}

func newReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Generate a terminal-formatted health report",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			a := analyzer.New(cfg)
			results, err := a.Run()
			if err != nil {
				return err
			}
			scorer := health.NewScorer(cfg)
			score := scorer.Calculate(results)
			tui.PrintReport(cfg, score, results)
			return nil
		},
	}
}

func newSnapshotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "snapshot",
		Short: "Output a JSON health snapshot for CI",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			a := analyzer.New(cfg)
			results, err := a.Run()
			if err != nil {
				return err
			}
			scorer := health.NewScorer(cfg)
			score := scorer.Calculate(results)
			return tui.PrintSnapshot(score, results)
		},
	}
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize drift configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.RunInitWizard()
		},
	}
}

func newCheckCmd() *cobra.Command {
	var failUnder float64

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check health score and exit with error if below threshold (for CI)",
		Long: `Check runs analysis and exits with code 1 if the health score is below the threshold.
Useful for CI pipelines to enforce code health standards.

Example:
  drift check --fail-under 70`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}

			a := analyzer.New(cfg)
			results, err := a.Run()
			if err != nil {
				return err
			}

			scorer := health.NewScorer(cfg)
			score := scorer.Calculate(results)

			fmt.Printf("Health Score: %.1f/100\n", score.Total)

			if score.Total < failUnder {
				fmt.Printf("❌ Score %.1f is below threshold %.1f\n", score.Total, failUnder)
				fmt.Printf("\nBreakdown:\n")
				fmt.Printf("  Complexity:  %.1f/100\n", score.Complexity)
				fmt.Printf("  Dependencies: %.1f/100\n", score.Deps)
				fmt.Printf("  Boundaries:   %.1f/100\n", score.Boundaries)
				fmt.Printf("  Dead Code:    %.1f/100\n", score.DeadCode)
				os.Exit(1)
			}

			fmt.Printf("✅ Score %.1f meets threshold %.1f\n", score.Total, failUnder)
			return nil
		},
	}

	cmd.Flags().Float64Var(&failUnder, "fail-under", 70.0, "Minimum health score required (0-100)")

	return cmd
}

func newFixCmd() *cobra.Command {
	var interactive bool
	var limit int

	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Interactively fix code health issues using GitHub Copilot CLI",
		Long: `Fix analyzes your codebase and uses GitHub Copilot CLI to suggest improvements
for code health issues like high complexity, outdated dependencies, and more.

Requires GitHub Copilot CLI to be installed:
  gh extension install github/gh-copilot

Example:
  drift fix                    # Interactive mode
  drift fix --limit 3          # Fix top 3 issues only
  drift fix --non-interactive  # Show suggestions without prompting`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}

			a := analyzer.New(cfg)
			results, err := a.Run()
			if err != nil {
				return err
			}

			scorer := health.NewScorer(cfg)
			score := scorer.Calculate(results)

			return runFixWorkflow(cfg, score, results, interactive, limit)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", true, "Prompt for each fix (use --non-interactive to disable)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 5, "Maximum number of issues to fix")

	return cmd
}

func runFixWorkflow(cfg *config.Config, score health.Score, results *analyzer.Results, interactive bool, limit int) error {
	// Check if gh copilot is available
	if !isGHCopilotAvailable() {
		fmt.Println("❌ GitHub Copilot CLI not found")
		fmt.Println("\nInstall it with:")
		fmt.Println("  gh extension install github/gh-copilot")
		fmt.Println("\nAlternatively, run 'drift report' for analysis without AI suggestions")
		return fmt.Errorf("gh copilot not available")
	}

	fmt.Printf("🔍 Analyzing codebase... (Score: %.1f/100)\n\n", score.Total)

	// Collect issues to fix
	var issues []fixIssue

	// Add complexity issues
	for i, fc := range results.Complexity {
		if i >= limit {
			break
		}
		if fc.Complexity > cfg.Thresholds.MaxComplexity {
			issues = append(issues, fixIssue{
				Type:        "complexity",
				Description: fmt.Sprintf("%s() in %s:%d (complexity: %d)", fc.Name, fc.File, fc.Line, fc.Complexity),
				File:        fc.File,
				Line:        fc.Line,
				Function:    fc.Name,
				Severity:    getSeverity(fc.Complexity, cfg.Thresholds.MaxComplexity),
			})
		}
	}

	if len(issues) == 0 {
		fmt.Println("✅ No issues found! Your codebase is healthy.")
		return nil
	}

	fmt.Printf("Found %d issue(s) to fix:\n\n", len(issues))

	for i, issue := range issues {
		fmt.Printf("%d. [%s] %s\n", i+1, issue.Severity, issue.Description)
	}

	fmt.Println()

	// Process each issue
	for i, issue := range issues {
		if !interactive {
			// Just show what would be fixed
			fmt.Printf("\n[%d/%d] %s\n", i+1, len(issues), issue.Description)
			fmt.Println("  (non-interactive mode: skipping)")
			continue
		}

		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("[%d/%d] %s\n", i+1, len(issues), issue.Description)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Get Copilot suggestion
		fmt.Println("\n🤖 Asking GitHub Copilot for suggestions...")
		suggestion, err := getCopilotSuggestion(cfg, issue)
		if err != nil {
			fmt.Printf("❌ Error getting suggestion: %v\n", err)
			continue
		}

		fmt.Println("\n" + suggestion)
		fmt.Println()

		// Ask user what to do
		fmt.Print("Apply this suggestion? [y/N/s(kip rest)] ")
		var response string
		fmt.Scanln(&response)

		switch response {
		case "y", "Y", "yes":
			fmt.Println("✅ Note: Copy the suggestion above and apply manually")
			fmt.Println("   (Automatic application coming in a future update)")
		case "s", "S", "skip":
			fmt.Println("⏭️  Skipping remaining issues")
			return nil
		default:
			fmt.Println("⏭️  Skipped")
		}
	}

	fmt.Println("\n✨ Fix workflow complete!")
	fmt.Println("💡 Run 'drift report' to see updated health metrics")

	return nil
}

type fixIssue struct {
	Type        string
	Description string
	File        string
	Line        int
	Function    string
	Severity    string
}

func getSeverity(complexity, threshold int) string {
	if complexity > threshold*2 {
		return "🔴 HIGH"
	} else if float64(complexity) > float64(threshold)*1.5 {
		return "🟡 MEDIUM"
	}
	return "🟢 LOW"
}

func isGHCopilotAvailable() bool {
	cmd := exec.Command("gh", "copilot", "--version")
	err := cmd.Run()
	return err == nil
}

func getCopilotSuggestion(cfg *config.Config, issue fixIssue) (string, error) {
	// Build prompt for Copilot
	prompt := buildCopilotPrompt(cfg, issue)

	// Call gh copilot suggest
	cmd := exec.Command("gh", "copilot", "suggest", prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh copilot failed: %w", err)
	}

	return string(output), nil
}

func buildCopilotPrompt(cfg *config.Config, issue fixIssue) string {
	// Read the function source code
	filePath := filepath.Join(cfg.Root, issue.File)
	sourceCode := readFunctionSource(filePath, issue.Line, 20)

	prompt := fmt.Sprintf(`Refactor this Go function to reduce cyclomatic complexity from %d to below %d.
Focus on extracting methods, simplifying conditionals, and improving readability.

Current code:
%s

Provide a refactored version with explanation.`,
		extractComplexity(issue.Description),
		cfg.Thresholds.MaxComplexity,
		sourceCode)

	return prompt
}

func readFunctionSource(filePath string, startLine, numLines int) string {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Sprintf("// Could not read file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	var lines []string

	for scanner.Scan() {
		if lineNum >= startLine && len(lines) < numLines {
			lines = append(lines, scanner.Text())
		}
		if len(lines) >= numLines {
			break
		}
		lineNum++
	}

	return strings.Join(lines, "\n")
}

func extractComplexity(description string) int {
	// Extract complexity number from description like "model.Update() (complexity: 25)"
	re := regexp.MustCompile(`complexity:\s*(\d+)`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		complexity, _ := strconv.Atoi(matches[1])
		return complexity
	}
	return 15 // default
}
