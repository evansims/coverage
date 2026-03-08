package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverReport(t *testing.T) {
	t.Run("finds lcov.info in coverage dir", func(t *testing.T) {
		dir := t.TempDir()
		coverageDir := filepath.Join(dir, "coverage")
		if err := os.Mkdir(coverageDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(coverageDir, "lcov.info"), []byte("SF:foo\nend_of_record\n"), 0644); err != nil {
			t.Fatal(err)
		}

		path, err := DiscoverReport("lcov", dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if path != "coverage/lcov.info" {
			t.Errorf("expected 'coverage/lcov.info', got %q", path)
		}
	})

	t.Run("finds cover.out for gocover", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "cover.out"), []byte("mode: set\n"), 0644); err != nil {
			t.Fatal(err)
		}

		path, err := DiscoverReport("gocover", dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if path != "cover.out" {
			t.Errorf("expected 'cover.out', got %q", path)
		}
	})

	t.Run("returns error when no file found", func(t *testing.T) {
		dir := t.TempDir()

		_, err := DiscoverReport("lcov", dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "auto-discovery") {
			t.Errorf("error should mention auto-discovery: %v", err)
		}
	})

	t.Run("returns error for unknown format", func(t *testing.T) {
		_, err := DiscoverReport("unknown", ".")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no default paths") {
			t.Errorf("error should mention no default paths: %v", err)
		}
	})

	t.Run("prefers first matching path", func(t *testing.T) {
		dir := t.TempDir()
		// Create both coverage/lcov.info and lcov.info — should find coverage/lcov.info first
		coverageDir := filepath.Join(dir, "coverage")
		if err := os.Mkdir(coverageDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(coverageDir, "lcov.info"), []byte("SF:foo\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "lcov.info"), []byte("SF:bar\n"), 0644); err != nil {
			t.Fatal(err)
		}

		path, err := DiscoverReport("lcov", dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if path != "coverage/lcov.info" {
			t.Errorf("expected 'coverage/lcov.info', got %q", path)
		}
	})
}
