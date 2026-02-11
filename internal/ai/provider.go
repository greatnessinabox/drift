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
