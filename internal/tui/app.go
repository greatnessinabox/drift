package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/greatnessinabox/drift/internal/ai"
	"github.com/greatnessinabox/drift/internal/analyzer"
	"github.com/greatnessinabox/drift/internal/config"
	"github.com/greatnessinabox/drift/internal/health"
	"github.com/greatnessinabox/drift/internal/history"
	"github.com/greatnessinabox/drift/internal/watcher"
)

type focusPanel int

const (
	panelScore focusPanel = iota
	panelComplexity
	panelDeps
	panelBoundaries
	panelActivity
	panelCount
)

type activityEntry struct {
	file      string
	timestamp time.Time
}

type model struct {
	cfg     *config.Config
	ana     *analyzer.Analyzer
	scorer  *health.Scorer
	score   health.Score
	results *analyzer.Results
	watch   *watcher.Watcher

	width  int
	height int
	focus  focusPanel

	activity []activityEntry
	spinner  spinner.Model

	// Animation
	displayScore float64
	targetScore  float64
	animating    bool

	// AI diagnosis
	showDiagnosis bool
	diagnosisText string
	diagnosing    bool

	// Sparkline history
	sparklineData *history.SparklineData

	quitting bool
}

type fileChangedMsg struct {
	path      string
	timestamp time.Time
}

type analysisCompleteMsg struct {
	results *analyzer.Results
	score   health.Score
}

type animateTickMsg struct{}

type diagnosisCompleteMsg struct {
	text string
}

type historyCompleteMsg struct {
	data *history.SparklineData
}

func New(cfg *config.Config, ana *analyzer.Analyzer, scorer *health.Scorer, score health.Score, results *analyzer.Results, w *watcher.Watcher) *model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorCyan)

	return &model{
		cfg:          cfg,
		ana:          ana,
		scorer:       scorer,
		score:        score,
		results:      results,
		watch:        w,
		displayScore: score.Total,
		targetScore:  score.Total,
		spinner:      s,
	}
}

func (m *model) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion())
	_, err := p.Run()
	if m.watch != nil {
		m.watch.Close()
	}
	return err
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.listenForChanges(),
		m.loadHistory(),
		tea.WindowSize(),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showDiagnosis {
			switch msg.String() {
			case "esc", "q":
				m.showDiagnosis = false
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			m.focus = (m.focus + 1) % panelCount
		case "shift+tab":
			m.focus = (m.focus - 1 + panelCount) % panelCount
		case "r":
			cmds = append(cmds, m.runAnalysis())
		case "d":
			if !m.diagnosing {
				m.diagnosing = true
				cmds = append(cmds, m.runDiagnosis())
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case fileChangedMsg:
		m.activity = append([]activityEntry{{
			file:      msg.path,
			timestamp: msg.timestamp,
		}}, m.activity...)
		if len(m.activity) > 20 {
			m.activity = m.activity[:20]
		}
		cmds = append(cmds, m.runAnalysis(), m.listenForChanges())

	case analysisCompleteMsg:
		m.results = msg.results
		m.score = msg.score
		m.targetScore = msg.score.Total
		if m.displayScore != m.targetScore {
			m.animating = true
			cmds = append(cmds, m.animateTick())
		}

	case historyCompleteMsg:
		m.sparklineData = msg.data

	case animateTickMsg:
		diff := m.targetScore - m.displayScore
		if abs(diff) < 0.5 {
			m.displayScore = m.targetScore
			m.animating = false
		} else {
			m.displayScore += diff * 0.15
			cmds = append(cmds, m.animateTick())
		}

	case diagnosisCompleteMsg:
		m.diagnosing = false
		m.showDiagnosis = true
		m.diagnosisText = msg.text

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 {
		return "Loading..."
	}

	if m.showDiagnosis {
		return m.viewDiagnosis()
	}

	var sections []string

	sections = append(sections, m.viewHeader())
	sections = append(sections, m.viewScore())

	midLeft := m.viewComplexity()
	midRight := m.viewDeps()
	midSection := lipgloss.JoinHorizontal(lipgloss.Top, midLeft, midRight)
	sections = append(sections, midSection)

	botLeft := m.viewBoundaries()
	botRight := m.viewActivity()
	botSection := lipgloss.JoinHorizontal(lipgloss.Top, botLeft, botRight)
	sections = append(sections, botSection)

	sections = append(sections, m.viewFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *model) viewHeader() string {
	logo := logoStyle.Render("◆ DRIFT")
	subtitle := lipgloss.NewStyle().Foreground(colorDim).Render(" — codebase health monitor")
	header := logo + subtitle

	langLabel := string(m.results.Language)
	if langLabel == "" || langLabel == "unknown" {
		langLabel = "unknown"
	}
	fileInfo := lipgloss.NewStyle().Foreground(colorDim).Render(
		fmt.Sprintf("%s · %d files · %d functions", langLabel, m.results.FileCount, m.results.FuncCount),
	)

	padding := m.width - lipgloss.Width(header) - lipgloss.Width(fileInfo) - 4
	if padding < 1 {
		padding = 1
	}

	return "  " + header + strings.Repeat(" ", padding) + fileInfo + "  "
}

func (m *model) viewScore() string {
	score := m.displayScore
	total := 100.0

	barWidth := 30
	filled := int(score / total * float64(barWidth))
	empty := barWidth - filled

	bar := ""
	for i := 0; i < filled; i++ {
		// Calculate color based on position in bar (gradient effect)
		position := float64(i) / float64(barWidth) * 100.0

		var style lipgloss.Style
		if position >= 80 {
			style = lipgloss.NewStyle().Foreground(colorGreen)
		} else if position >= 60 {
			// Transition zone from yellow-green
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ACD32")) // yellow-green
		} else if position >= 40 {
			style = lipgloss.NewStyle().Foreground(colorYellow)
		} else if position >= 20 {
			// Transition zone from yellow to red
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")) // dark orange
		} else {
			style = lipgloss.NewStyle().Foreground(colorRed)
		}

		bar += style.Render("█")
	}
	for i := 0; i < empty; i++ {
		bar += lipgloss.NewStyle().Foreground(colorDim).Render("░")
	}

	scoreText := scoreStyle(score).Render(fmt.Sprintf("%.0f", score))
	scoreLabel := lipgloss.NewStyle().Foreground(colorDim).Render("/100")

	delta := ""
	if m.score.Delta > 0 {
		delta = scoreDeltaUpStyle.Render(fmt.Sprintf(" ▲ +%.1f", m.score.Delta))
	} else if m.score.Delta < 0 {
		delta = scoreDeltaDownStyle.Render(fmt.Sprintf(" ▼ %.1f", m.score.Delta))
	}

	spinner := ""
	if m.animating {
		spinner = " " + m.spinner.View()
	}

	line := fmt.Sprintf("  %s  %s%s%s%s", bar, scoreText, scoreLabel, delta, spinner)

	// Add sparkline if available
	var content string
	if m.sparklineData != nil && len(m.sparklineData.HealthScore) > 0 {
		spark := sparkline(m.sparklineData.HealthScore)
		sparkLabel := lipgloss.NewStyle().Foreground(colorDim).Render(" trend: ")
		content = line + "\n  " + sparkLabel + spark
	} else {
		content = line
	}

	width := m.width
	style := panelStyle.Width(width - 4)
	return style.Render(content)
}

func (m *model) viewComplexity() string {
	halfWidth := (m.width - 4) / 2
	style := panelStyle.Width(halfWidth)

	title := panelTitleStyle.Render("COMPLEXITY")

	var lines []string
	lines = append(lines, title)

	// Add sparkline if available
	if m.sparklineData != nil && len(m.sparklineData.AvgComplexity) > 0 {
		spark := sparkline(m.sparklineData.AvgComplexity)
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("  "+spark))
	}

	count := 8
	if len(m.results.Complexity) < count {
		count = len(m.results.Complexity)
	}

	if count == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("  No functions found"))
	}

	maxComplexity := 30
	for i := 0; i < count; i++ {
		fc := m.results.Complexity[i]

		var icon string
		if fc.Complexity > 20 {
			icon = statusBad.String()
		} else if fc.Complexity > 10 {
			icon = statusWarn.String()
		} else {
			icon = statusOK.String()
		}

		name := truncate(fc.Name, 18)
		loc := fmt.Sprintf("%s:%d", fc.File, fc.Line)
		loc = truncate(loc, 16)

		bar := complexityBar(fc.Complexity, maxComplexity)

		line := fmt.Sprintf("  %s %-18s %3d %s", icon, name, fc.Complexity, bar)
		lines = append(lines, line)
	}

	focusStyle := style
	if m.focus == panelComplexity {
		focusStyle = style.BorderForeground(colorCyan)
	}

	return focusStyle.Render(strings.Join(lines, "\n"))
}

func (m *model) viewDeps() string {
	halfWidth := (m.width - 4) / 2
	style := panelStyle.Width(halfWidth)

	title := panelTitleStyle.Render("DEPENDENCIES")

	var lines []string
	lines = append(lines, title)

	if len(m.results.Dependencies) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("  No dependencies found"))
	}

	count := 8
	if len(m.results.Dependencies) < count {
		count = len(m.results.Dependencies)
	}

	for i := 0; i < count; i++ {
		dep := m.results.Dependencies[i]

		var icon string
		var staleText string
		switch dep.Status {
		case "current":
			icon = statusOK.String()
			staleText = lipgloss.NewStyle().Foreground(colorGreen).Render("current")
		case "stale":
			icon = statusWarn.String()
			staleText = lipgloss.NewStyle().Foreground(colorYellow).Render(fmt.Sprintf("%dd old", dep.StaleDays))
		case "outdated":
			icon = statusBad.String()
			staleText = lipgloss.NewStyle().Foreground(colorRed).Render(fmt.Sprintf("%dd old", dep.StaleDays))
		default:
			icon = lipgloss.NewStyle().Foreground(colorDim).Render("?")
			staleText = lipgloss.NewStyle().Foreground(colorDim).Render("unknown")
		}

		name := truncate(dep.Module, 18)
		ver := truncate(dep.CurrentVersion, 10)

		line := fmt.Sprintf("  %s %-18s %-10s %s", icon, name, ver, staleText)
		lines = append(lines, line)
	}

	focusStyle := style
	if m.focus == panelDeps {
		focusStyle = style.BorderForeground(colorCyan)
	}

	return focusStyle.Render(strings.Join(lines, "\n"))
}

func (m *model) viewBoundaries() string {
	halfWidth := (m.width - 4) / 2
	style := panelStyle.Width(halfWidth)

	title := panelTitleStyle.Render("BOUNDARIES")

	var lines []string
	lines = append(lines, title)

	// Add sparkline if available
	if m.sparklineData != nil && len(m.sparklineData.ViolationCount) > 0 {
		spark := sparkline(m.sparklineData.ViolationCount)
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("  "+spark))
	}

	if len(m.cfg.Boundaries) == 0 && len(m.results.Violations) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("  No boundary rules defined"))
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("  Add rules in .drift.yaml"))
	}

	if len(m.results.Violations) > 0 {
		for _, v := range m.results.Violations {
			line := fmt.Sprintf("  %s %s → %s (%s:%d)",
				statusBad.String(),
				v.From, v.To,
				v.File, v.Line,
			)
			lines = append(lines, line)
		}
	} else if len(m.cfg.Boundaries) > 0 {
		lines = append(lines, fmt.Sprintf("  %s All boundaries clean", statusOK.String()))
	}

	focusStyle := style
	if m.focus == panelBoundaries {
		focusStyle = style.BorderForeground(colorCyan)
	}

	return focusStyle.Render(strings.Join(lines, "\n"))
}

func (m *model) viewActivity() string {
	halfWidth := (m.width - 4) / 2
	style := panelStyle.Width(halfWidth)

	title := panelTitleStyle.Render("ACTIVITY")

	var lines []string
	lines = append(lines, title)

	// Add dead code sparkline if available
	if m.sparklineData != nil && len(m.sparklineData.DeadCodeCount) > 0 {
		spark := sparkline(m.sparklineData.DeadCodeCount)
		sparkLabel := lipgloss.NewStyle().Foreground(colorDim).Render("  dead code: ")
		lines = append(lines, sparkLabel+spark)
	}

	if len(m.activity) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("  Watching for changes..."))
		langHint := string(m.results.Language)
		if langHint == "" {
			langHint = "a source"
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("  Edit "+langHint+" file to see updates"))
	}

	count := 6
	if len(m.activity) < count {
		count = len(m.activity)
	}

	for i := 0; i < count; i++ {
		entry := m.activity[i]
		ts := activityTimeStyle.Render(entry.timestamp.Format("15:04:05"))
		file := activityFileStyle.Render(filepath.Base(entry.file))
		line := fmt.Sprintf("  %s  %s modified", ts, file)
		lines = append(lines, line)
	}

	focusStyle := style
	if m.focus == panelActivity {
		focusStyle = style.BorderForeground(colorCyan)
	}

	return focusStyle.Render(strings.Join(lines, "\n"))
}

func (m *model) viewFooter() string {
	keys := []struct{ key, desc string }{
		{"tab", "navigate"},
		{"d", "diagnose"},
		{"r", "refresh"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		part := footerKeyStyle.Render("["+k.key+"]") + " " + k.desc
		parts = append(parts, part)
	}

	footer := strings.Join(parts, "  ")
	return footerStyle.Width(m.width).Render(footer)
}

func (m *model) viewDiagnosis() string {
	if m.diagnosing {
		content := diagnosisTitleStyle.Render("AI DIAGNOSIS") + "\n\n" +
			m.spinner.View() + " Analyzing codebase..."

		style := diagnosisStyle.Width(m.width - 8).Height(m.height - 6)
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			style.Render(content),
		)
	}

	content := diagnosisTitleStyle.Render("◆ AI DIAGNOSIS") + "\n\n" + m.diagnosisText +
		"\n\n" + footerKeyStyle.Render("[esc]") + " close"

	style := diagnosisStyle.Width(m.width - 8).Height(m.height - 6)
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		style.Render(content),
	)
}

func (m *model) listenForChanges() tea.Cmd {
	return func() tea.Msg {
		if m.watch == nil {
			select {}
		}
		event := <-m.watch.Events
		return fileChangedMsg{
			path:      event.Path,
			timestamp: event.Timestamp,
		}
	}
}

func (m *model) runAnalysis() tea.Cmd {
	return func() tea.Msg {
		results, err := m.ana.Run()
		if err != nil {
			return nil
		}
		score := m.scorer.Calculate(results)
		return analysisCompleteMsg{results: results, score: score}
	}
}

func (m *model) runDiagnosis() tea.Cmd {
	return func() tea.Msg {
		result, err := ai.RunDiagnosis(m.cfg, m.score, m.results)
		if err != nil {
			result = buildDiagnosisText(m.score, m.results, m.cfg)
			result = lipgloss.NewStyle().Foreground(colorDim).Render("(AI unavailable, showing local analysis)") + "\n\n" + result
		} else {
			result = lipgloss.NewStyle().Foreground(colorPurple).Render("Powered by "+m.cfg.AI.Provider) + "\n\n" + result
		}
		return diagnosisCompleteMsg{text: result}
	}
}

func (m *model) animateTick() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return animateTickMsg{}
	})
}

func (m *model) loadHistory() tea.Cmd {
	return func() tea.Msg {
		histAna, err := history.New(m.cfg)
		if err != nil {
			// Not a git repo or error loading, skip history
			return historyCompleteMsg{data: &history.SparklineData{}}
		}

		data, err := histAna.Walk(10)
		if err != nil {
			// Error walking commits, skip history
			return historyCompleteMsg{data: &history.SparklineData{}}
		}

		return historyCompleteMsg{data: data}
	}
}

func buildDiagnosisText(score health.Score, results *analyzer.Results, cfg *config.Config) string {
	var lines []string

	lines = append(lines, fmt.Sprintf("Health Score: %.0f/100\n", score.Total))

	if len(results.Complexity) > 0 {
		lines = append(lines, "Top Complex Functions:")
		count := 3
		if len(results.Complexity) < count {
			count = len(results.Complexity)
		}
		for i := 0; i < count; i++ {
			fc := results.Complexity[i]
			severity := "moderate"
			if fc.Complexity > 20 {
				severity = "high"
			}
			lines = append(lines, fmt.Sprintf("  %d. %s() in %s (complexity: %d, severity: %s)",
				i+1, fc.Name, fc.File, fc.Complexity, severity))
		}
		lines = append(lines, "")
	}

	staleCount := 0
	for _, dep := range results.Dependencies {
		if dep.Status != "current" {
			staleCount++
		}
	}
	if staleCount > 0 {
		lines = append(lines, fmt.Sprintf("Stale Dependencies: %d dependency(ies) behind latest", staleCount))
		for _, dep := range results.Dependencies {
			if dep.Status == "outdated" {
				lines = append(lines, fmt.Sprintf("  · %s is %dd behind (current: %s, latest: %s)",
					dep.Module, dep.StaleDays, dep.CurrentVersion, dep.LatestVersion))
			}
		}
		lines = append(lines, "")
	}

	if len(results.Violations) > 0 {
		lines = append(lines, fmt.Sprintf("Boundary Violations: %d import(s) cross architectural boundaries", len(results.Violations)))
		for _, v := range results.Violations {
			lines = append(lines, fmt.Sprintf("  · %s imports from %s (%s:%d)", v.From, v.To, v.File, v.Line))
		}
	}

	if len(lines) == 1 {
		lines = append(lines, "Your codebase looks healthy! No major issues detected.")
	}

	return strings.Join(lines, "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func PrintReport(cfg *config.Config, score health.Score, results *analyzer.Results) {
	fmt.Println()
	fmt.Println(logoStyle.Render("◆ DRIFT REPORT"))
	fmt.Println()

	fmt.Printf("  Health Score: %s\n", scoreStyle(score.Total).Render(fmt.Sprintf("%.0f/100", score.Total)))
	fmt.Println()

	fmt.Println(panelTitleStyle.Render("  COMPLEXITY"))
	count := 10
	if len(results.Complexity) < count {
		count = len(results.Complexity)
	}
	for i := 0; i < count; i++ {
		fc := results.Complexity[i]
		fmt.Printf("    %s %s:%d %s() — complexity %d\n",
			func() string {
				if fc.Complexity > 20 {
					return statusBad.String()
				} else if fc.Complexity > 10 {
					return statusWarn.String()
				}
				return statusOK.String()
			}(),
			fc.File, fc.Line, fc.Name, fc.Complexity)
	}
	fmt.Println()

	fmt.Println(panelTitleStyle.Render("  DEPENDENCIES"))
	for _, dep := range results.Dependencies {
		icon := statusOK.String()
		if dep.Status == "stale" {
			icon = statusWarn.String()
		} else if dep.Status == "outdated" {
			icon = statusBad.String()
		}
		fmt.Printf("    %s %-20s %s → %s\n", icon, dep.Module, dep.CurrentVersion, dep.LatestVersion)
	}
	fmt.Println()

	if len(results.Violations) > 0 {
		fmt.Println(panelTitleStyle.Render("  BOUNDARY VIOLATIONS"))
		for _, v := range results.Violations {
			fmt.Printf("    %s %s → %s (%s:%d)\n", statusBad.String(), v.From, v.To, v.File, v.Line)
		}
		fmt.Println()
	}
}

func PrintSnapshot(score health.Score, results *analyzer.Results) error {
	snapshot := map[string]interface{}{
		"language": string(results.Language),
		"score": map[string]interface{}{
			"total":      score.Total,
			"complexity": score.Complexity,
			"deps":       score.Deps,
			"boundaries": score.Boundaries,
		},
		"summary": map[string]interface{}{
			"files":      results.FileCount,
			"functions":  results.FuncCount,
			"violations": len(results.Violations),
			"deps":       len(results.Dependencies),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(snapshot)
}
