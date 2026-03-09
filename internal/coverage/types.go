package coverage

// Weights defines relative weights for computing a weighted coverage score.
// When a metric is unavailable, its weight is redistributed proportionally.
type Weights struct {
	Line     float64
	Branch   float64
	Function float64
}

// DefaultWeights returns the default coverage score weights.
func DefaultWeights() Weights {
	return Weights{Line: 50, Branch: 30, Function: 20}
}

// Threshold defines coverage percentage thresholds.
type Threshold struct {
	Line        *float64
	Branch      *float64
	Function    *float64
	MinCoverage *float64 // weighted score minimum
	Weights     Weights
}

// CoverageResult holds parsed coverage metrics from a report.
type CoverageResult struct {
	Line     *Metric
	Branch   *Metric
	Function *Metric
	Files    []FileCoverage

	// Detail fields for merge support — populated by parsers.
	// Only one of FileDetails or BlockDetails will be set.
	FileDetails  map[string]*FileLineDetail            // file → detail (line-based formats)
	BlockDetails map[string]map[string]*BlockEntry      // file → block_key → entry (gocover)
}

// FileCoverage holds per-file coverage metrics for suggestions.
type FileCoverage struct {
	Path     string
	Line     *Metric
	Branch   *Metric
	Function *Metric
}

// FileLineDetail holds per-line coverage data for accurate merge operations.
// Used by LCOV, Cobertura, Clover, and JaCoCo parsers.
type FileLineDetail struct {
	Lines     map[int]int64    // line number → execution count
	Branches  map[string]int64 // branch key → taken count (format-specific key)
	Functions map[string]int64 // function name → execution count
}

// BlockEntry holds coverage data for a gocover statement block.
type BlockEntry struct {
	Stmts int64
	Count int64
}

// Metric holds hit/total counts for a coverage metric.
type Metric struct {
	Hit   int64
	Total int64
}

// Pct returns the coverage percentage, or 0 if total is 0.
func (m *Metric) Pct() float64 {
	if m.Total == 0 {
		return 0
	}
	return float64(m.Hit) / float64(m.Total) * 100
}

// CoverageScore computes a weighted coverage score from available metrics.
// Weights for unavailable metrics are redistributed proportionally.
// Returns 0 if no metrics are available.
func CoverageScore(line, branch, function *float64, w Weights) float64 {
	type entry struct {
		pct    float64
		weight float64
	}
	var entries []entry
	if line != nil && w.Line > 0 {
		entries = append(entries, entry{*line, w.Line})
	}
	if branch != nil && w.Branch > 0 {
		entries = append(entries, entry{*branch, w.Branch})
	}
	if function != nil && w.Function > 0 {
		entries = append(entries, entry{*function, w.Function})
	}
	if len(entries) == 0 {
		return 0
	}
	var totalWeight, weighted float64
	for _, e := range entries {
		totalWeight += e.weight
		weighted += e.pct * e.weight
	}
	return weighted / totalWeight
}

// Violation records a threshold that was not met.
type Violation struct {
	Entry    string
	Metric   string
	Actual   float64
	Required float64
}

// SkippedThreshold records a threshold that was configured but had no
// corresponding metric data from the coverage report.
type SkippedThreshold struct {
	Entry  string
	Metric string
}
