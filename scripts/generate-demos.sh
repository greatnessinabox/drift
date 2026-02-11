#!/usr/bin/env bash
# Automated demo generation script for drift
# Generates all demo GIFs/videos for GitHub Copilot CLI Challenge submission

set -e

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                                                               ║"
echo "║  🎬 drift Demo Generation Script                             ║"
echo "║                                                               ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

# Check if VHS is installed
if ! command -v vhs &> /dev/null; then
    echo "❌ VHS is not installed!"
    echo ""
    echo "Install VHS:"
    echo "  macOS:   brew install vhs"
    echo "  Linux:   See https://github.com/charmbracelet/vhs#installation"
    echo ""
    exit 1
fi

echo "✅ VHS is installed: $(vhs --version)"
echo ""

# Function to generate a demo
generate_demo() {
    local tape_file=$1
    local demo_name=$2
    local description=$3
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🎬 Generating: $demo_name"
    echo "   $description"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    if [ ! -f "$tape_file" ]; then
        echo "❌ Tape file not found: $tape_file"
        return 1
    fi
    
    echo "   Running: vhs $tape_file"
    if vhs "$tape_file"; then
        local output_file="${tape_file%.tape}.gif"
        if [ -f "$output_file" ]; then
            local size=$(du -h "$output_file" | cut -f1)
            echo ""
            echo "   ✅ Generated: $output_file ($size)"
        else
            echo ""
            echo "   ⚠️  Output file not found (may have used different name)"
        fi
    else
        echo ""
        echo "   ❌ Failed to generate $demo_name"
        return 1
    fi
    
    echo ""
}

# Generate all demos
echo "Starting demo generation..."
echo ""

generate_demo "demo-quick.tape" "Quick Demo (30s)" "Installation, report, CI check"
generate_demo "demo-tui.tape" "TUI Demo (50s)" "Full TUI walkthrough with all features"
generate_demo "demo-dashboard.tape" "Dashboard Demo (40s)" "Dashboard showcase with sparklines"
generate_demo "demo.tape" "Full Demo (60s)" "Complete feature demonstration"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Demo generation complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Generated files:"
ls -lh *.gif 2>/dev/null | awk '{print "  📹 " $9 " (" $5 ")"}'
echo ""
echo "Next steps:"
echo "  1. Review generated GIFs"
echo "  2. Use in README.md: ![Demo](demo-quick.gif)"
echo "  3. Upload to DEV.to blog post"
echo "  4. Include in challenge submission"
echo ""
echo "🎉 Ready for submission!"
