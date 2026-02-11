# Recording Demos with VHS

This repository includes VHS tape files for creating automated terminal demos of drift.

## Prerequisites

1. Install VHS (https://github.com/charmbracelet/vhs):

```bash
# macOS
brew install vhs
```

2. Build drift:

```bash
go build ./cmd/drift/
```

All tape files include a hidden build step, but having drift pre-built speeds up recording.

## Available Demos

### 1. `demo-quick.tape` (30 seconds)
**Best for:** README header, social media, quick previews

Shows:
- Quick report output
- CI check command
- JSON snapshot
- Installation instructions

```bash
vhs demo-quick.tape
# Generates: demo-quick.gif
```

### 2. `demo-dashboard.tape` (40 seconds)
**Best for:** Showcasing the live TUI

Shows:
- Full TUI dashboard with sparklines
- Panel navigation
- Refresh animation
- Real-time monitoring

```bash
vhs demo-dashboard.tape
# Generates: demo-dashboard.gif
```

### 3. `demo-tui.tape` (50 seconds)
**Best for:** Complete feature walkthrough

Shows:
- Report command
- CI integration
- Live TUI with all features
- Navigation and interaction

```bash
vhs demo-tui.tape
# Generates: demo-tui.gif
```

### 4. `demo.tape` (60 seconds)
**Best for:** Comprehensive demo with all commands

Shows:
- Help menu
- All CLI commands (report, check pass/fail, snapshot)
- TUI dashboard with navigation
- Complete workflow

```bash
vhs demo.tape
# Generates: demo.gif
```

## Recording

```bash
vhs demo-quick.tape
```

The GIF will be generated in the current directory. Each tape file includes a hidden `go build` step so the binary is always fresh.

## Customization

Edit any `.tape` file to customize:

- **Output format**: Change `Output demo.gif` to `.mp4`, `.webm`, etc.
- **Terminal size**: Adjust `Set Width` and `Set Height`
- **Theme**: Change `Set Theme` (see VHS docs for available themes)
- **Speed**: Adjust `Set TypingSpeed` and `Set PlaybackSpeed`
- **Timing**: Modify `Sleep` durations between commands

## Tips

- **Optimize size**: Keep GIFs under 5MB for GitHub README embeds
- **Loop**: VHS GIFs loop by default - design accordingly
- **Compress**: Use `gifsicle` to reduce file size if needed
