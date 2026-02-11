# VHS tape - Showcase drift's live TUI with sparklines
# Best for showing the interactive dashboard
# Run with: vhs demo-dashboard.tape

Output demo-dashboard.gif

Set FontSize 15
Set Width 1600
Set Height 1000
Set Theme "Catppuccin Mocha"
Set TypingSpeed 100ms
Set PlaybackSpeed 1.0
Set Padding 20

# Build drift first (hidden)
Hide
Type "go build ./cmd/drift/"
Enter
Sleep 3s
Show

# Launch drift directly
Type "./drift"
Sleep 500ms
Enter

# Let the dashboard fully load with sparklines
Sleep 5s

# Show the sparklines are visible
Sleep 2s

# Navigate through panels - Complexity
Tab
Sleep 2.5s

# Navigate to Dependencies
Tab
Sleep 2.5s

# Navigate to Boundaries
Tab
Sleep 2.5s

# Navigate to Activity
Tab
Sleep 2.5s

# Back to Score panel
Shift+Tab
Sleep 500ms
Shift+Tab
Sleep 500ms
Shift+Tab
Sleep 500ms
Shift+Tab
Sleep 2s

# Trigger a refresh to show animation
Type "r"
Sleep 3s

# Let the score animation complete
Sleep 2s

# Show keyboard shortcuts in footer
Sleep 2s

# Exit gracefully
Type "q"
Sleep 1s
