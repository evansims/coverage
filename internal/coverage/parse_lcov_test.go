package coverage

import (
	"os"
	"path/filepath"
	"strings"
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

func TestParseLcovSummaryFallback(t *testing.T) {
	// LCOV with only summary lines (LF/LH) and no DA lines
	data := []byte(`SF:src/app.js
LF:100
LH:80
BRF:20
BRH:15
FNF:10
FNH:8
end_of_record
SF:src/lib.js
LF:50
LH:40
BRF:10
BRH:5
FNF:5
FNH:3
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use summary-line fallback
	assertMetric(t, "line", result.Line, &Metric{Hit: 120, Total: 150})
	assertMetric(t, "branch", result.Branch, &Metric{Hit: 20, Total: 30})
	assertMetric(t, "function", result.Function, &Metric{Hit: 11, Total: 15})
}

func TestParseLcovBranchNotTaken(t *testing.T) {
	// BRDA with "-" for not-taken branch
	data := []byte(`SF:src/main.js
DA:1,1
BRDA:1,0,0,1
BRDA:1,0,1,-
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertMetric(t, "line", result.Line, &Metric{Hit: 1, Total: 1})
	assertMetric(t, "branch", result.Branch, &Metric{Hit: 1, Total: 2})
}

func TestParseLcovFunctionDeclarations(t *testing.T) {
	// FN declares a function, FNDA provides the count
	data := []byte(`SF:src/utils.js
FN:1,foo
FN:10,bar
FNDA:5,foo
DA:1,5
DA:2,5
DA:10,0
DA:11,0
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertMetric(t, "line", result.Line, &Metric{Hit: 2, Total: 4})
	// foo covered (count=5), bar uncovered (declared via FN but no FNDA or FNDA=0)
	assertMetric(t, "function", result.Function, &Metric{Hit: 1, Total: 2})
}

func TestParseLcovNoEndOfRecord(t *testing.T) {
	// File without trailing end_of_record — should still parse
	data := []byte(`SF:src/main.go
DA:1,1
DA:2,1
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertMetric(t, "line", result.Line, &Metric{Hit: 2, Total: 2})
}

func TestParseLcovMalformedLines(t *testing.T) {
	// Malformed DA, BRDA, FNDA lines should be skipped gracefully
	data := []byte(`SF:src/main.go
DA:1,1
DA:bad
DA:2,notanumber
BRDA:incomplete
BRDA:1,0,0,notanumber
FNDA:bad
FNDA:notanumber,func
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the valid DA:1,1 should be parsed
	assertMetric(t, "line", result.Line, &Metric{Hit: 1, Total: 1})
}

func TestParseLcovEmptyRecords(t *testing.T) {
	data := []byte(`
end_of_record
end_of_record
`)

	_, err := parseLcov(data)
	if err == nil {
		t.Fatal("expected error for empty records")
	}
}

func TestParseLcovFileDetails(t *testing.T) {
	data := []byte(`SF:src/main.go
DA:1,1
DA:2,0
DA:3,5
BRDA:3,0,0,2
BRDA:3,0,1,0
FNDA:3,main
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FileDetails == nil {
		t.Fatal("expected FileDetails to be populated")
	}
	detail := result.FileDetails["src/main.go"]
	if detail == nil {
		t.Fatal("expected detail for src/main.go")
	}
	if detail.Lines[1] != 1 {
		t.Errorf("line 1 count = %d, want 1", detail.Lines[1])
	}
	if detail.Lines[2] != 0 {
		t.Errorf("line 2 count = %d, want 0", detail.Lines[2])
	}
	if len(detail.Branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(detail.Branches))
	}
	if detail.Functions["main"] != 3 {
		t.Errorf("function main count = %d, want 3", detail.Functions["main"])
	}
}

func TestParseLcovSummaryFallbackNoBranches(t *testing.T) {
	// LCOV with summary lines only, no branches or functions
	data := []byte(`SF:src/app.js
LF:100
LH:80
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertMetric(t, "line", result.Line, &Metric{Hit: 80, Total: 100})
	if result.Branch != nil {
		t.Error("expected nil branch when no BRF data")
	}
	if result.Function != nil {
		t.Error("expected nil function when no FNF data")
	}
}

func TestParseLcovSummaryFallbackDuplicateFiles(t *testing.T) {
	// Summary-only LCOV with duplicate source files should accumulate
	data := []byte(`SF:src/app.js
LF:100
LH:80
BRF:10
BRH:5
FNF:5
FNH:3
end_of_record
SF:src/app.js
LF:50
LH:40
BRF:10
BRH:8
FNF:5
FNH:5
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Summaries should accumulate for duplicate files
	assertMetric(t, "line", result.Line, &Metric{Hit: 120, Total: 150})
	assertMetric(t, "branch", result.Branch, &Metric{Hit: 13, Total: 20})
	assertMetric(t, "function", result.Function, &Metric{Hit: 8, Total: 10})
}

func TestParseLcovBlankLines(t *testing.T) {
	// LCOV with blank and whitespace lines should be handled
	data := []byte(`
SF:src/main.go

DA:1,1
DA:2,0

end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertMetric(t, "line", result.Line, &Metric{Hit: 1, Total: 2})
}

func TestParseLcovDuplicateSourceFileBranchMerge(t *testing.T) {
	// Test that duplicate source files merge branches and functions using max semantics
	data := []byte(`SF:src/main.go
DA:1,1
BRDA:1,0,0,2
BRDA:1,0,1,0
FNDA:3,main
FNDA:0,helper
end_of_record
SF:src/main.go
DA:1,5
BRDA:1,0,0,1
BRDA:1,0,1,3
FNDA:1,main
FNDA:2,helper
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	detail := result.FileDetails["src/main.go"]
	if detail == nil {
		t.Fatal("expected detail for src/main.go")
	}

	// Lines: max(1, 5) = 5
	if detail.Lines[1] != 5 {
		t.Errorf("line 1 count = %d, want 5 (max merge)", detail.Lines[1])
	}

	// Branches: "1:0:0" = max(2,1)=2, "1:0:1" = max(0,3)=3
	if detail.Branches["1:0:0"] != 2 {
		t.Errorf("branch 1:0:0 = %d, want 2", detail.Branches["1:0:0"])
	}
	if detail.Branches["1:0:1"] != 3 {
		t.Errorf("branch 1:0:1 = %d, want 3", detail.Branches["1:0:1"])
	}

	// Functions: main = max(3,1)=3, helper = max(0,2)=2
	if detail.Functions["main"] != 3 {
		t.Errorf("function main = %d, want 3", detail.Functions["main"])
	}
	if detail.Functions["helper"] != 2 {
		t.Errorf("function helper = %d, want 2", detail.Functions["helper"])
	}
}

func TestParseLcovDABadLineNumber(t *testing.T) {
	// DA where line number is not an integer — should skip
	data := []byte(`SF:src/main.go
DA:1,1
DA:notanumber,5
DA:3,1
end_of_record
`)
	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only lines 1 and 3 should be counted
	assertMetric(t, "line", result.Line, &Metric{Hit: 2, Total: 2})
}

func TestParseLcovScannerError(t *testing.T) {
	// Create data with a line exceeding bufio.Scanner's max token size (64KB)
	longLine := make([]byte, 70000)
	for i := range longLine {
		longLine[i] = 'x'
	}
	data := append([]byte("SF:src/main.go\nDA:1,1\n"), longLine...)
	data = append(data, '\n')

	_, err := parseLcov(data)
	if err == nil {
		t.Fatal("expected error for extremely long line")
	}
	if !strings.Contains(err.Error(), "reading lcov data") {
		t.Errorf("error should mention reading: %v", err)
	}
}

func TestParseLcovMalformedFNDA(t *testing.T) {
	// FNDA with insufficient parts
	data := []byte(`SF:src/main.go
DA:1,1
FNDA:nopair
FN:badline
end_of_record
`)

	result, err := parseLcov(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only the DA line should contribute
	assertMetric(t, "line", result.Line, &Metric{Hit: 1, Total: 1})
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
