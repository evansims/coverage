package coverage

import "testing"

func TestGetParser(t *testing.T) {
	formats := []string{"lcov", "gocover", "cobertura", "clover", "jacoco"}
	for _, f := range formats {
		t.Run(f, func(t *testing.T) {
			p, err := getParser(f)
			if err != nil {
				t.Fatalf("getParser(%q) returned error: %v", f, err)
			}
			if p == nil {
				t.Fatalf("getParser(%q) returned nil", f)
			}
		})
	}

	t.Run("unknown format", func(t *testing.T) {
		_, err := getParser("unknown")
		if err == nil {
			t.Fatal("expected error for unknown format")
		}
	})
}
