package coverage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Run is the main entry point for the coverage action.
func Run() error {
	inp, err := ParseInputs()
	if err != nil {
		return err
	}

	reportPath := filepath.Join(inp.WorkDir, inp.Path)
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return fmt.Errorf("reading coverage file %q: %w", reportPath, err)
	}

	parser, err := getParser(inp.Format)
	if err != nil {
		return err
	}

	result, err := parser(data)
	if err != nil {
		return fmt.Errorf("parsing coverage: %w", err)
	}
	result.Name = inp.Name

	cr := CheckThresholds(result, &inp.Threshold)

	entryResult := EntryResult{
		Name:   inp.Name,
		Passed: cr.Passed,
	}
	if result.Line != nil {
		pct := result.Line.Pct()
		entryResult.Line = &pct
	}
	if result.Branch != nil {
		pct := result.Branch.Pct()
		entryResult.Branch = &pct
	}
	if result.Function != nil {
		pct := result.Function.Pct()
		entryResult.Function = &pct
	}

	results := []EntryResult{entryResult}

	for _, s := range cr.Skipped {
		EmitAnnotation("notice", fmt.Sprintf("%s: %s threshold configured but not reported by %s format — skipped",
			s.Entry, s.Metric, inp.Format))
	}

	// Emit annotations
	for _, v := range cr.Violations {
		level := "error"
		if !inp.FailOnError {
			level = "warning"
		}
		EmitAnnotation(level, FormatViolation(v))
	}

	if cr.Passed {
		var parts []string
		if entryResult.Line != nil {
			parts = append(parts, fmt.Sprintf("line %.1f%%", *entryResult.Line))
		}
		if entryResult.Branch != nil {
			parts = append(parts, fmt.Sprintf("branch %.1f%%", *entryResult.Branch))
		}
		if entryResult.Function != nil {
			parts = append(parts, fmt.Sprintf("function %.1f%%", *entryResult.Function))
		}
		msg := fmt.Sprintf("%s: %s — all thresholds met", inp.Name, strings.Join(parts, ", "))
		EmitAnnotation("notice", msg)
	}

	// Write job summary and outputs
	if err := WriteJobSummary(results); err != nil {
		EmitAnnotation("warning", fmt.Sprintf("failed to write job summary: %v", err))
	}

	if err := WriteOutputs(cr.Passed, results); err != nil {
		EmitAnnotation("warning", fmt.Sprintf("failed to write outputs: %v", err))
	}

	if !cr.Passed && inp.FailOnError {
		return fmt.Errorf("coverage below threshold for: %s", inp.Name)
	}

	return nil
}
