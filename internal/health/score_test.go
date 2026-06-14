package health

import (
	"testing"

	"github.com/greatnessinabox/drift/internal/analyzer"
	"github.com/greatnessinabox/drift/internal/config"
)

func newScorer() *Scorer {
	return NewScorer(config.Defaults())
}

func TestCalculate_PerfectScore(t *testing.T) {
	got := newScorer().Calculate(&analyzer.Results{})

	for name, v := range map[string]float64{
		"Total": got.Total, "Complexity": got.Complexity, "Deps": got.Deps,
		"Boundaries": got.Boundaries, "DeadCode": got.DeadCode, "Coverage": got.Coverage,
	} {
		if v != 100 {
			t.Errorf("empty results: %s = %v, want 100", name, v)
		}
	}
}

func TestComplexityScore(t *testing.T) {
	// default threshold = 15; penalty per func = min((c-threshold)/threshold*20, 20)
	tests := []struct {
		name       string
		complexity []int
		want       float64
	}{
		{"none", nil, 100},
		{"under threshold", []int{10, 15}, 100},
		{"one over", []int{20}, 93.33333333333333},  // excess 5 -> 6.667
		{"penalty capped per func", []int{100}, 80}, // capped at 20
		{"two maxed", []int{100, 100}, 60},          // 40 penalty
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &analyzer.Results{Complexity: make([]analyzer.FunctionComplexity, 0, len(tt.complexity))}
			for _, c := range tt.complexity {
				r.Complexity = append(r.Complexity, analyzer.FunctionComplexity{Complexity: c})
			}
			if got := newScorer().complexityScore(r); !approx(got, tt.want) {
				t.Errorf("complexityScore = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDepsScore(t *testing.T) {
	// default maxStale = 90; penalty per dep = min(stale/90*15, 15)
	tests := []struct {
		name string
		days []int
		want float64
	}{
		{"none", nil, 100},
		{"fresh", []int{0}, 100},
		{"half stale", []int{45}, 92.5},
		{"fully stale", []int{90}, 85},
		{"over capped", []int{900}, 85},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &analyzer.Results{}
			for _, d := range tt.days {
				r.Dependencies = append(r.Dependencies, analyzer.DepStatus{StaleDays: d})
			}
			if got := newScorer().depsScore(r); !approx(got, tt.want) {
				t.Errorf("depsScore = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoundariesScore(t *testing.T) {
	tests := []struct {
		n    int
		want float64
	}{{0, 100}, {1, 90}, {5, 50}, {11, 0}} // 10 per violation, clamped at 0
	for _, tt := range tests {
		r := &analyzer.Results{Violations: make([]analyzer.BoundaryViolation, tt.n)}
		if got := newScorer().boundariesScore(r); got != tt.want {
			t.Errorf("boundariesScore(%d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}

func TestDeadCodeScore(t *testing.T) {
	tests := []struct {
		n    int
		want float64
	}{{0, 100}, {1, 95}, {10, 50}, {21, 0}} // 5 per dead func, clamped at 0
	for _, tt := range tests {
		r := &analyzer.Results{DeadCode: make([]analyzer.DeadFunction, tt.n)}
		if got := newScorer().deadCodeScore(r); got != tt.want {
			t.Errorf("deadCodeScore(%d) = %v, want %v", tt.n, got, tt.want)
		}
	}
}

func TestCalculate_WeightsApplied(t *testing.T) {
	cfg := config.Defaults()
	cfg.Weights = config.WeightConfig{Complexity: 1, Deps: 0, Boundaries: 0, DeadCode: 0, Coverage: 0}
	// only complexity counts; one maxed-out func -> complexityScore 80 -> total 80
	r := &analyzer.Results{Complexity: []analyzer.FunctionComplexity{{Complexity: 100}}}

	if got := NewScorer(cfg).Calculate(r).Total; got != 80 {
		t.Errorf("weighted total = %v, want 80", got)
	}
}

func TestCalculate_Delta(t *testing.T) {
	s := newScorer()

	first := s.Calculate(&analyzer.Results{})
	if first.Delta != 0 {
		t.Errorf("first delta = %v, want 0 (no previous)", first.Delta)
	}

	second := s.Calculate(&analyzer.Results{
		DeadCode: make([]analyzer.DeadFunction, 4), // drops the score
	})
	if second.Delta >= 0 {
		t.Errorf("delta after regression = %v, want negative", second.Delta)
	}
	if !approx(second.Delta, second.Total-first.Total) {
		t.Errorf("delta %v != total change %v", second.Delta, second.Total-first.Total)
	}
}

func approx(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 0.0001
}
