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

// readCoverage computes overall coverage from a report under root, preferring
// lcov.info (cross-language) and falling back to Go's coverage.out.
// ponytail: lcov + go coverage.out; add cobertura/clover if a user needs them.
func readCoverage(root string) Coverage {
	if p := findFile(root, "lcov.info", "coverage/lcov.info"); p != "" {
		return parseLcov(p)
	}
	if p := findFile(root, "coverage.out"); p != "" {
		return parseGoCover(p)
	}
	return Coverage{}
}

func findFile(root string, names ...string) string {
	for _, name := range names {
		p := filepath.Join(root, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// parseLcov sums LF (lines found) and LH (lines hit) records.
func parseLcov(path string) Coverage {
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
			if n, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "LF:"))); err == nil {
				found += n
			}
		case strings.HasPrefix(line, "LH:"):
			if n, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "LH:"))); err == nil {
				hit += n
			}
		}
	}
	// A failed read is reported as unmeasured rather than as partial coverage.
	if err := sc.Err(); err != nil || found == 0 {
		return Coverage{}
	}
	return Coverage{Percent: float64(hit) / float64(found) * 100, Measured: true}
}

// parseGoCover sums statements from a `go test -coverprofile` report. Each
// non-header line is "file:startLine.col,endLine.col numStmts count".
func parseGoCover(path string) Coverage {
	f, err := os.Open(path)
	if err != nil {
		return Coverage{}
	}
	defer f.Close()

	var total, covered int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 3 {
			continue // "mode: set" header or malformed line
		}
		stmts, err1 := strconv.Atoi(fields[1])
		count, err2 := strconv.Atoi(fields[2])
		if err1 != nil || err2 != nil {
			continue
		}
		total += stmts
		if count > 0 {
			covered += stmts
		}
	}
	if err := sc.Err(); err != nil || total == 0 {
		return Coverage{}
	}
	return Coverage{Percent: float64(covered) / float64(total) * 100, Measured: true}
}
