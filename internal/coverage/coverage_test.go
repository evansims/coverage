package coverage

import "testing"

func TestRun(t *testing.T) {
	if err := Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
}
