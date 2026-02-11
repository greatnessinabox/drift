package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/greatnessinabox/drift/internal/config"
)

type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(cfg config.AIConfig) (*OpenAIProvider, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient()

	model := cfg.Model
	if model == "" {
		model = "gpt-4o"
	}

	return &OpenAIProvider{
		client: &client,
		model:  model,
	}, nil
}

func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

func (p *OpenAIProvider) Diagnose(ctx context.Context, prompt string) (string, error) {
	resp, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: p.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a code health analyst. Analyze the codebase metrics provided and give actionable, concise recommendations. Focus on the most impactful issues first. Be specific about file names and function names. Keep your response under 500 words."),
			openai.UserMessage(prompt),
		},
		MaxCompletionTokens: openai.Int(1024),
	})
	if err != nil {
		return "", fmt.Errorf("openai API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
