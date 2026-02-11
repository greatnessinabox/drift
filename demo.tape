# VHS tape for drift demo
# Install VHS: https://github.com/charmbracelet/vhs
# Run with: vhs demo.tape

Output demo.gif

Set FontSize 16
Set Width 1400
Set Height 900
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms
Set PlaybackSpeed 0.8

# Title screen
Type "# drift - Real-time codebase health dashboard"
Sleep 1s
Enter
Sleep 500ms
Type "# Monitors Go code quality with AI diagnostics"
Sleep 2s
Enter
Enter

# Show help
Type "drift --help"
Sleep 500ms
Enter
Sleep 3s

# Clear screen
Ctrl+L

# Show report command
Type "# Generate a health report"
Sleep 500ms
Enter
Type "drift report"
Sleep 500ms
Enter
Sleep 4s

# Clear screen
Ctrl+L

# Show check command for CI
Type "# Check health score for CI pipelines"
Sleep 500ms
Enter
Type "drift check --fail-under 70"
Sleep 500ms
Enter
Sleep 2s

# Clear screen
Ctrl+L

# Show check with failing threshold
Type "# Fail if score is too low"
Sleep 500ms
Enter
Type "drift check --fail-under 95"
Sleep 500ms
Enter
Sleep 2s

# Clear screen
Ctrl+L

# Show the live dashboard
Type "# Launch the live TUI dashboard"
Sleep 500ms
Enter
Type "drift"
Sleep 500ms
Enter
Sleep 3s

# Let the dashboard display
Sleep 2s

# Show navigation hint
Sleep 1s

# Tab through panels
Sleep 1s
Type@500ms "Tab"
Sleep 1s
Type@500ms "Tab"
Sleep 1s
Type@500ms "Tab"
Sleep 1s

# Wait before showing refresh
Sleep 1s

# Show refresh
Type "r"
Sleep 2s

# Show diagnosis hint
Sleep 1s

# Exit the dashboard
Type "q"
Sleep 500ms

# Clear and show final message
Ctrl+L
Type "# Try drift in your Go projects!"
Sleep 500ms
Enter
Type "# go install github.com/greatnessinabox/drift@latest"
Sleep 3s
