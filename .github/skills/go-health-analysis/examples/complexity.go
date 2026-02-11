package examples

// This file demonstrates cyclomatic complexity patterns for the go-health-analysis skill.

// SimpleFunction has complexity 1 (base only)
func SimpleFunction() int {
	return 42
}

// MediumFunction has complexity 5 (1 base + 2 if + 1 for + 1 range)
func MediumFunction(items []string) []string {
	var results []string
	if len(items) == 0 {
		return nil
	}
	for i := 0; i < len(items); i++ {
		if items[i] != "" {
			results = append(results, items[i])
		}
	}
	for range items {
		// additional processing
	}
	return results
}

// ComplexFunction has high complexity and should be refactored
// Complexity: 1 (base) + 3 (if) + 1 (for) + 4 (case) + 2 (&&/||) = 11
func ComplexFunction(input string, mode int) (string, error) {
	if input == "" {
		return "", nil
	}

	if mode < 0 || mode > 10 {
		return "", nil
	}

	var result string
	for i := 0; i < len(input); i++ {
		switch mode {
		case 0:
			result += "a"
		case 1:
			result += "b"
		case 2:
			if input[i] == 'x' && len(result) > 0 {
				result += "c"
			}
		case 3:
			result += "d"
		}
	}

	return result, nil
}
