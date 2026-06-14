package ai

import (
	"testing"

	"github.com/greatnessinabox/drift/internal/config"
)

func TestMaxTokensOrDefault(t *testing.T) {
	tests := []struct {
		in   int
		want int64
	}{{0, 1024}, {-5, 1024}, {2048, 2048}}
	for _, tt := range tests {
		if got := maxTokensOrDefault(tt.in); got != tt.want {
			t.Errorf("maxTokensOrDefault(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestNewAnthropicProvider_TokenBudget(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	p, err := NewAnthropicProvider(config.AIConfig{MaxTokens: 4096})
	if err != nil {
		t.Fatal(err)
	}
	if p.maxTokens != 4096 {
		t.Errorf("configured budget = %d, want 4096", p.maxTokens)
	}

	def, err := NewAnthropicProvider(config.AIConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if def.maxTokens != 1024 {
		t.Errorf("default budget = %d, want 1024", def.maxTokens)
	}
}

func TestNewOpenAIProvider_TokenBudget(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	p, err := NewOpenAIProvider(config.AIConfig{MaxTokens: 4096})
	if err != nil {
		t.Fatal(err)
	}
	if p.maxTokens != 4096 {
		t.Errorf("configured budget = %d, want 4096", p.maxTokens)
	}
}
