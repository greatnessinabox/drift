# Recording Demos with VHS

This repository includes VHS tape files for creating automated terminal demos of drift.

## Available Demos

### 1. `demo-quick.tape` (30 seconds)
**Best for:** README header, social media, quick previews

Shows:
- Quick report output
- CI check command
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
- All CLI commands
- TUI dashboard
- Complete workflow

```bash
vhs demo.tape
# Generates: demo.gif
```

## Prerequisites

Install VHS (https://github.com/charmbracelet/vhs):

```bash
# macOS
brew install vhs

# Arch Linux
pacman -S vhs

# Nix
nix-env -iA nixpkgs.vhs

# From source
go install github.com/charmbracelet/vhs@latest
```

## Recording

1. Ensure drift binary is built and in the current directory:
   ```bash
   go build ./cmd/drift/
   ```

2. Run VHS with your chosen tape file:
   ```bash
   vhs demo-quick.tape
   ```

3. The GIF will be generated in the current directory.

## Customization

Edit any `.tape` file to customize:

- **Output format**: Change `Output demo.gif` to `.mp4`, `.webm`, etc.
- **Terminal size**: Adjust `Set Width` and `Set Height`
- **Theme**: Change `Set Theme` (see VHS docs for available themes)
- **Speed**: Adjust `Set TypingSpeed` and `Set PlaybackSpeed`
- **Timing**: Modify `Sleep` durations between commands

## Tips

- **Test first**: Run VHS with a short tape to verify setup
- **Clean state**: Clear terminal history before recording
- **Consistent state**: Use `Hide` blocks to set up environment
- **Optimize size**: Keep GIFs under 5MB for GitHub
- **Loop**: VHS GIFs loop by default - design accordingly

## Hosting Demos

### In README
```markdown
![drift demo](demo-quick.gif)
```

### On GitHub
```markdown
![drift demo](https://github.com/greatnessinabox/drift/raw/main/demo-quick.gif)
```

### On your website
Host the GIF and embed:
```html
<img src="drift-demo.gif" alt="drift demo" />
```

## Troubleshooting

### VHS not found
- Ensure VHS is installed: `which vhs`
- Add to PATH if needed

### Drift command not found
- Build drift: `go build ./cmd/drift/`
- Add `./` prefix in tape file if needed

### GIF too large
- Reduce dimensions: Lower `Width` and `Height`
- Reduce duration: Cut unnecessary `Sleep` commands
- Compress: Use tools like `gifsicle`

### Wrong directory
- Use `Hide` block to navigate: `Type "cd /path/to/drift"`
- Update paths in tape files to match your setup

## Examples

See `VHS_README.md` for additional usage examples and integration tips.
