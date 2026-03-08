package coverage

import "fmt"

func CheckThresholds(result *CoverageResult, threshold *Threshold) (bool, []Violation) {
	var violations []Violation

	violations = checkMetric(violations, result.Name, "line", result.Line, threshold.Line)
	violations = checkMetric(violations, result.Name, "branch", result.Branch, threshold.Branch)
	violations = checkMetric(violations, result.Name, "function", result.Function, threshold.Function)

	return len(violations) == 0, violations
}

func checkMetric(violations []Violation, entry, metric string, m *Metric, threshold *float64) []Violation {
	if threshold == nil || m == nil {
		return violations
	}

	pct := m.Pct()
	if pct < *threshold {
		violations = append(violations, Violation{
			Entry:    entry,
			Metric:   metric,
			Actual:   pct,
			Required: *threshold,
		})
	}

	return violations
}

func FormatViolation(v Violation) string {
	return fmt.Sprintf("%s: %s coverage %.1f%% < %.1f%% required",
		v.Entry, v.Metric, v.Actual, v.Required)
}
