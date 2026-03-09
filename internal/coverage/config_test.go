package coverage

import (
	"strings"
	"testing"
)

func TestParseInputs(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantErr     string
		wantFormats []string
	}{
		{
			name: "valid minimal",
			env: map[string]string{
				"INPUT_PATH":           "cover.out",
				"INPUT_FORMAT":         "gocover",
				"INPUT_MIN-LINE": "80",
			},
			wantFormats: []string{"gocover"},
		},
		{
			name: "valid all thresholds",
			env: map[string]string{
				"INPUT_PATH":               "lcov.info",
				"INPUT_FORMAT":             "lcov",
					"INPUT_MIN-LINE":     "80",
				"INPUT_MIN-BRANCH":   "70",
				"INPUT_MIN-FUNCTION": "75",
			},
			wantFormats: []string{"lcov"},
		},
		{
			name: "path optional",
			env: map[string]string{
				"INPUT_FORMAT":         "lcov",
				"INPUT_MIN-LINE": "80",
			},
			wantFormats: []string{"lcov"},
		},
		{
			name: "multiple formats",
			env: map[string]string{
				"INPUT_FORMAT":         "gocover,lcov",
				"INPUT_MIN-LINE": "80",
			},
			wantFormats: []string{"gocover", "lcov"},
		},
		{
			name: "multiple formats with spaces",
			env: map[string]string{
				"INPUT_FORMAT":         "gocover, lcov, cobertura",
				"INPUT_MIN-LINE": "80",
			},
			wantFormats: []string{"gocover", "lcov", "cobertura"},
		},
		{
			name: "multiple formats newline-separated",
			env: map[string]string{
				"INPUT_FORMAT":         "gocover\nlcov\ncobertura",
				"INPUT_MIN-LINE": "80",
			},
			wantFormats: []string{"gocover", "lcov", "cobertura"},
		},
		{
			name: "mixed newlines and commas",
			env: map[string]string{
				"INPUT_FORMAT":         "gocover,lcov\ncobertura",
				"INPUT_MIN-LINE": "80",
			},
			wantFormats: []string{"gocover", "lcov", "cobertura"},
		},
		{
			name: "format auto-detected when omitted",
			env: map[string]string{
				"INPUT_PATH":           "cover.out",
				"INPUT_MIN-LINE": "80",
			},
			wantFormats: formatOrder,
		},
		{
			name: "invalid format",
			env: map[string]string{
				"INPUT_PATH":           "cover.out",
				"INPUT_FORMAT":         "invalid",
				"INPUT_MIN-LINE": "80",
			},
			wantErr: "not valid",
		},
		{
			name: "one invalid in multi-format",
			env: map[string]string{
				"INPUT_FORMAT":         "gocover,invalid",
				"INPUT_MIN-LINE": "80",
			},
			wantErr: "not valid",
		},
		{
			name: "no thresholds is valid",
			env: map[string]string{
				"INPUT_PATH":   "cover.out",
				"INPUT_FORMAT": "gocover",
			},
			wantFormats: []string{"gocover"},
		},
		{
			name: "negative threshold",
			env: map[string]string{
				"INPUT_PATH":           "cover.out",
				"INPUT_FORMAT":         "lcov",
				"INPUT_MIN-LINE": "-5",
			},
			wantErr: "between 0 and 100",
		},
		{
			name: "threshold over 100",
			env: map[string]string{
				"INPUT_PATH":           "cover.out",
				"INPUT_FORMAT":         "lcov",
				"INPUT_MIN-LINE": "200",
			},
			wantErr: "between 0 and 100",
		},
		{
			name: "non-numeric threshold",
			env: map[string]string{
				"INPUT_PATH":           "cover.out",
				"INPUT_FORMAT":         "lcov",
				"INPUT_MIN-LINE": "abc",
			},
			wantErr: "not a valid number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all input env vars
			for _, key := range []string{
				"INPUT_PATH", "INPUT_FORMAT",
				"INPUT_WORKING-DIRECTORY", "INPUT_FAIL-ON-ERROR",
				"INPUT_MIN-COVERAGE", "INPUT_MIN-LINE", "INPUT_MIN-BRANCH", "INPUT_MIN-FUNCTION",
				"INPUT_SUGGESTIONS",
			} {
				t.Setenv(key, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			inp, err := ParseInputs()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if inp.Path != tt.env["INPUT_PATH"] {
				t.Errorf("path = %q, want %q", inp.Path, tt.env["INPUT_PATH"])
			}
			if len(inp.Formats) != len(tt.wantFormats) {
				t.Fatalf("formats = %v, want %v", inp.Formats, tt.wantFormats)
			}
			for i, f := range inp.Formats {
				if f != tt.wantFormats[i] {
					t.Errorf("formats[%d] = %q, want %q", i, f, tt.wantFormats[i])
				}
			}
		})
	}
}

func TestParseInputsAutoFormat(t *testing.T) {
	for _, key := range []string{
		"INPUT_PATH", "INPUT_FORMAT",
		"INPUT_WORKING-DIRECTORY", "INPUT_FAIL-ON-ERROR",
		"INPUT_MIN-COVERAGE", "INPUT_MIN-LINE", "INPUT_MIN-BRANCH", "INPUT_MIN-FUNCTION",
	} {
		t.Setenv(key, "")
	}

	t.Setenv("INPUT_PATH", "cover.out")
	t.Setenv("INPUT_MIN-LINE", "80")

	inp, err := ParseInputs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !inp.AutoFormat {
		t.Error("expected AutoFormat to be true when format is omitted")
	}
	if len(inp.Formats) != len(formatOrder) {
		t.Errorf("expected %d formats, got %d", len(formatOrder), len(inp.Formats))
	}

	// Verify explicit format sets AutoFormat to false
	t.Setenv("INPUT_FORMAT", "lcov")
	inp, err = ParseInputs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inp.AutoFormat {
		t.Error("expected AutoFormat to be false when format is specified")
	}
}

func TestParseInputsDefaults(t *testing.T) {
	for _, key := range []string{
		"INPUT_PATH", "INPUT_FORMAT",
		"INPUT_WORKING-DIRECTORY", "INPUT_FAIL-ON-ERROR",
		"INPUT_MIN-COVERAGE", "INPUT_MIN-LINE", "INPUT_MIN-BRANCH", "INPUT_MIN-FUNCTION",
	} {
		t.Setenv(key, "")
	}

	t.Setenv("INPUT_PATH", "cover.out")
	t.Setenv("INPUT_FORMAT", "gocover")
	t.Setenv("INPUT_MIN-LINE", "80")

	inp, err := ParseInputs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inp.WorkDir != "." {
		t.Errorf("workdir should default to '.', got %q", inp.WorkDir)
	}
	if !inp.FailOnError {
		t.Error("fail-on-error should default to true")
	}
}

func TestParseInputsMinCoverage(t *testing.T) {
	clear := func(t *testing.T) {
		t.Helper()
		for _, key := range []string{
			"INPUT_PATH", "INPUT_FORMAT",
			"INPUT_WORKING-DIRECTORY", "INPUT_FAIL-ON-ERROR",
			"INPUT_MIN-COVERAGE", "INPUT_MIN-LINE", "INPUT_MIN-BRANCH", "INPUT_MIN-FUNCTION",
			"INPUT_WEIGHT-LINE", "INPUT_WEIGHT-BRANCH", "INPUT_WEIGHT-FUNCTION",
			"INPUT_SUGGESTIONS",
		} {
			t.Setenv(key, "")
		}
	}

	t.Run("sets weighted score threshold only", func(t *testing.T) {
		clear(t)
		t.Setenv("INPUT_FORMAT", "gocover")
		t.Setenv("INPUT_MIN-COVERAGE", "80")

		inp, err := ParseInputs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// min-coverage sets the weighted score threshold, not individual metrics
		if inp.Threshold.MinCoverage == nil || *inp.Threshold.MinCoverage != 80 {
			t.Errorf("MinCoverage = %v, want 80", inp.Threshold.MinCoverage)
		}
		// Individual metrics should remain nil
		if inp.Threshold.Line != nil {
			t.Errorf("Line = %v, want nil", inp.Threshold.Line)
		}
		if inp.Threshold.Branch != nil {
			t.Errorf("Branch = %v, want nil", inp.Threshold.Branch)
		}
		if inp.Threshold.Function != nil {
			t.Errorf("Function = %v, want nil", inp.Threshold.Function)
		}
	})

	t.Run("min-coverage with individual hard floors", func(t *testing.T) {
		clear(t)
		t.Setenv("INPUT_FORMAT", "gocover")
		t.Setenv("INPUT_MIN-COVERAGE", "80")
		t.Setenv("INPUT_MIN-BRANCH", "60")

		inp, err := ParseInputs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inp.Threshold.MinCoverage == nil || *inp.Threshold.MinCoverage != 80 {
			t.Errorf("MinCoverage = %v, want 80", inp.Threshold.MinCoverage)
		}
		if inp.Threshold.Branch == nil || *inp.Threshold.Branch != 60 {
			t.Errorf("Branch = %v, want 60", inp.Threshold.Branch)
		}
		// Line and Function should remain nil (no individual floor set)
		if inp.Threshold.Line != nil {
			t.Errorf("Line = %v, want nil", inp.Threshold.Line)
		}
		if inp.Threshold.Function != nil {
			t.Errorf("Function = %v, want nil", inp.Threshold.Function)
		}
	})

	t.Run("individual thresholds without min-coverage", func(t *testing.T) {
		clear(t)
		t.Setenv("INPUT_FORMAT", "gocover")
		t.Setenv("INPUT_MIN-LINE", "90")

		inp, err := ParseInputs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inp.Threshold.MinCoverage != nil {
			t.Errorf("MinCoverage = %v, want nil", inp.Threshold.MinCoverage)
		}
		if inp.Threshold.Line == nil || *inp.Threshold.Line != 90 {
			t.Errorf("Line = %v, want 90", inp.Threshold.Line)
		}
		if inp.Threshold.Branch != nil {
			t.Errorf("Branch = %v, want nil", inp.Threshold.Branch)
		}
		if inp.Threshold.Function != nil {
			t.Errorf("Function = %v, want nil", inp.Threshold.Function)
		}
	})

	t.Run("invalid min-coverage", func(t *testing.T) {
		clear(t)
		t.Setenv("INPUT_FORMAT", "gocover")
		t.Setenv("INPUT_MIN-COVERAGE", "abc")

		_, err := ParseInputs()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "min-coverage") {
			t.Errorf("error %q should mention min-coverage", err.Error())
		}
	})
}

func TestParseInputsWeights(t *testing.T) {
	clear := func(t *testing.T) {
		t.Helper()
		for _, key := range []string{
			"INPUT_PATH", "INPUT_FORMAT",
			"INPUT_WORKING-DIRECTORY", "INPUT_FAIL-ON-ERROR",
			"INPUT_MIN-COVERAGE", "INPUT_MIN-LINE", "INPUT_MIN-BRANCH", "INPUT_MIN-FUNCTION",
			"INPUT_WEIGHT-LINE", "INPUT_WEIGHT-BRANCH", "INPUT_WEIGHT-FUNCTION",
			"INPUT_SUGGESTIONS",
		} {
			t.Setenv(key, "")
		}
	}

	t.Run("defaults", func(t *testing.T) {
		clear(t)
		t.Setenv("INPUT_FORMAT", "gocover")

		inp, err := ParseInputs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		dw := DefaultWeights()
		if inp.Threshold.Weights.Line != dw.Line {
			t.Errorf("weight-line = %v, want %v", inp.Threshold.Weights.Line, dw.Line)
		}
		if inp.Threshold.Weights.Branch != dw.Branch {
			t.Errorf("weight-branch = %v, want %v", inp.Threshold.Weights.Branch, dw.Branch)
		}
		if inp.Threshold.Weights.Function != dw.Function {
			t.Errorf("weight-function = %v, want %v", inp.Threshold.Weights.Function, dw.Function)
		}
	})

	t.Run("custom weights", func(t *testing.T) {
		clear(t)
		t.Setenv("INPUT_FORMAT", "gocover")
		t.Setenv("INPUT_WEIGHT-LINE", "100")
		t.Setenv("INPUT_WEIGHT-BRANCH", "0")
		t.Setenv("INPUT_WEIGHT-FUNCTION", "0")

		inp, err := ParseInputs()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inp.Threshold.Weights.Line != 100 {
			t.Errorf("weight-line = %v, want 100", inp.Threshold.Weights.Line)
		}
		if inp.Threshold.Weights.Branch != 0 {
			t.Errorf("weight-branch = %v, want 0", inp.Threshold.Weights.Branch)
		}
		if inp.Threshold.Weights.Function != 0 {
			t.Errorf("weight-function = %v, want 0", inp.Threshold.Weights.Function)
		}
	})

	t.Run("invalid weight", func(t *testing.T) {
		clear(t)
		t.Setenv("INPUT_FORMAT", "gocover")
		t.Setenv("INPUT_WEIGHT-LINE", "abc")

		_, err := ParseInputs()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "weight-line") {
			t.Errorf("error %q should mention weight-line", err.Error())
		}
	})
}

