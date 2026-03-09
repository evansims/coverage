package coverage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// EntryResult holds the coverage results for a single entry, formatted for output.
type EntryResult struct {
	Name     string   `json:"name"`
	Score    *float64 `json:"score,omitempty"`
	Line     *float64 `json:"line,omitempty"`
	Branch   *float64 `json:"branch,omitempty"`
	Function *float64 `json:"function,omitempty"`
	Passed   bool     `json:"passed"`
}

// EmitAnnotation writes a GitHub Actions workflow command to stdout.
// Message is sanitized to prevent workflow command injection.
func EmitAnnotation(level, message string) {
	fmt.Printf("::%s::%s\n", level, sanitizeWorkflowCommand(message))
}

// sanitizeWorkflowCommand strips characters that could inject additional
// GitHub Actions workflow commands into stdout.
func sanitizeWorkflowCommand(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "::", ": :")
	return s
}

// sanitizeMarkdown escapes characters that could break markdown table formatting.
func sanitizeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

// randomDelimiter generates a unique delimiter for GITHUB_OUTPUT multiline values.
// Uses crypto/rand to prevent injection via crafted coverage file paths that match
// the delimiter string.
func randomDelimiter(prefix string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use a long prefix unlikely to appear in output content.
		// This path should never be reached on any modern OS.
		return prefix + "_FALLBACK_9f8e7d6c5b4a3210"
	}
	return prefix + "_" + hex.EncodeToString(b)
}

// WriteJobSummary writes a markdown coverage table to $GITHUB_STEP_SUMMARY.
// When hasTotal is true, the last entry in results is rendered as a bold total
// footer row separated from the per-format rows above it.
func WriteJobSummary(results []EntryResult, hasTotal bool, suggestions []Suggestion) (err error) {
	summaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryPath == "" {
		return nil // not running in GitHub Actions
	}

	// Determine which metric columns have data across all results
	var hasLine, hasBranch, hasFunction bool
	for _, r := range results {
		if r.Line != nil {
			hasLine = true
		}
		if r.Branch != nil {
			hasBranch = true
		}
		if r.Function != nil {
			hasFunction = true
		}
	}

	var sb strings.Builder
	sb.WriteString("## Coverage Results\n\n")

	// Build header dynamically based on available metrics
	header := "| Name | Score"
	separator := "|------|-------"
	if hasLine {
		header += " | Line"
		separator += "|------"
	}
	if hasBranch {
		header += " | Branch"
		separator += "|--------"
	}
	if hasFunction {
		header += " | Function"
		separator += "|----------"
	}
	header += " | Status |\n"
	separator += "|--------|\n"
	sb.WriteString(header)
	sb.WriteString(separator)

	// Separate regular rows from total row
	regularRows := results
	var totalRow *EntryResult
	if hasTotal && len(results) > 1 {
		regularRows = results[:len(results)-1]
		last := results[len(results)-1]
		totalRow = &last
	}

	for _, r := range regularRows {
		status := "Pass"
		if !r.Passed {
			status = "**Fail**"
		}

		fmt.Fprintf(&sb, "| %s | %s", sanitizeMarkdown(r.Name), fmtPct(r.Score))
		if hasLine {
			fmt.Fprintf(&sb, " | %s", fmtPct(r.Line))
		}
		if hasBranch {
			fmt.Fprintf(&sb, " | %s", fmtPct(r.Branch))
		}
		if hasFunction {
			fmt.Fprintf(&sb, " | %s", fmtPct(r.Function))
		}
		fmt.Fprintf(&sb, " | %s |\n", status)
	}

	// Render total footer row with bold formatting
	if totalRow != nil {
		status := "**Pass**"
		if !totalRow.Passed {
			status = "**Fail**"
		}

		fmt.Fprintf(&sb, "| **%s** | **%s** ", sanitizeMarkdown(totalRow.Name), fmtPct(totalRow.Score))
		if hasLine {
			fmt.Fprintf(&sb, "| **%s** ", fmtPct(totalRow.Line))
		}
		if hasBranch {
			fmt.Fprintf(&sb, "| **%s** ", fmtPct(totalRow.Branch))
		}
		if hasFunction {
			fmt.Fprintf(&sb, "| **%s** ", fmtPct(totalRow.Function))
		}
		fmt.Fprintf(&sb, "| %s |\n", status)
	}

	sb.WriteString("\n")

	if suggestionsSection := FormatSuggestions(suggestions); suggestionsSection != "" {
		sb.WriteString(suggestionsSection)
	}

	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening step summary file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing step summary file: %w", cerr)
		}
	}()

	_, err = f.WriteString(sb.String())
	if err != nil {
		return fmt.Errorf("writing step summary: %w", err)
	}

	return nil
}

// WriteOutputs writes action outputs to $GITHUB_OUTPUT.
// Uses randomized multiline delimiter syntax for all multi-value outputs to
// prevent injection via crafted coverage file paths or content.
// Badge outputs are generated from the last entry's line coverage (the total).
// If baseline is non-nil, baseline JSON is written as a multiline output.
// If sarifJSON is non-empty, SARIF JSON is written as a multiline output.
func WriteOutputs(passed bool, results []EntryResult, baseline *BaselineData, sarifJSON string) (err error) {
	outputPath := os.Getenv("GITHUB_OUTPUT")
	if outputPath == "" {
		return nil
	}

	f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening output file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing output file: %w", cerr)
		}
	}()

	return writeOutputsTo(f, passed, results, baseline, sarifJSON)
}

// writeOutputsTo writes action outputs to the given writer.
// Extracted for testability — allows injecting a failing writer.
func writeOutputsTo(w io.Writer, passed bool, results []EntryResult, baseline *BaselineData, sarifJSON string) error {
	if _, err := fmt.Fprintf(w, "passed=%v\n", passed); err != nil {
		return fmt.Errorf("writing passed output: %w", err)
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("marshaling results: %w", err)
	}

	// Use randomized multiline delimiter syntax to prevent output injection
	delimiter := randomDelimiter("COVERLINT_RESULTS")
	if _, err = fmt.Fprintf(w, "results<<%s\n%s\n%s\n", delimiter, string(resultsJSON), delimiter); err != nil {
		return fmt.Errorf("writing results output: %w", err)
	}

	// Generate badge outputs from the total/last entry's coverage score
	if len(results) > 0 {
		total := results[len(results)-1]
		if total.Score != nil {
			svgDelimiter := randomDelimiter("COVERLINT_SVG")
			svg := GenerateBadgeSVG(*total.Score)
			if _, err = fmt.Fprintf(w, "badge-svg<<%s\n%s\n%s\n", svgDelimiter, svg, svgDelimiter); err != nil {
				return fmt.Errorf("writing badge-svg output: %w", err)
			}

			jsonDelimiter := randomDelimiter("COVERLINT_JSON")
			badgeJSON := GenerateBadgeJSON(*total.Score)
			if _, err = fmt.Fprintf(w, "badge-json<<%s\n%s\n%s\n", jsonDelimiter, badgeJSON, jsonDelimiter); err != nil {
				return fmt.Errorf("writing badge-json output: %w", err)
			}
		}
	}

	if baseline != nil {
		baselineJSON, merr := json.Marshal(baseline)
		if merr != nil {
			return fmt.Errorf("marshaling baseline: %w", merr)
		}
		baselineDelimiter := randomDelimiter("COVERLINT_BASELINE")
		if _, err = fmt.Fprintf(w, "baseline<<%s\n%s\n%s\n", baselineDelimiter, string(baselineJSON), baselineDelimiter); err != nil {
			return fmt.Errorf("writing baseline output: %w", err)
		}
	}

	if sarifJSON != "" {
		sarifDelimiter := randomDelimiter("COVERLINT_SARIF")
		if _, err = fmt.Fprintf(w, "sarif<<%s\n%s\n%s\n", sarifDelimiter, sarifJSON, sarifDelimiter); err != nil {
			return fmt.Errorf("writing sarif output: %w", err)
		}
	}

	return nil
}

func fmtPct(p *float64) string {
	if p == nil {
		return "N/A"
	}
	return fmt.Sprintf("%.1f%%", *p)
}
