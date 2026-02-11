# Scripts

Automation scripts for drift development and demos.

## Demo Generation

### generate-demos.sh

Automatically generates all demo GIFs using VHS.

**Usage:**
```bash
./scripts/generate-demos.sh
```

**Requirements:**
- VHS installed (`brew install vhs`)
- VHS tape files in project root (demo*.tape)

**Generates:**
- `demo-quick.gif` - 30s quick demo
- `demo-tui.gif` - 50s TUI walkthrough
- `demo-dashboard.gif` - 40s dashboard showcase
- `demo.gif` - 60s full demo

**Alternative:**
Use Make targets:
```bash
make demos           # Generate all
make demo-quick      # Just quick demo
make demo-clean      # Remove all demos
```

## Benefits

- **100% Automated** - No manual recording needed
- **Consistent Quality** - Same output every time
- **Reproducible** - Can regenerate anytime
- **Fast** - ~2-3 minutes for all demos
- **Professional** - VHS produces beautiful GIFs

## For Challenge Submission

1. Run: `./scripts/generate-demos.sh`
2. Review generated GIFs
3. Upload to DEV.to blog post
4. Include in README.md
5. Submit!

🎉 Zero manual work required!
