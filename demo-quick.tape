# VHS tape for drift demo - Quick CLI demo (30 seconds)
# Perfect for README header or social media
# Run with: vhs demo-quick.tape

Output demo-quick.gif

Set FontSize 18
Set Width 1200
Set Height 700
Set Theme "GitHub Dark"
Set TypingSpeed 100ms
Set PlaybackSpeed 1.2

# Build drift first (hidden)
Hide
Type "go build ./cmd/drift/"
Enter
Sleep 3s
Show

# Quick intro
Type "# drift - Multi-language codebase health"
Sleep 1s
Enter
Enter

# Show report on a Go project
Type "./drift report"
Sleep 500ms
Enter
Sleep 4s

# Clear
Ctrl+L
Sleep 300ms

# Show check for CI
Type "./drift check --fail-under 70"
Sleep 500ms
Enter
Sleep 3s

# Clear
Ctrl+L
Sleep 300ms

# Show JSON snapshot
Type "./drift snapshot | head -10"
Sleep 500ms
Enter
Sleep 3s

# Clear and final message
Ctrl+L
Sleep 300ms
Type "# go install github.com/greatnessinabox/drift/cmd/drift@latest"
Sleep 3s
