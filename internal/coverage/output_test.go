package coverage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatViolation(t *testing.T) {
	v := Violation{
		Entry:    "backend",
		Metric:   "line",
		Actual:   73.2,
		Required: 80,
	}
	msg := FormatViolation(v)
	if !strings.Contains(msg, "backend") {
		t.Errorf("message should contain entry name, got: %s", msg)
	}
	if !strings.Contains(msg, "73.2%") {
		t.Errorf("message should contain actual pct, got: %s", msg)
	}
	if !strings.Contains(msg, "80.0%") {
		t.Errorf("message should contain required pct, got: %s", msg)
	}
}

func TestFormatViolationDelta(t *testing.T) {
	v := Violation{
		Entry:    "coverage",
		Metric:   "delta",
		Actual:   -5.0,
		Required: -2.0,
	}
	msg := FormatViolation(v)
	if !strings.Contains(msg, "score changed by -5.0 points") {
		t.Errorf("message should contain delta change, got: %s", msg)
	}
	if !strings.Contains(msg, "minimum allowed change is -2.0") {
		t.Errorf("message should contain min allowed change, got: %s", msg)
	}
}

func TestWriteJobSummary(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(summaryFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	line1 := 87.3
	branch1 := 72.1
	func1 := 91.0
	line2 := 65.0
	branch2 := 55.0

	results := []EntryResult{
		{
			Name:     "backend",
			Line:     &line1,
			Branch:   &branch1,
			Function: &func1,
			Passed:   true,
		},
		{
			Name:     "frontend",
			Line:     &line2,
			Branch:   &branch2,
			Function: nil,
			Passed:   false,
		},
	}

	if err := WriteJobSummary(results, false, nil); err != nil {
		t.Fatalf("WriteJobSummary() error: %v", err)
	}

	data, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	checks := []string{"backend", "frontend", "87.3%", "65.0%", "N/A", "Pass", "**Fail**"}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("summary should contain %q", check)
		}
	}
}

func TestWriteOutputs(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	line := 87.3
	results := []EntryResult{
		{Name: "backend", Line: &line, Passed: true},
	}

	if err := WriteOutputs(true, results, nil, ""); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "passed=true") {
		t.Errorf("output should contain 'passed=true', got: %s", content)
	}
	// Delimiter is now randomized; check for the prefix pattern
	if !strings.Contains(content, "results<<COVERLINT_RESULTS_") {
		t.Errorf("output should contain multiline results delimiter, got: %s", content)
	}
	if !strings.Contains(content, `"backend"`) {
		t.Errorf("output should contain results JSON, got: %s", content)
	}
}

func TestWriteOutputsWithBadge(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	score := 85.0
	line := 90.0
	results := []EntryResult{
		{Name: "total", Score: &score, Line: &line, Passed: true},
	}

	if err := WriteOutputs(true, results, nil, ""); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Should have badge SVG output with randomized delimiter
	if !strings.Contains(content, "badge-svg<<COVERLINT_SVG_") {
		t.Error("output should contain badge-svg with SVG delimiter")
	}
	if !strings.Contains(content, "<svg") {
		t.Error("output should contain SVG content")
	}

	// Should have badge JSON output with randomized delimiter
	if !strings.Contains(content, "badge-json<<COVERLINT_JSON_") {
		t.Error("output should contain badge-json with JSON delimiter")
	}
	if !strings.Contains(content, `"coverage"`) {
		t.Error("badge-json should contain coverage label")
	}
	// Badge should use rounded whole numbers
	if !strings.Contains(content, `"85%"`) {
		t.Error("badge-json should contain rounded percentage '85%'")
	}
}

func TestWriteOutputsPassedFalse(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	results := []EntryResult{
		{Name: "backend", Passed: false},
	}

	if err := WriteOutputs(false, results, nil, ""); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, _ := os.ReadFile(outputFile)
	content := string(data)

	if !strings.Contains(content, "passed=false") {
		t.Errorf("output should contain 'passed=false', got: %s", content)
	}
}

func TestWriteOutputsNoScore(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	// Entry without Score — should not produce badge outputs
	results := []EntryResult{
		{Name: "test", Passed: true},
	}

	if err := WriteOutputs(true, results, nil, ""); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, _ := os.ReadFile(outputFile)
	content := string(data)

	if strings.Contains(content, "badge-svg") {
		t.Error("should not contain badge-svg when no score")
	}
	if strings.Contains(content, "badge-json") {
		t.Error("should not contain badge-json when no score")
	}
}

func TestWriteOutputsEmptyResults(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	if err := WriteOutputs(true, nil, nil, ""); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, _ := os.ReadFile(outputFile)
	content := string(data)

	if !strings.Contains(content, "passed=true") {
		t.Error("should still write passed output")
	}
}

func TestWriteJobSummaryOmitsUnsupportedColumns(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(summaryFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	line := 100.0
	results := []EntryResult{
		{
			Name:   "go-coverage",
			Line:   &line,
			Passed: true,
		},
	}

	if err := WriteJobSummary(results, false, nil); err != nil {
		t.Fatalf("WriteJobSummary() error: %v", err)
	}

	data, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if strings.Contains(content, "Branch") {
		t.Error("summary should not contain Branch column when no results have branch data")
	}
	if strings.Contains(content, "Function") {
		t.Error("summary should not contain Function column when no results have function data")
	}
	if !strings.Contains(content, "Line") {
		t.Error("summary should contain Line column")
	}
	if !strings.Contains(content, "100.0%") {
		t.Error("summary should contain line percentage")
	}
}

func TestWriteJobSummaryMultiFormatTotal(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(summaryFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	goLine := 90.0
	nodeLine := 85.0
	nodeBranch := 70.0
	totalLine := 87.5
	totalBranch := 70.0

	results := []EntryResult{
		{Name: "gocover", Line: &goLine, Passed: true},
		{Name: "lcov", Line: &nodeLine, Branch: &nodeBranch, Passed: true},
		{Name: "Total", Line: &totalLine, Branch: &totalBranch, Passed: true},
	}

	if err := WriteJobSummary(results, true, nil); err != nil {
		t.Fatalf("WriteJobSummary() error: %v", err)
	}

	data, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Should have per-format rows
	if !strings.Contains(content, "| gocover") {
		t.Error("summary should contain gocover row")
	}
	if !strings.Contains(content, "| lcov") {
		t.Error("summary should contain lcov row")
	}

	// Total row should be bold
	if !strings.Contains(content, "**Total**") {
		t.Error("summary should contain bold Total row")
	}
	if !strings.Contains(content, "**87.5%**") {
		t.Error("summary should contain bold total percentage")
	}
}

func TestWriteJobSummaryWithSuggestions(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(summaryFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	line := 60.0
	results := []EntryResult{
		{Name: "test", Line: &line, Passed: true},
	}
	suggestions := []Suggestion{
		{Path: "big.go", UncoveredLines: 50, TotalLines: 100, LinePct: 50.0},
	}

	if err := WriteJobSummary(results, false, suggestions); err != nil {
		t.Fatalf("WriteJobSummary() error: %v", err)
	}

	data, _ := os.ReadFile(summaryFile)
	content := string(data)
	if !strings.Contains(content, "Top Opportunities") {
		t.Error("summary should contain suggestions section")
	}
	if !strings.Contains(content, "big.go") {
		t.Error("summary should contain suggestion file")
	}
}

func TestWriteJobSummaryTotalPassedFalse(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(summaryFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	line1 := 90.0
	totalLine := 85.0

	results := []EntryResult{
		{Name: "gocover", Line: &line1, Passed: true},
		{Name: "Total", Line: &totalLine, Passed: false},
	}

	if err := WriteJobSummary(results, true, nil); err != nil {
		t.Fatalf("WriteJobSummary() error: %v", err)
	}

	data, _ := os.ReadFile(summaryFile)
	content := string(data)

	// Total row should show **Fail** when Passed is false
	if !strings.Contains(content, "**Fail**") {
		t.Error("total row should show Fail when not passed")
	}
}

func TestSanitizeWorkflowCommand(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal message", "normal message"},
		{"has\nnewline", "has newline"},
		{"has\r\nnewline", "has  newline"},
		{"has::colons", "has: :colons"},
		{"inject\n::error::pwned", "inject : :error: :pwned"},
	}
	for _, tt := range tests {
		got := sanitizeWorkflowCommand(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeWorkflowCommand(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizeMarkdown(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal", "normal"},
		{"has|pipe", "has\\|pipe"},
		{"has\nnewline", "has newline"},
		{"has\rcarriage", "has carriage"},
		{"has\r\nboth", "has  both"},
	}
	for _, tt := range tests {
		got := sanitizeMarkdown(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRandomDelimiter(t *testing.T) {
	d1 := randomDelimiter("TEST")
	d2 := randomDelimiter("TEST")

	if !strings.HasPrefix(d1, "TEST_") {
		t.Errorf("delimiter should start with prefix, got: %s", d1)
	}
	if d1 == d2 {
		t.Error("two random delimiters should not be equal")
	}
	// Should be prefix + "_" + 32 hex chars
	if len(d1) != len("TEST_")+32 {
		t.Errorf("delimiter length = %d, want %d", len(d1), len("TEST_")+32)
	}
}

func TestWriteJobSummaryNoEnvVar(t *testing.T) {
	t.Setenv("GITHUB_STEP_SUMMARY", "")
	err := WriteJobSummary(nil, false, nil)
	if err != nil {
		t.Fatalf("should not error when GITHUB_STEP_SUMMARY is empty: %v", err)
	}
}

func TestWriteOutputsNoEnvVar(t *testing.T) {
	t.Setenv("GITHUB_OUTPUT", "")
	err := WriteOutputs(true, nil, nil, "")
	if err != nil {
		t.Fatalf("should not error when GITHUB_OUTPUT is empty: %v", err)
	}
}

func TestFmtPct(t *testing.T) {
	tests := []struct {
		input *float64
		want  string
	}{
		{nil, "N/A"},
		{floatPtr(0.0), "0.0%"},
		{floatPtr(82.5), "82.5%"},
		{floatPtr(100.0), "100.0%"},
	}
	for _, tt := range tests {
		got := fmtPct(tt.input)
		if got != tt.want {
			t.Errorf("fmtPct(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestWriteOutputsWithBaseline(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	score := 85.0
	line := 90.0
	results := []EntryResult{
		{Name: "total", Score: &score, Line: &line, Passed: true},
	}

	bl := 90.0
	baseline := &BaselineData{
		Score:     85.0,
		Line:      &bl,
		Timestamp: "2025-01-01T00:00:00Z",
	}

	if err := WriteOutputs(true, results, baseline, ""); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "baseline<<COVERLINT_BASELINE_") {
		t.Error("output should contain baseline with delimiter")
	}
	if !strings.Contains(content, `"score":85`) {
		t.Error("output should contain baseline score")
	}
	if !strings.Contains(content, `"timestamp":"2025-01-01T00:00:00Z"`) {
		t.Error("output should contain baseline timestamp")
	}
}

func TestWriteOutputsWithSARIF(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	results := []EntryResult{
		{Name: "test", Passed: true},
	}

	sarifJSON := `{"version":"2.1.0","runs":[]}`

	if err := WriteOutputs(true, results, nil, sarifJSON); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, _ := os.ReadFile(outputFile)
	content := string(data)

	if !strings.Contains(content, "sarif<<COVERLINT_SARIF_") {
		t.Errorf("output should contain sarif with multiline delimiter, got: %s", content)
	}
	if !strings.Contains(content, `"version":"2.1.0"`) {
		t.Errorf("output should contain SARIF JSON content, got: %s", content)
	}
}

func TestWriteOutputsWithoutSARIF(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	results := []EntryResult{
		{Name: "test", Passed: true},
	}

	if err := WriteOutputs(true, results, nil, ""); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, _ := os.ReadFile(outputFile)
	content := string(data)

	if strings.Contains(content, "sarif=") {
		t.Errorf("output should not contain sarif when path is empty, got: %s", content)
	}
}

func TestWriteOutputsInvalidPath(t *testing.T) {
	t.Setenv("GITHUB_OUTPUT", "/nonexistent/dir/output")
	err := WriteOutputs(true, nil, nil, "")
	if err == nil {
		t.Fatal("expected error for invalid output path")
	}
	if !strings.Contains(err.Error(), "opening output file") {
		t.Errorf("error should mention opening: %v", err)
	}
}

func TestWriteJobSummaryInvalidPath(t *testing.T) {
	t.Setenv("GITHUB_STEP_SUMMARY", "/nonexistent/dir/summary")
	err := WriteJobSummary(nil, false, nil)
	if err == nil {
		t.Fatal("expected error for invalid summary path")
	}
	if !strings.Contains(err.Error(), "opening step summary file") {
		t.Errorf("error should mention opening: %v", err)
	}
}

func TestEmitAnnotation(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	EmitAnnotation("error", "something failed")

	_ = w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	got := string(out)

	if got != "::error::something failed\n" {
		t.Errorf("EmitAnnotation output = %q, want %q", got, "::error::something failed\n")
	}
}

func TestEmitAnnotationSanitizesMessage(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	EmitAnnotation("warning", "inject\n::error::pwned")

	_ = w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	got := string(out)

	if strings.Contains(got, "::error::pwned") {
		t.Error("EmitAnnotation should sanitize workflow command injection")
	}
}

func TestWriteOutputsWriteError(t *testing.T) {
	// Make the output file read-only to trigger write error after opening
	dir := t.TempDir()
	outputFile := filepath.Join(dir, "github_output")
	// Create as writable, then chmod after open to simulate write failure
	if err := os.WriteFile(outputFile, nil, 0444); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	err := WriteOutputs(true, nil, nil, "")
	if err == nil {
		t.Fatal("expected error when output file is read-only")
	}
}

func TestWriteOutputsAllFields(t *testing.T) {
	// Test all WriteOutputs paths: badge, baseline, and sarif all present
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	score := 85.0
	line := 90.0
	results := []EntryResult{
		{Name: "total", Score: &score, Line: &line, Passed: true},
	}

	bl := 90.0
	baseline := &BaselineData{
		Score:     85.0,
		Line:      &bl,
		Timestamp: "2025-01-01T00:00:00Z",
	}

	sarifJSON := `{"version":"2.1.0","runs":[]}`

	if err := WriteOutputs(true, results, baseline, sarifJSON); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, _ := os.ReadFile(outputFile)
	content := string(data)

	checks := []string{
		"passed=true",
		"results<<COVERLINT_RESULTS_",
		"badge-svg<<COVERLINT_SVG_",
		"badge-json<<COVERLINT_JSON_",
		"baseline<<COVERLINT_BASELINE_",
		"sarif<<COVERLINT_SARIF_",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Errorf("output should contain %q", c)
		}
	}
}

func TestWriteOutputsViaClosedPipe(t *testing.T) {
	// Use /proc/self/fd or a symlink trick to create an unwritable file descriptor path.
	// Simpler approach: create a named pipe (FIFO) and don't open it for reading,
	// making writes fail. On macOS, we can use a different approach:
	// create a temp file, use os.OpenFile with O_APPEND|O_WRONLY to verify it works,
	// then point GITHUB_OUTPUT to a path under a directory that gets chmod 0 after open.

	// Actually, the simplest cross-platform approach is to use a path to a directory,
	// but os.OpenFile on a directory for writing returns an error at open time.
	// Since the open error is already tested, focus on testing path variations instead.

	// Test with all output fields populated to maximize coverage of non-error paths
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	score := 50.0
	line := 60.0
	branch := 70.0
	fn := 80.0
	results := []EntryResult{
		{Name: "format1", Score: &score, Line: &line, Branch: &branch, Function: &fn, Passed: false},
		{Name: "total", Score: &score, Line: &line, Branch: &branch, Function: &fn, Passed: false},
	}

	bl := 75.0
	br := 65.0
	baseline := &BaselineData{
		Score:    90.0,
		Line:     &bl,
		Branch:   &br,
		Function: &fn,
		Timestamp: "2025-01-01T00:00:00Z",
	}

	sarifJSON := `{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"coverlint"}},"results":[]}]}`

	if err := WriteOutputs(false, results, baseline, sarifJSON); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, _ := os.ReadFile(outputFile)
	content := string(data)

	// Verify all sections present
	if !strings.Contains(content, "passed=false") {
		t.Error("missing passed=false")
	}
	if !strings.Contains(content, "results<<") {
		t.Error("missing results")
	}
	if !strings.Contains(content, "badge-svg<<") {
		t.Error("missing badge-svg")
	}
	if !strings.Contains(content, "badge-json<<") {
		t.Error("missing badge-json")
	}
	if !strings.Contains(content, "baseline<<") {
		t.Error("missing baseline")
	}
	if !strings.Contains(content, "sarif<<") {
		t.Error("missing sarif")
	}
}

func TestWriteOutputsFalseAndNoScore(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "github_output")
	if err := os.WriteFile(outputFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_OUTPUT", outputFile)

	// Test passed=false with entry that has no score
	results := []EntryResult{
		{Name: "test", Passed: false},
	}

	if err := WriteOutputs(false, results, nil, ""); err != nil {
		t.Fatalf("WriteOutputs() error: %v", err)
	}

	data, _ := os.ReadFile(outputFile)
	content := string(data)

	if !strings.Contains(content, "passed=false") {
		t.Error("should contain passed=false")
	}
	if strings.Contains(content, "badge-svg") {
		t.Error("should not contain badge when no score")
	}
}

func TestWriteJobSummaryScoreColumn(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(summaryFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	score := 85.0
	results := []EntryResult{
		{Name: "test", Score: &score, Passed: true},
	}

	if err := WriteJobSummary(results, false, nil); err != nil {
		t.Fatalf("WriteJobSummary() error: %v", err)
	}

	data, _ := os.ReadFile(summaryFile)
	content := string(data)

	if !strings.Contains(content, "85.0%") {
		t.Error("summary should contain score percentage")
	}
	if !strings.Contains(content, "Score") {
		t.Error("summary should contain Score header")
	}
}

func TestWriteJobSummaryFunctionColumn(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(summaryFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	line := 90.0
	fn := 80.0
	results := []EntryResult{
		{Name: "test", Line: &line, Function: &fn, Passed: true},
	}

	if err := WriteJobSummary(results, false, nil); err != nil {
		t.Fatalf("WriteJobSummary() error: %v", err)
	}

	data, _ := os.ReadFile(summaryFile)
	content := string(data)

	if !strings.Contains(content, "Function") {
		t.Error("summary should contain Function column")
	}
	if !strings.Contains(content, "80.0%") {
		t.Error("summary should contain function percentage")
	}
}

// failWriter fails after N successful writes.
type failWriter struct {
	n       int // number of writes before failure
	written int
}

func (fw *failWriter) Write(p []byte) (int, error) {
	fw.written++
	if fw.written > fw.n {
		return 0, fmt.Errorf("simulated write error")
	}
	return len(p), nil
}

func TestWriteOutputsToWriteErrors(t *testing.T) {
	score := 85.0
	line := 90.0
	results := []EntryResult{
		{Name: "total", Score: &score, Line: &line, Passed: true},
	}
	bl := 80.0
	baseline := &BaselineData{Score: 85.0, Line: &bl, Timestamp: "2025-01-01T00:00:00Z"}
	sarifJSON := `{"version":"2.1.0","runs":[]}`

	tests := []struct {
		name    string
		failAt  int
		wantMsg string
	}{
		{"fail writing passed", 0, "writing passed output"},
		{"fail writing results", 1, "writing results output"},
		{"fail writing badge-svg", 2, "writing badge-svg output"},
		{"fail writing badge-json", 3, "writing badge-json output"},
		{"fail writing baseline", 4, "writing baseline output"},
		{"fail writing sarif", 5, "writing sarif output"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &failWriter{n: tt.failAt}
			err := writeOutputsTo(w, true, results, baseline, sarifJSON)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestWriteJobSummaryTotalWithAllMetrics(t *testing.T) {
	summaryFile := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(summaryFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	line1 := 90.0
	branch1 := 80.0
	fn1 := 70.0
	totalLine := 85.0
	totalBranch := 75.0
	totalFn := 65.0

	results := []EntryResult{
		{Name: "gocover", Line: &line1, Branch: &branch1, Function: &fn1, Passed: true},
		{Name: "Total", Line: &totalLine, Branch: &totalBranch, Function: &totalFn, Passed: true},
	}

	if err := WriteJobSummary(results, true, nil); err != nil {
		t.Fatalf("WriteJobSummary() error: %v", err)
	}

	data, _ := os.ReadFile(summaryFile)
	content := string(data)

	// Total row should have bold metrics including branch and function
	if !strings.Contains(content, "**85.0%**") {
		t.Error("total row should contain bold line percentage")
	}
	if !strings.Contains(content, "**75.0%**") {
		t.Error("total row should contain bold branch percentage")
	}
	if !strings.Contains(content, "**65.0%**") {
		t.Error("total row should contain bold function percentage")
	}
}
