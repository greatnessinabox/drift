// Package ai abstracts AI providers and builds prompts to diagnose codebase health issues.
package ai

import (
	"context"
	"fmt"

	"github.com/greatnessinabox/drift/internal/config"
)

type Provider interface {
	Diagnose(ctx context.Context, prompt string) (string, error)
	Name() string
}

func NewProvider(cfg config.AIConfig) (Provider, error) {
	switch cfg.Provider {
	case "anthropic":
		return NewAnthropicProvider(cfg)
	case "openai":
		return NewOpenAIProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown AI provider: %q (supported: anthropic, openai)", cfg.Provider)
	}
}

// maxTokensOrDefault falls back to a sane response budget when none is configured.
func maxTokensOrDefault(n int) int64 {
	if n <= 0 {
		return 1024
	}
	return int64(n)
}
