package main

import (
	"errors"
	"strings"
	"testing"
)

func TestRenderFixPlan(t *testing.T) {
	plan := []fixSuggestion{
		{issue: fixIssue{Description: "foo() too complex"}, suggestion: "extract a helper"},
		{issue: fixIssue{Description: "bar() too complex"}, err: errors.New("boom")},
	}

	md := renderFixPlan(plan)

	for _, want := range []string{
		"# drift fix plan (2 issue(s))",
		"## 1. foo() too complex",
		"extract a helper",
		"## 2. bar() too complex",
		"Could not generate a suggestion: boom",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("plan missing %q\n---\n%s", want, md)
		}
	}
}
