# VHS tape for drift demo - Full feature showcase (50 seconds)
# Shows multi-language support with both Go and TypeScript analysis
# Run with: vhs demo-tui.tape

Output demo-tui.gif

Set FontSize 14
Set Width 1600
Set Height 1000
Set Theme "Dracula"
Set TypingSpeed 80ms
Set PlaybackSpeed 1.0

# Build drift (hidden)
Hide
Type "go build ./cmd/drift/"
Enter
Sleep 3s
Show

# Title
Type "# drift - Polyglot codebase health monitoring"
Sleep 1.5s
Enter
Type "# Go, TypeScript, Python, Rust, Java"
Sleep 1.5s
Enter
Enter

# Show report on Go project (drift itself)
Type "./drift report"
Sleep 500ms
Enter
Sleep 4s

# Clear
Ctrl+L
Sleep 500ms

# Show CI check
Type "# CI Integration - check health score"
Sleep 800ms
Enter
Type "./drift check --fail-under 90"
Sleep 500ms
Enter
Sleep 3s

# Clear for TUI
Ctrl+L
Sleep 500ms

# Launch TUI
Type "# Live TUI Dashboard with real-time monitoring"
Sleep 800ms
Enter
Type "./drift"
Sleep 800ms
Enter

# Let TUI load and show sparklines
Sleep 4s

# Navigate through panels to show focus
Sleep 1.5s
Tab
Sleep 2s
Tab
Sleep 2s
Tab
Sleep 2s
Tab
Sleep 2s

# Back to first panel
Shift+Tab
Sleep 1.5s

# Refresh
Type "r"
Sleep 3s

# Show footer hints
Sleep 2s

# Exit
Type "q"
Sleep 1s

# Final screen
Ctrl+L
Sleep 500ms
Type "# go install github.com/greatnessinabox/drift/cmd/drift@latest"
Sleep 3s
