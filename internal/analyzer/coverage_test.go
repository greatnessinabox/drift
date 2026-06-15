package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCoverage(t *testing.T) {
	dir := t.TempDir()
	lcov := "SF:foo.go\nLF:10\nLH:8\nend_of_record\nSF:bar.go\nLF:10\nLH:2\nend_of_record\n"
	if err := os.WriteFile(filepath.Join(dir, "lcov.info"), []byte(lcov), 0o644); err != nil {
		t.Fatal(err)
	}

	cov := readCoverage(dir)
	if !cov.Measured {
		t.Fatal("Measured = false, want true")
	}
	if cov.Percent != 50 { // 10 hit / 20 found
		t.Errorf("Percent = %v, want 50", cov.Percent)
	}
}

func TestReadCoverage_NoReport(t *testing.T) {
	if cov := readCoverage(t.TempDir()); cov.Measured {
		t.Error("Measured = true with no report, want false")
	}
}

func TestReadCoverage_MalformedLine(t *testing.T) {
	dir := t.TempDir()
	// A non-numeric LF line must be skipped, not counted as zero.
	lcov := "SF:foo.go\nLF:10\nLH:5\nLF:bogus\nend_of_record\n"
	if err := os.WriteFile(filepath.Join(dir, "lcov.info"), []byte(lcov), 0o644); err != nil {
		t.Fatal(err)
	}
	if cov := readCoverage(dir); !cov.Measured || cov.Percent != 50 { // 5/10, bogus ignored
		t.Errorf("got %+v, want {Percent:50 Measured:true}", cov)
	}
}
