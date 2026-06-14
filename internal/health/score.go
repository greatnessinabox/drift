// Package health computes a weighted codebase health score from analyzer
// results. Each metric (complexity, deps, boundaries, dead code, coverage) is
// scored 0-100, then combined as a weighted average using cfg.Weights. Weights
// are renormalized over the metrics actually present, so a missing metric (e.g.
// coverage with no lcov.info report) neither inflates nor tanks the total.
package health

import (
	"math"

	"github.com/greatnessinabox/drift/internal/analyzer"
	"github.com/greatnessinabox/drift/internal/config"
)

type Score struct {
	Total            float64
	Complexity       float64
	Deps             float64
	Boundaries       float64
	DeadCode         float64
	Coverage         float64
	CoverageMeasured bool
	Delta            float64
}

type Scorer struct {
	cfg      *config.Config
	previous float64
}

func NewScorer(cfg *config.Config) *Scorer {
	return &Scorer{cfg: cfg, previous: -1}
}

func (s *Scorer) Calculate(r *analyzer.Results) Score {
	score := Score{
		Complexity: s.complexityScore(r),
		Deps:       s.depsScore(r),
		Boundaries: s.boundariesScore(r),
		DeadCode:   s.deadCodeScore(r),
	}

	w := s.cfg.Weights
	weightedSum := score.Complexity*w.Complexity +
		score.Deps*w.Deps +
		score.Boundaries*w.Boundaries +
		score.DeadCode*w.DeadCode
	totalWeight := w.Complexity + w.Deps + w.Boundaries + w.DeadCode

	// Coverage only counts when an lcov report was found. Otherwise the
	// remaining weights are renormalized so the score isn't inflated by an
	// assumed-perfect metric.
	if r.Coverage.Measured {
		score.Coverage = r.Coverage.Percent
		score.CoverageMeasured = true
		weightedSum += score.Coverage * w.Coverage
		totalWeight += w.Coverage
	}

	if totalWeight > 0 {
		score.Total = weightedSum / totalWeight
	}
	score.Total = math.Round(score.Total*10) / 10

	if s.previous >= 0 {
		score.Delta = score.Total - s.previous
	}
	s.previous = score.Total

	return score
}

func (s *Scorer) complexityScore(r *analyzer.Results) float64 {
	if len(r.Complexity) == 0 {
		return 100
	}

	threshold := float64(s.cfg.Thresholds.MaxComplexity)
	if threshold == 0 {
		threshold = 15
	}

	var totalPenalty float64
	for _, fc := range r.Complexity {
		if float64(fc.Complexity) > threshold {
			excess := float64(fc.Complexity) - threshold
			penalty := math.Min(excess/threshold*20, 20)
			totalPenalty += penalty
		}
	}

	score := 100 - totalPenalty
	return math.Max(0, math.Min(100, score))
}

func (s *Scorer) depsScore(r *analyzer.Results) float64 {
	if len(r.Dependencies) == 0 {
		return 100
	}

	maxStale := float64(s.cfg.Thresholds.MaxStaleDays)
	if maxStale == 0 {
		maxStale = 90
	}

	var totalPenalty float64
	for _, dep := range r.Dependencies {
		if dep.StaleDays > 0 {
			ratio := float64(dep.StaleDays) / maxStale
			penalty := math.Min(ratio*15, 15)
			totalPenalty += penalty
		}
	}

	score := 100 - totalPenalty
	return math.Max(0, math.Min(100, score))
}

func (s *Scorer) boundariesScore(r *analyzer.Results) float64 {
	if len(r.Violations) == 0 {
		return 100
	}

	penalty := float64(len(r.Violations)) * 10
	score := 100 - penalty
	return math.Max(0, math.Min(100, score))
}

func (s *Scorer) deadCodeScore(r *analyzer.Results) float64 {
	if len(r.DeadCode) == 0 {
		return 100
	}

	penalty := float64(len(r.DeadCode)) * 5
	score := 100 - penalty
	return math.Max(0, math.Min(100, score))
}
