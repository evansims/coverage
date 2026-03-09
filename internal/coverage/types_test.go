package coverage

import (
	"math"
	"testing"
)

func TestCoverageScore(t *testing.T) {
	tests := []struct {
		name     string
		line     *float64
		branch   *float64
		function *float64
		weights  Weights
		want     float64
	}{
		{
			name:     "all metrics with default weights",
			line:     floatPtr(80),
			branch:   floatPtr(60),
			function: floatPtr(90),
			weights:  DefaultWeights(), // 50/30/20
			want:     76, // (80*50 + 60*30 + 90*20) / 100
		},
		{
			name:     "line only redistributes weight",
			line:     floatPtr(80),
			branch:   nil,
			function: nil,
			weights:  DefaultWeights(),
			want:     80, // only line available, gets all weight
		},
		{
			name:     "line and branch only",
			line:     floatPtr(80),
			branch:   floatPtr(60),
			function: nil,
			weights:  DefaultWeights(),
			want:     72.5, // (80*50 + 60*30) / 80
		},
		{
			name:     "zero weight for branch excluded",
			line:     floatPtr(80),
			branch:   floatPtr(60),
			function: floatPtr(90),
			weights:  Weights{Line: 100, Branch: 0, Function: 0},
			want:     80, // only line counts
		},
		{
			name:     "equal weights",
			line:     floatPtr(80),
			branch:   floatPtr(60),
			function: floatPtr(100),
			weights:  Weights{Line: 1, Branch: 1, Function: 1},
			want:     80, // (80+60+100)/3
		},
		{
			name:     "no metrics available",
			line:     nil,
			branch:   nil,
			function: nil,
			weights:  DefaultWeights(),
			want:     0,
		},
		{
			name:     "100% across all metrics",
			line:     floatPtr(100),
			branch:   floatPtr(100),
			function: floatPtr(100),
			weights:  DefaultWeights(),
			want:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CoverageScore(tt.line, tt.branch, tt.function, tt.weights)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("CoverageScore() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}

func TestMetricPct(t *testing.T) {
	tests := []struct {
		name string
		m    Metric
		want float64
	}{
		{"normal", Metric{Hit: 80, Total: 100}, 80.0},
		{"zero total", Metric{Hit: 0, Total: 0}, 0.0},
		{"all hit", Metric{Hit: 50, Total: 50}, 100.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.Pct()
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("Pct() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}
