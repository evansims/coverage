package coverage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// BaselineData holds previous coverage data for delta comparison.
type BaselineData struct {
	Score    float64  `json:"score"`
	Line     *float64 `json:"line,omitempty"`
	Branch   *float64 `json:"branch,omitempty"`
	Function *float64 `json:"function,omitempty"`

	Timestamp string `json:"timestamp"`
}

// GenerateBaseline creates a BaselineData snapshot from the given results.
// It uses the last entry (the total/combined result).
func GenerateBaseline(results []EntryResult) BaselineData {
	bd := BaselineData{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	if len(results) == 0 {
		return bd
	}

	last := results[len(results)-1]
	if last.Score != nil {
		bd.Score = *last.Score
	}
	bd.Line = last.Line
	bd.Branch = last.Branch
	bd.Function = last.Function

	return bd
}

// LoadBaseline loads baseline coverage data from a local file or HTTPS URL.
func LoadBaseline(source string) (*BaselineData, error) {
	var data []byte
	var err error

	if strings.HasPrefix(source, "https://") {
		data, err = fetchBaselineHTTPS(source)
	} else {
		data, err = readBaselineFile(source)
	}
	if err != nil {
		return nil, err
	}

	var bd BaselineData
	if err := json.Unmarshal(data, &bd); err != nil {
		return nil, fmt.Errorf("parsing baseline JSON: %w", err)
	}

	return &bd, nil
}

func readBaselineFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading baseline file %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("reading baseline file %q: %w", path, err)
	}
	if info.Size() > maxCoverageFileSize {
		return nil, fmt.Errorf("baseline file %q exceeds maximum size of %d bytes (%d bytes)", path, maxCoverageFileSize, info.Size())
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading baseline file %q: %w", path, err)
	}

	return data, nil
}

func fetchBaselineHTTPS(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url) //nolint:gosec // URL is user-provided input validated to be HTTPS
	if err != nil {
		return nil, fmt.Errorf("fetching baseline from %q: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching baseline from %q: HTTP %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxCoverageFileSize))
	if err != nil {
		return nil, fmt.Errorf("reading baseline response from %q: %w", url, err)
	}

	return data, nil
}

// CompareBaseline checks whether the coverage delta meets the minimum allowed change.
// Returns a Violation if the score dropped more than allowed by minDelta.
func CompareBaseline(prev *BaselineData, currentScore float64, minDelta *float64) []Violation {
	if minDelta == nil {
		return nil
	}

	delta := currentScore - prev.Score
	if delta < *minDelta {
		return []Violation{
			{
				Entry:    "coverage",
				Metric:   "delta",
				Actual:   delta,
				Required: *minDelta,
			},
		}
	}

	return nil
}
