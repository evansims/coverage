package coverage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCobertura(t *testing.T) {
	tests := []struct {
		name       string
		fixture    string
		wantLine   *Metric
		wantBranch *Metric
		wantErr    bool
	}{
		{
			name:       "basic coverage",
			fixture:    "cobertura/basic.xml",
			wantLine:   &Metric{Hit: 170, Total: 200},
			wantBranch: &Metric{Hit: 35, Total: 50},
		},
		{
			name:       "no branches",
			fixture:    "cobertura/no_branches.xml",
			wantLine:   &Metric{Hit: 90, Total: 100},
			wantBranch: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", "testdata", tt.fixture))
			if err != nil {
				t.Fatal(err)
			}
			result, err := parseCobertura(data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseCobertura() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			assertMetric(t, "line", result.Line, tt.wantLine)
			assertMetric(t, "branch", result.Branch, tt.wantBranch)
			if result.Function != nil {
				t.Errorf("function: expected nil for cobertura, got %+v", result.Function)
			}
		})
	}
}
