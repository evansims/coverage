package coverage

import "testing"

func floatPtr(f float64) *float64 { return &f }

func TestCheckThresholds(t *testing.T) {
	tests := []struct {
		name           string
		result         CoverageResult
		threshold      Threshold
		wantPassed     bool
		wantViolations int
	}{
		{
			name: "all pass",
			result: CoverageResult{
				Name: "test",
				Line: &Metric{Hit: 85, Total: 100},
			},
			threshold:      Threshold{Line: floatPtr(80)},
			wantPassed:     true,
			wantViolations: 0,
		},
		{
			name: "line fails",
			result: CoverageResult{
				Name: "test",
				Line: &Metric{Hit: 70, Total: 100},
			},
			threshold:      Threshold{Line: floatPtr(80)},
			wantPassed:     false,
			wantViolations: 1,
		},
		{
			name: "multiple failures",
			result: CoverageResult{
				Name:     "test",
				Line:     &Metric{Hit: 70, Total: 100},
				Branch:   &Metric{Hit: 50, Total: 100},
				Function: &Metric{Hit: 60, Total: 100},
			},
			threshold:      Threshold{Line: floatPtr(80), Branch: floatPtr(70), Function: floatPtr(80)},
			wantPassed:     false,
			wantViolations: 3,
		},
		{
			name: "metric nil but threshold set - skip no violation",
			result: CoverageResult{
				Name: "test",
				Line: &Metric{Hit: 90, Total: 100},
			},
			threshold:      Threshold{Line: floatPtr(80), Branch: floatPtr(70)},
			wantPassed:     true,
			wantViolations: 0,
		},
		{
			name: "exactly at threshold passes",
			result: CoverageResult{
				Name: "test",
				Line: &Metric{Hit: 80, Total: 100},
			},
			threshold:      Threshold{Line: floatPtr(80)},
			wantPassed:     true,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, violations := CheckThresholds(&tt.result, &tt.threshold)
			if passed != tt.wantPassed {
				t.Errorf("passed = %v, want %v", passed, tt.wantPassed)
			}
			if len(violations) != tt.wantViolations {
				t.Errorf("got %d violations, want %d: %+v", len(violations), tt.wantViolations, violations)
			}
		})
	}
}
