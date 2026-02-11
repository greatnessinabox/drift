# AI Agent Integration Guide

drift works seamlessly with popular AI coding assistants. While built for the **GitHub Copilot CLI Challenge**, drift's architecture (`.github/copilot-instructions.md`, structured JSON output, and MCP compatibility) makes it work with any AI agent.

## 🤖 Popular AI Coding Assistants

### 1. GitHub Copilot ⭐ (Featured)

**Official GitHub AI platform** with IDE extensions, CLI, and autonomous agents.

**Products:**
- Copilot in IDE (VS Code, JetBrains, etc.)
- Copilot CLI (terminal commands)
- Copilot Agents (autonomous coding on GitHub)
- Copilot Spaces (shared team knowledge)

**drift Integration:**
- Custom agent: `.github/agents/drift-dev.agent.md`
- Interactive fixing: `drift fix` command
- CI/CD: GitHub Actions with AI-generated PR comments

**Usage:**
```bash
# Interactive fixing
drift fix

# Custom agent (requires Copilot CLI: brew install copilot-cli)
copilot --agent drift-dev "analyze src/"
```

**Pricing:** Free tier, Pro ($10/mo), Business/Enterprise  
**URL:** https://github.com/features/copilot

---

### 2. Claude Code 🎯 (Anthropic Official)

**Anthropic's official agentic coding tool** - works in terminal, IDE, browser, and desktop app.

**Key Features:**
- Multi-surface: CLI, VS Code, JetBrains, Desktop, Web, Mobile
- Reads codebase, edits files, runs commands
- Interactive diff viewing
- MCP support
- CLAUDE.md files for project instructions

**drift Integration:**
- Reads `.github/copilot-instructions.md` for architecture context
- Executes `drift` commands with approval
- Parses drift JSON output

**Usage:**
```bash
# Install
curl -fsSL https://claude.ai/install.sh | bash

# Use in your project
cd your-project
claude "Run drift report and refactor the high-complexity functions 
following drift's architecture patterns"
```

**In VS Code:** Install "Claude Code" extension, use @-mentions for context.

**Pricing:** Free tier, Pro ($20/mo), Team ($30/user/mo)  
**URL:** https://code.claude.com

---

### 3. Cursor 🔥 (AI-First Editor)

**AI-native code editor** (fork of VSCode) used by Stripe, OpenAI, Nvidia, and thousands of companies.

**Key Features:**
- Multiple LLMs: Claude, GPT-4, Gemini, Grok, Cursor models
- Tab completion, Cmd+K edits, Agent mode
- Built-in terminal and debugging
- Codebase-wide understanding

**drift Integration:**
- Reads `.github/copilot-instructions.md` automatically
- Understands drift JSON output
- Agent mode can execute drift

**Usage:**
```
Composer Mode (Cmd/Ctrl+I):
@Docs .github/copilot-instructions.md

drift shows high complexity in model.Update() (25).
Refactor to <10 following drift's patterns.
```

**Pricing:** Hobby ($20/mo), Pro ($20/mo), Business ($40/user/mo)  
**URL:** https://www.cursor.com

---

### 4. Cline 🔧 (Community VSCode Extension)

**Popular community VSCode extension** using Claude Sonnet for agentic coding (not official Anthropic).

**Key Features:**
- Human-in-the-loop approval for file changes
- Terminal command execution (with permission)
- Browser automation for web dev
- MCP support

**drift Integration:**
- Reads `.github/copilot-instructions.md` when added to workspace
- Executes drift commands with approval
- Understands project structure

**Usage in VSCode:**
```
@workspace Read .github/copilot-instructions.md

Run `drift report`.

Based on the output, refactor the top 3 high-complexity 
functions to reduce complexity below 10.
```

**Pricing:** Free (uses your Claude API key)  
**URL:** https://github.com/cline/cline

---

### 5. Aider 💬 (CLI Tool)

**AI pair programming in your terminal** - works with any LLM.

**Key Features:**
- Best with Claude 3.7 Sonnet, DeepSeek R1/V3, GPT-4o
- Supports local models (Ollama, etc.)
- 100+ programming languages
- Git integration (auto-commits)

**drift Integration:**
- Reads `.github/copilot-instructions.md` when added to chat
- Executes drift commands and sees output
- Perfect for batch refactoring

**Usage:**
```bash
# Install
pip install aider-chat

# Use with drift
aider --model claude-3-7-sonnet-20250219 \
      --message "Read .github/copilot-instructions.md" \
      internal/tui/app.go

# In aider:
> Run drift report and show functions with complexity >15
> Refactor to reduce below 10 following drift's patterns
```

**Pricing:** Free & open source (you pay for LLM API)  
**URL:** https://aider.chat

---

## 🌐 Model Context Protocol (MCP)

**MCP** is the open standard (February 2026) for connecting AI models to tools and data sources.

**Supported by:** VS Code, JetBrains, Cursor, Claude Code, and more.

**drift & MCP:** drift's architecture is MCP-compatible. The `.github/copilot-instructions.md` file and structured JSON output work as MCP context.

**More info:** https://modelcontextprotocol.io

---

## 🎯 Quick Comparison

| Tool | Type | Best For | Pricing | Official |
|------|------|----------|---------|----------|
| **GitHub Copilot** | Platform | Enterprise, GitHub workflows | $10+/mo | ✅ GitHub |
| **Claude Code** | Multi-surface | Terminal + IDE agentic coding | $20+/mo | ✅ Anthropic |
| **Cursor** | Editor | AI-first development | $20+/mo | ✅ Cursor |
| **Cline** | VSCode Ext | Approval-based agentic coding | Free + API | ❌ Community |
| **Aider** | CLI | Terminal, batch refactoring | Free + API | ❌ Open Source |

---

## 🚀 Universal Integration Pattern

All AI assistants can use drift's **three integration points**:

### 1. Architecture Context
**File:** `.github/copilot-instructions.md`

Contains drift's architecture, LanguageAnalyzer interface, and development patterns.

**Usage:** Add to your AI's context when making drift-related changes.

### 2. Structured Output
```bash
drift report   # Human-readable
drift snapshot # JSON for AI parsing
```

**Usage:** Share with AI for structured refactoring suggestions.

### 3. Interactive Analysis
```bash
drift         # Live TUI with real-time metrics
drift fix     # Interactive fixing with Copilot (if installed)
```

---

## 💡 Recommended Workflow

Works with any AI assistant:

```
1. Run: drift report
2. Identify: High-complexity functions (🔴 HIGH)
3. Share: Output + .github/copilot-instructions.md with AI
4. Request: "Refactor to complexity <10, maintain tests"
5. Review: AI's proposed changes
6. Apply: Accept/reject changes
7. Verify: drift report (confirm improvement)
8. Commit: If health score improved
```

---

## 🆕 What's Current (February 2026)

- **MCP** is the standard for AI tool integration
- **Claude Code** is Anthropic's official agentic coding tool (CLI, IDE, web, desktop)
- **Cline** (formerly "Claude Dev") is the popular community VSCode extension
- **Cursor** supports multiple LLMs and has agent mode
- **GitHub Copilot** has Spaces and autonomous agents

---

## 🤝 Contributing

Found a better workflow? Using drift with a different AI assistant?

- Open an issue to share your experience
- Submit a PR with improvements
- Help us make drift work better with more tools

drift is designed to be AI-agnostic. Help us expand support! 🚀
