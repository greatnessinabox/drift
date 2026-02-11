package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Root     string   `yaml:"root"`
	Language string   `yaml:"language"` // empty = auto-detect; "go", "typescript", "python", "rust", "java"
	Exclude  []string `yaml:"exclude"`

	Weights WeightConfig `yaml:"weights"`

	Boundaries []BoundaryRule `yaml:"boundaries"`

	AI AIConfig `yaml:"ai"`

	Thresholds ThresholdConfig `yaml:"thresholds"`
}

type WeightConfig struct {
	Complexity float64 `yaml:"complexity"`
	Deps       float64 `yaml:"deps"`
	Boundaries float64 `yaml:"boundaries"`
	DeadCode   float64 `yaml:"dead_code"`
	Coverage   float64 `yaml:"coverage"`
}

type BoundaryRule struct {
	Deny string `yaml:"deny"` // e.g. "pkg/api -> internal/db"
}

type AIConfig struct {
	Provider string `yaml:"provider"` // "anthropic" or "openai"
	Model    string `yaml:"model"`    // e.g. "claude-sonnet-4-5-20250929" or "gpt-4o"
}

type ThresholdConfig struct {
	MaxComplexity int     `yaml:"max_complexity"` // per-function complexity threshold
	MaxStaleDays  int     `yaml:"max_stale_days"` // dependency staleness threshold
	MinScore      float64 `yaml:"min_score"`      // minimum acceptable health score
}

func Defaults() *Config {
	cwd, _ := os.Getwd()
	return &Config{
		Root: cwd,
		Exclude: []string{
			"vendor",
			"node_modules",
			".git",
			"testdata",
			"__pycache__",
			".venv",
			"target",
			"dist",
			"build",
		},
		Weights: WeightConfig{
			Complexity: 0.30,
			Deps:       0.20,
			Boundaries: 0.20,
			DeadCode:   0.15,
			Coverage:   0.15,
		},
		AI: AIConfig{
			Provider: "anthropic",
			Model:    "claude-sonnet-4-5-20250929",
		},
		Thresholds: ThresholdConfig{
			MaxComplexity: 15,
			MaxStaleDays:  90,
			MinScore:      70,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Defaults()

	if path == "" {
		path = findConfigFile()
	}

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if cfg.Root == "" {
		cwd, _ := os.Getwd()
		cfg.Root = cwd
	}

	if !filepath.IsAbs(cfg.Root) {
		abs, err := filepath.Abs(cfg.Root)
		if err != nil {
			return nil, fmt.Errorf("resolving root path: %w", err)
		}
		cfg.Root = abs
	}

	return cfg, nil
}

func findConfigFile() string {
	candidates := []string{
		".drift.yaml",
		".drift.yml",
		"drift.yaml",
		"drift.yml",
	}

	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}

	return ""
}

func RunInitWizard() error {
	cfg := Defaults()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	header := "# drift configuration\n# https://github.com/greatnessinabox/drift\n\n"

	if err := os.WriteFile(".drift.yaml", []byte(header+string(data)), 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Println("Created .drift.yaml with default settings")
	return nil
}
