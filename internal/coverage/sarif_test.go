package coverage

import (
	"fmt"
	"testing"
)

func TestGenerateSARIF_WithLineDetails(t *testing.T) {
	fileDetails := map[string]*FileLineDetail{
		"src/main.go": {
			Lines: map[int]int64{
				1:  1,
				2:  1,
				3:  0, // uncovered
				4:  1,
				5:  0, // uncovered
				6:  1,
				7:  1,
				8:  1,
				9:  1,
				10: 1,
			},
		},
	}

	doc := GenerateSARIF(fileDetails, nil, defaultSARIFMaxResults)

	if doc.Version != "2.1.0" {
		t.Errorf("version = %q, want %q", doc.Version, "2.1.0")
	}
	if doc.Schema != "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json" {
		t.Errorf("unexpected schema: %s", doc.Schema)
	}
	if len(doc.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(doc.Runs))
	}

	run := doc.Runs[0]
	if run.Tool.Driver.Name != "coverlint" {
		t.Errorf("driver name = %q, want %q", run.Tool.Driver.Name, "coverlint")
	}
	if len(run.Tool.Driver.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(run.Tool.Driver.Rules))
	}

	// Should have 2 uncovered lines (lines 3 and 5)
	if len(run.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(run.Results))
	}

	// Results should be sorted by path and line
	r0 := run.Results[0]
	if r0.RuleID != "coverage/uncovered-line" {
		t.Errorf("result[0] ruleId = %q, want %q", r0.RuleID, "coverage/uncovered-line")
	}
	loc0 := r0.Locations[0].PhysicalLocation
	if loc0.ArtifactLocation.URI != "src/main.go" {
		t.Errorf("result[0] uri = %q, want %q", loc0.ArtifactLocation.URI, "src/main.go")
	}
	if loc0.Region == nil || loc0.Region.StartLine != 3 {
		t.Errorf("result[0] startLine = %v, want 3", loc0.Region)
	}

	r1 := run.Results[1]
	loc1 := r1.Locations[0].PhysicalLocation
	if loc1.Region == nil || loc1.Region.StartLine != 5 {
		t.Errorf("result[1] startLine = %v, want 5", loc1.Region)
	}
}

func TestGenerateSARIF_WithBlockDetails(t *testing.T) {
	blockDetails := map[string]map[string]*BlockEntry{
		"pkg/handler.go": {
			"pkg/handler.go:5.1,10.1":  {Stmts: 3, Count: 0}, // uncovered
			"pkg/handler.go:15.1,20.1": {Stmts: 2, Count: 5}, // covered
			"pkg/handler.go:25.1,30.1": {Stmts: 4, Count: 0}, // uncovered
		},
	}

	doc := GenerateSARIF(nil, blockDetails, defaultSARIFMaxResults)

	if len(doc.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(doc.Runs))
	}

	run := doc.Runs[0]
	// Should have 2 uncovered blocks
	if len(run.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(run.Results))
	}

	// Results should be sorted
	r0 := run.Results[0]
	if r0.RuleID != "coverage/uncovered-block" {
		t.Errorf("result[0] ruleId = %q, want %q", r0.RuleID, "coverage/uncovered-block")
	}
	loc0 := r0.Locations[0].PhysicalLocation
	if loc0.Region == nil || loc0.Region.StartLine != 5 || loc0.Region.EndLine != 10 {
		t.Errorf("result[0] region = %v, want startLine=5, endLine=10", loc0.Region)
	}

	r1 := run.Results[1]
	loc1 := r1.Locations[0].PhysicalLocation
	if loc1.Region == nil || loc1.Region.StartLine != 25 || loc1.Region.EndLine != 30 {
		t.Errorf("result[1] region = %v, want startLine=25, endLine=30", loc1.Region)
	}
}

func TestGenerateSARIF_Empty(t *testing.T) {
	doc := GenerateSARIF(nil, nil, defaultSARIFMaxResults)

	if len(doc.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(doc.Runs))
	}
	if len(doc.Runs[0].Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(doc.Runs[0].Results))
	}
}

func TestGenerateSARIF_ResultsCapped(t *testing.T) {
	// Create file details with more uncovered lines than the cap
	maxResults := 100
	lines := make(map[int]int64)
	for i := 1; i <= maxResults+50; i++ {
		lines[i] = 0 // all uncovered
	}
	fileDetails := map[string]*FileLineDetail{
		"big.go": {Lines: lines},
	}

	doc := GenerateSARIF(fileDetails, nil, maxResults)

	if len(doc.Runs[0].Results) != maxResults {
		t.Errorf("results = %d, want %d (capped)", len(doc.Runs[0].Results), maxResults)
	}
}

func TestGenerateSARIF_MultipleFiles(t *testing.T) {
	fileDetails := map[string]*FileLineDetail{
		"a.go": {Lines: map[int]int64{1: 0}},
		"b.go": {Lines: map[int]int64{1: 1, 2: 0}},
	}

	doc := GenerateSARIF(fileDetails, nil, defaultSARIFMaxResults)

	// Should have 2 results (one from a.go line 1, one from b.go line 2)
	if len(doc.Runs[0].Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(doc.Runs[0].Results))
	}

	// Should be sorted: a.go before b.go
	if doc.Runs[0].Results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI != "a.go" {
		t.Error("results should be sorted by path, expected a.go first")
	}
}

func TestParseBlockRange(t *testing.T) {
	tests := []struct {
		key       string
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		{"file.go:5.1,10.1", 5, 10, false},
		{"pkg/handler.go:15.3,20.8", 15, 20, false},
		{"a.go:1.1,1.1", 1, 1, false},
		{"invalid", 0, 0, true},
		{"file.go:bad", 0, 0, true},
		{"file.go:5.1", 0, 0, true},            // missing comma
		{"file.go:abc.1,10.1", 0, 0, true},     // non-numeric start
		{"file.go:5.1,abc.1", 0, 0, true},      // non-numeric end
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			start, end, err := parseBlockRange(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBlockRange(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if start != tt.wantStart || end != tt.wantEnd {
					t.Errorf("parseBlockRange(%q) = (%d, %d), want (%d, %d)", tt.key, start, end, tt.wantStart, tt.wantEnd)
				}
			}
		})
	}
}

func TestSanitizeSARIFPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"src/main.go", "src/main.go"},
		{"../../../etc/passwd", "etc/passwd"},
		{"src/../../../etc/passwd", "src/etc/passwd"},
		{"/absolute/path.go", "absolute/path.go"},
		{"", "unknown"},
		{"a/b/../c", "a/b/c"},                       // .. stripped but other parts kept
		{"src\\windows\\path.go", "src/windows/path.go"}, // backslash normalized
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeSARIFPath(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeSARIFPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeSARIFPathLength(t *testing.T) {
	// Paths longer than 1024 chars should be truncated
	long := make([]byte, 2000)
	for i := range long {
		long[i] = 'a'
	}
	got := sanitizeSARIFPath(string(long))
	if len(got) != 1024 {
		t.Errorf("len = %d, want 1024", len(got))
	}
}

func TestBlockMessage(t *testing.T) {
	tests := []struct {
		start int
		end   int
		want  string
	}{
		{5, 5, "Block at line 5 is not covered by tests"},
		{5, 10, "Block at lines 5-10 is not covered by tests"},
		{1, 1, "Block at line 1 is not covered by tests"},
	}
	for _, tt := range tests {
		got := blockMessage(tt.start, tt.end)
		if got != tt.want {
			t.Errorf("blockMessage(%d, %d) = %q, want %q", tt.start, tt.end, got, tt.want)
		}
	}
}

func TestGenerateSARIF_BlockSameLine(t *testing.T) {
	// Block where startLine == endLine should NOT set EndLine
	blockDetails := map[string]map[string]*BlockEntry{
		"pkg/handler.go": {
			"pkg/handler.go:5.1,5.10": {Stmts: 1, Count: 0}, // same line
		},
	}

	doc := GenerateSARIF(nil, blockDetails, defaultSARIFMaxResults)

	if len(doc.Runs[0].Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(doc.Runs[0].Results))
	}

	region := doc.Runs[0].Results[0].Locations[0].PhysicalLocation.Region
	if region.StartLine != 5 {
		t.Errorf("startLine = %d, want 5", region.StartLine)
	}
	if region.EndLine != 0 {
		t.Errorf("endLine = %d, want 0 (omitted for same line)", region.EndLine)
	}
}

func TestGenerateSARIF_FileWithNilLines(t *testing.T) {
	// File detail with nil Lines map should be skipped
	fileDetails := map[string]*FileLineDetail{
		"empty.go": {Lines: nil},
		"real.go":  {Lines: map[int]int64{1: 0}},
	}

	doc := GenerateSARIF(fileDetails, nil, defaultSARIFMaxResults)

	// Only real.go should produce a result
	if len(doc.Runs[0].Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(doc.Runs[0].Results))
	}
	uri := doc.Runs[0].Results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI
	if uri != "real.go" {
		t.Errorf("expected real.go, got %s", uri)
	}
}

func TestGenerateSARIF_BlockMultipleFiles(t *testing.T) {
	blockDetails := map[string]map[string]*BlockEntry{
		"b.go": {
			"b.go:10.1,20.1": {Stmts: 3, Count: 0},
		},
		"a.go": {
			"a.go:5.1,15.1": {Stmts: 2, Count: 0},
		},
	}

	doc := GenerateSARIF(nil, blockDetails, defaultSARIFMaxResults)

	if len(doc.Runs[0].Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(doc.Runs[0].Results))
	}
	// Should be sorted: a.go before b.go
	firstURI := doc.Runs[0].Results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI
	if firstURI != "a.go" {
		t.Errorf("expected a.go first (sorted), got %s", firstURI)
	}
}

func TestGenerateSARIF_ResultMessage(t *testing.T) {
	fileDetails := map[string]*FileLineDetail{
		"main.go": {Lines: map[int]int64{10: 0}},
	}

	doc := GenerateSARIF(fileDetails, nil, defaultSARIFMaxResults)

	if len(doc.Runs[0].Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(doc.Runs[0].Results))
	}

	msg := doc.Runs[0].Results[0].Message.Text
	expected := fmt.Sprintf("Line %d is not covered by tests", 10)
	if msg != expected {
		t.Errorf("message = %q, want %q", msg, expected)
	}
}
