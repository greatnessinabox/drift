# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in drift, please report it responsibly.

**Do not open a public issue for security vulnerabilities.**

Instead, email the maintainer directly or open a private security advisory on GitHub.

## Scope

drift is a local analysis tool. It:

- Reads source files from your local filesystem (read-only)
- Makes outbound HTTPS requests to package registries (npm, PyPI, crates.io, Maven Central, Go proxy) to check dependency versions
- Optionally calls AI APIs (Anthropic, OpenAI) if configured with API keys via environment variables
- Does not transmit your source code anywhere

## API Keys

AI diagnostic features require API keys set via environment variables (`ANTHROPIC_API_KEY` or `OPENAI_API_KEY`). These keys are never written to disk or logged.

## Supported Versions

Only the latest release is supported with security updates.
