# VHS Recording Instructions for drift

This directory contains a VHS tape file for creating an automated demo of drift.

## Prerequisites

Install VHS (Video-based Terminal Emulator):

```bash
# macOS
brew install vhs

# Linux
# Install from https://github.com/charmbracelet/vhs
```

## Recording the Demo

```bash
cd /path/to/drift
vhs demo.tape
```

This will generate `demo.gif` showing:
- ✨ drift's help menu
- 📊 Terminal health report
- ✅ CI check command with pass/fail scenarios
- 🎨 Live TUI dashboard with sparklines
- ⌨️  Panel navigation
- 🔄 Refresh functionality

## Customization

Edit `demo.tape` to customize:
- Terminal size: `Set Width/Height`
- Colors: `Set Theme`
- Typing speed: `Set TypingSpeed`
- Output format: Change `Output demo.gif` to `.mp4`, `.webm`, etc.

## Tips

- Test the demo first with a shorter version
- Ensure drift binary is in PATH before recording
- Use `Sleep` commands to control pacing
- Keep the demo under 60 seconds for README/docs

## Example Usage in README

Once recorded, embed in README.md:

```markdown
![drift demo](demo.gif)
```

Or host on GitHub and reference:

```markdown
![drift demo](https://github.com/greatnessinabox/drift/raw/main/demo.gif)
```
