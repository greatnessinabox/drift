# VHS tape for drift demo - Comprehensive (60 seconds)
# Shows all commands and multi-language support
# Install VHS: https://github.com/charmbracelet/vhs
# Run with: vhs demo.tape

Output demo.gif

Set FontSize 16
Set Width 1400
Set Height 900
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms
Set PlaybackSpeed 0.8

# Build drift (hidden)
Hide
Type "go build ./cmd/drift/"
Enter
Sleep 3s
Show

# Title screen
Type "# drift - Real-time codebase health dashboard"
Sleep 1s
Enter
Sleep 500ms
Type "# Supports Go, TypeScript, Python, Rust, and Java"
Sleep 2s
Enter
Enter

# Show help
Type "./drift --help"
Sleep 800ms
Enter
Sleep 4s

# Clear screen
Ctrl+L
Sleep 500ms

# Show report command
Type "# Generate a health report"
Sleep 800ms
Enter
Type "./drift report"
Sleep 800ms
Enter
Sleep 5s

# Clear screen
Ctrl+L
Sleep 500ms

# Show check command for CI
Type "# CI health gate - passes"
Sleep 800ms
Enter
Type "./drift check --fail-under 70"
Sleep 800ms
Enter
Sleep 3s

# Clear screen
Ctrl+L
Sleep 500ms

# Show check with failing threshold
Type "# CI health gate - fails"
Sleep 800ms
Enter
Type "./drift check --fail-under 95"
Sleep 800ms
Enter
Sleep 3s

# Clear screen
Ctrl+L
Sleep 500ms

# Show JSON snapshot
Type "# JSON snapshot for pipelines"
Sleep 800ms
Enter
Type "./drift snapshot"
Sleep 800ms
Enter
Sleep 4s

# Clear screen
Ctrl+L
Sleep 500ms

# Show the live dashboard
Type "# Launch the live TUI dashboard"
Sleep 800ms
Enter
Type "./drift"
Sleep 800ms
Enter

# Let the dashboard load and sparklines render
Sleep 5s

# Let the dashboard display
Sleep 2s

# Tab through panels
Tab
Sleep 2s
Tab
Sleep 2s
Tab
Sleep 2s
Tab
Sleep 2s

# Show refresh
Type "r"
Sleep 3s

# Exit the dashboard
Type "q"
Sleep 1s

# Clear and show final message
Ctrl+L
Sleep 500ms
Type "# go install github.com/greatnessinabox/drift/cmd/drift@latest"
Sleep 3s
