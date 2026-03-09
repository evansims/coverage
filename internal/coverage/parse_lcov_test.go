package coverage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLcov(t *testing.T) {
	tests := []struct {
		name         string
		fixture      string
		wantLine     *Metric
		wantBranch   *Metric
		wantFunction *Metric
		wantErr      bool
	}{
		{
			name:         "basic coverage",
			fixture:      "lcov/basic.info",
			wantLine:     &Metric{Hit: 3, Total: 4},
			wantBranch:   &Metric{Hit: 1, Total: 2},
			wantFunction: &Metric{Hit: 1, Total: 1},
		},
		{
			name:         "multi file sums counters",
			fixture:      "lcov/multi_file.info",
			wantLine:     &Metric{Hit: 23, Total: 30},
			wantBranch:   &Metric{Hit: 7, Total: 10},
			wantFunction: &Metric{Hit: 3, Total: 5},
		},
		{
			name:         "no branches reported",
			fixture:      "lcov/no_branches.info",
			wantLine:     &Metric{Hit: 40, Total: 50},
			wantBranch:   nil,
			wantFunction: &Metric{Hit: 4, Total: 5},
		},
		{
			name:         "zero branch totals treated as no branch data",
			fixture:      "lcov/zero_branches.info",
			wantLine:     &Metric{Hit: 4, Total: 4},
			wantBranch:   nil,
			wantFunction: &Metric{Hit: 2, Total: 2},
		},
		{
			name:    "empty file",
			fixture: "lcov/empty.info",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", "testdata", tt.fixture))
			if err != nil {
				t.Fatal(err)
			}
			result, err := parseLcov(data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseLcov() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			assertMetric(t, "line", result.Line, tt.wantLine)
			assertMetric(t, "branch", result.Branch, tt.wantBranch)
			assertMetric(t, "function", result.Function, tt.wantFunction)
		})
	}
}

func TestParseLcovDuplicateSourceFiles(t *testing.T) {
	// Simulates lcov --add-tracefile output where the same file appears twice
	data := []byte(`SF:src/main.go
DA:1,1
DA:2,0
DA:3,1
LF:3
LH:2
end_of_record
SF:src/main.go
DA:2,1
DA:4,1
LF:2
LH:2
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Line 2 should be merged: max(0, 1) = 1
	// Total unique lines: 1, 2, 3, 4 = 4 lines
	// Hit lines: 1(1), 2(1), 3(1), 4(1) = 4 hit
	if result.Line == nil {
		t.Fatal("expected line metric")
	}
	if result.Line.Total != 4 {
		t.Errorf("expected 4 total lines, got %d", result.Line.Total)
	}
	if result.Line.Hit != 4 {
		t.Errorf("expected 4 hit lines, got %d", result.Line.Hit)
	}
}

func assertMetric(t *testing.T, name string, got, want *Metric) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("%s: expected nil, got %+v", name, got)
		}
		return
	}
	if got == nil {
		t.Fatalf("%s: expected %+v, got nil", name, want)
	}
	if got.Hit != want.Hit || got.Total != want.Total {
		t.Errorf("%s: got {Hit:%d, Total:%d}, want {Hit:%d, Total:%d}",
			name, got.Hit, got.Total, want.Hit, want.Total)
	}
}
