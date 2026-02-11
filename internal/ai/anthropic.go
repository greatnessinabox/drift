package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/greatnessinabox/drift/internal/config"
)

type AnthropicProvider struct {
	client *anthropic.Client
	model  anthropic.Model
}

func NewAnthropicProvider(cfg config.AIConfig) (*AnthropicProvider, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	client := anthropic.NewClient()

	model := anthropic.Model(cfg.Model)
	if cfg.Model == "" {
		model = anthropic.ModelClaudeSonnet4_5_20250929
	}

	return &AnthropicProvider{
		client: &client,
		model:  model,
	}, nil
}

func (p *AnthropicProvider) Name() string {
	return "Anthropic Claude"
}

func (p *AnthropicProvider) Diagnose(ctx context.Context, prompt string) (string, error) {
	resp, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		System: []anthropic.TextBlockParam{
			{
				Text: "You are a code health analyst. Analyze the codebase metrics provided and give actionable, concise recommendations. Focus on the most impactful issues first. Be specific about file names and function names. Keep your response under 500 words.",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic API error: %w", err)
	}

	for _, block := range resp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("no text response from Anthropic")
}
