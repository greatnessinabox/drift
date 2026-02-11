package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	colorGreen   = lipgloss.Color("#00FF87")
	colorYellow  = lipgloss.Color("#FFD700")
	colorRed     = lipgloss.Color("#FF6B6B")
	colorCyan    = lipgloss.Color("#00E5FF")
	colorDim     = lipgloss.Color("#666666")
	colorWhite   = lipgloss.Color("#FAFAFA")
	colorBg      = lipgloss.Color("#1A1A2E")
	colorPanel   = lipgloss.Color("#16213E")
	colorAccent  = lipgloss.Color("#E94560")
	colorPurple  = lipgloss.Color("#A855F7")
	colorBorder  = lipgloss.Color("#333366")

	// Title
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan).
			Align(lipgloss.Center)

	// Score
	scoreHighStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen)

	scoreMedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorYellow)

	scoreLowStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorRed)

	scoreDeltaUpStyle = lipgloss.NewStyle().
				Foreground(colorGreen)

	scoreDeltaDownStyle = lipgloss.NewStyle().
				Foreground(colorRed)

	// Panels
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan).
			MarginBottom(1)

	// Status indicators
	statusOK   = lipgloss.NewStyle().Foreground(colorGreen).SetString("✓")
	statusWarn = lipgloss.NewStyle().Foreground(colorYellow).SetString("⚠")
	statusBad  = lipgloss.NewStyle().Foreground(colorRed).SetString("✗")

	// Activity feed
	activityTimeStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	activityFileStyle = lipgloss.NewStyle().
				Foreground(colorWhite)

	// Footer
	footerStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Align(lipgloss.Center)

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	// AI Diagnosis
	diagnosisStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2)

	diagnosisTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPurple)

	// Complexity bar
	complexityBarFull  = lipgloss.NewStyle().Foreground(colorRed).SetString("█")
	complexityBarWarn  = lipgloss.NewStyle().Foreground(colorYellow).SetString("█")
	complexityBarGood  = lipgloss.NewStyle().Foreground(colorGreen).SetString("█")
	complexityBarEmpty = lipgloss.NewStyle().Foreground(colorDim).SetString("░")

	// Logo
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)
)

func scoreStyle(score float64) lipgloss.Style {
	if score >= 80 {
		return scoreHighStyle
	}
	if score >= 50 {
		return scoreMedStyle
	}
	return scoreLowStyle
}

func complexityBar(complexity, maxDisplay int) string {
	if maxDisplay == 0 {
		maxDisplay = 30
	}
	barLen := 15
	filled := complexity * barLen / maxDisplay
	if filled > barLen {
		filled = barLen
	}

	bar := ""
	for i := 0; i < filled; i++ {
		if complexity > 20 {
			bar += complexityBarFull.String()
		} else if complexity > 10 {
			bar += complexityBarWarn.String()
		} else {
			bar += complexityBarGood.String()
		}
	}
	for i := filled; i < barLen; i++ {
		bar += complexityBarEmpty.String()
	}
	return bar
}
