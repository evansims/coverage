package coverage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseJacoco(t *testing.T) {
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
			fixture:      "jacoco/basic.xml",
			wantLine:     &Metric{Hit: 170, Total: 178},
			wantBranch:   &Metric{Hit: 35, Total: 40},
			wantFunction: &Metric{Hit: 32, Total: 35},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", "testdata", tt.fixture))
			if err != nil {
				t.Fatal(err)
			}
			result, err := parseJacoco(data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseJacoco() error = %v, wantErr %v", err, tt.wantErr)
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
