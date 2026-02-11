package history

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/greatnessinabox/drift/internal/analyzer"
	"github.com/greatnessinabox/drift/internal/config"
	"github.com/greatnessinabox/drift/internal/health"
)

type SparklineData struct {
	HealthScore    []float64
	AvgComplexity  []float64
	ViolationCount []float64
	DeadCodeCount  []float64
}

type Analyzer struct {
	cfg    *config.Config
	repo   *git.Repository
	scorer *health.Scorer
}

func New(cfg *config.Config) (*Analyzer, error) {
	repoPath := cfg.Root
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("opening git repo: %w", err)
	}

	return &Analyzer{
		cfg:    cfg,
		repo:   repo,
		scorer: health.NewScorer(cfg),
	}, nil
}

func (a *Analyzer) Walk(maxCommits int) (*SparklineData, error) {
	if maxCommits <= 0 {
		maxCommits = 10
	}

	ref, err := a.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("getting HEAD: %w", err)
	}

	commits, err := a.getCommits(ref.Hash(), maxCommits)
	if err != nil {
		return nil, fmt.Errorf("getting commits: %w", err)
	}

	if len(commits) == 0 {
		return &SparklineData{}, nil
	}

	data := &SparklineData{
		HealthScore:    make([]float64, 0, len(commits)),
		AvgComplexity:  make([]float64, 0, len(commits)),
		ViolationCount: make([]float64, 0, len(commits)),
		DeadCodeCount:  make([]float64, 0, len(commits)),
	}

	// Analyze commits from oldest to newest
	for i := len(commits) - 1; i >= 0; i-- {
		commit := commits[i]
		
		results, err := a.analyzeCommit(commit)
		if err != nil {
			// Skip commits that fail to analyze
			continue
		}

		score := a.scorer.Calculate(results)
		data.HealthScore = append(data.HealthScore, score.Total)
		data.AvgComplexity = append(data.AvgComplexity, avgComplexity(results.Complexity))
		data.ViolationCount = append(data.ViolationCount, float64(len(results.Violations)))
		data.DeadCodeCount = append(data.DeadCodeCount, float64(len(results.DeadCode)))
	}

	return data, nil
}

func (a *Analyzer) getCommits(hash plumbing.Hash, max int) ([]*object.Commit, error) {
	cIter, err := a.repo.Log(&git.LogOptions{From: hash})
	if err != nil {
		return nil, err
	}
	defer cIter.Close()

	var commits []*object.Commit
	count := 0

	err = cIter.ForEach(func(c *object.Commit) error {
		if count >= max {
			return fmt.Errorf("max reached")
		}
		commits = append(commits, c)
		count++
		return nil
	})

	if err != nil && err.Error() != "max reached" {
		return nil, err
	}

	return commits, nil
}

func (a *Analyzer) analyzeCommit(commit *object.Commit) (*analyzer.Results, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("getting tree: %w", err)
	}

	// Create a temporary directory for this commit's files
	tmpDir, err := os.MkdirTemp("", "drift-history-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract source files from this commit using detected language
	langAna := analyzer.NewLanguageAnalyzer(analyzer.DetectLanguage(a.cfg.Root))
	extSet := make(map[string]bool)
	for _, ext := range langAna.Extensions() {
		extSet[ext] = true
	}

	err = tree.Files().ForEach(func(f *object.File) error {
		if !extSet[filepath.Ext(f.Name)] {
			return nil
		}

		// Skip excluded directories
		for _, exclude := range a.cfg.Exclude {
			if matched, _ := filepath.Match(exclude+"/*", f.Name); matched {
				return nil
			}
		}

		content, err := f.Contents()
		if err != nil {
			return err
		}

		dest := filepath.Join(tmpDir, f.Name)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}

		return os.WriteFile(dest, []byte(content), 0644)
	})

	if err != nil {
		return nil, fmt.Errorf("extracting files: %w", err)
	}

	// Create a temporary config pointing to the temp directory
	tmpCfg := *a.cfg
	tmpCfg.Root = tmpDir

	// Analyze the extracted files
	ana := analyzer.New(&tmpCfg)
	return ana.Run()
}

func avgComplexity(complexities []analyzer.FunctionComplexity) float64 {
	if len(complexities) == 0 {
		return 0
	}

	sum := 0
	for _, c := range complexities {
		sum += c.Complexity
	}
	return float64(sum) / float64(len(complexities))
}
