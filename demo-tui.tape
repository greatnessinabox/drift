# VHS tape for drift demo - Interactive version
# This version focuses on the live TUI experience
# Run with: vhs demo-tui.tape

Output demo-tui.gif

Set FontSize 14
Set Width 1600
Set Height 1000
Set Theme "Dracula"
Set TypingSpeed 80ms
Set PlaybackSpeed 1.0

Hide

# Navigate to drift directory and ensure we're in the right place
Type "cd /Users/marquis.nobles/marquis.codes/drift"
Enter
Sleep 500ms

Show

# Title
Type "# drift - Real-time Go codebase health monitoring"
Sleep 1.5s
Enter
Sleep 300ms

# Quick feature list
Type "# ✨ Features: Live TUI • AI Diagnostics • Sparkline Trends • CI Integration"
Sleep 2s
Enter
Enter

# Show quick report first
Type "drift report"
Sleep 300ms
Enter
Sleep 3.5s

# Clear
Ctrl+L

# Show CI check
Type "# CI Integration - check health score"
Sleep 500ms
Enter
Type "drift check --fail-under 90"
Sleep 300ms
Enter
Sleep 2.5s

# Clear for TUI
Ctrl+L

# Launch TUI
Type "# Live TUI Dashboard with real-time monitoring"
Sleep 500ms
Enter
Type "drift"
Sleep 500ms
Enter

# Let TUI load and show sparklines
Sleep 3s

# Navigate through panels to show focus
Sleep 1s
Tab
Sleep 1.5s
Tab
Sleep 1.5s
Tab
Sleep 1.5s
Tab
Sleep 1.5s

# Back to first panel
Shift+Tab
Sleep 1s

# Refresh
Type "r"
Sleep 2s

# Show footer hints
Sleep 2s

# Exit
Type "q"
Sleep 500ms

# Final screen
Ctrl+L
Type "# Get started:"
Enter
Type "go install github.com/greatnessinabox/drift@latest"
Sleep 2s
Enter
Type "cd your-go-project && drift"
Sleep 2s
