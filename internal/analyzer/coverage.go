package analyzer

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Coverage holds overall line-coverage parsed from a standard report.
// Measured is false when no report was found, so scoring can exclude it
// instead of assuming a value.
type Coverage struct {
	Percent  float64
	Measured bool
}

// readCoverage computes overall line coverage from an lcov report under root.
// ponytail: lcov.info only (cross-language); add go coverage.out / cobertura
// if a user actually needs them.
func readCoverage(root string) Coverage {
	path := findCoverageFile(root)
	if path == "" {
		return Coverage{}
	}

	f, err := os.Open(path)
	if err != nil {
		return Coverage{}
	}
	defer f.Close()

	var found, hit int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "LF:"):
			n, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "LF:")))
			found += n
		case strings.HasPrefix(line, "LH:"):
			n, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "LH:")))
			hit += n
		}
	}
	if found == 0 {
		return Coverage{}
	}

	return Coverage{Percent: float64(hit) / float64(found) * 100, Measured: true}
}

func findCoverageFile(root string) string {
	for _, name := range []string{"lcov.info", "coverage/lcov.info"} {
		p := filepath.Join(root, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
