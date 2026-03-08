package coverage

import (
	"encoding/xml"
	"fmt"
)

type coberturaCoverage struct {
	XMLName         xml.Name `xml:"coverage"`
	LinesCovered    int64    `xml:"lines-covered,attr"`
	LinesValid      int64    `xml:"lines-valid,attr"`
	BranchesCovered int64    `xml:"branches-covered,attr"`
	BranchesValid   int64    `xml:"branches-valid,attr"`
}

func parseCobertura(data []byte) (*CoverageResult, error) {
	var cov coberturaCoverage
	if err := xml.Unmarshal(data, &cov); err != nil {
		return nil, fmt.Errorf("parsing cobertura XML: %w", err)
	}

	result := &CoverageResult{
		Line: &Metric{Hit: cov.LinesCovered, Total: cov.LinesValid},
	}

	if cov.BranchesValid > 0 {
		result.Branch = &Metric{Hit: cov.BranchesCovered, Total: cov.BranchesValid}
	}

	return result, nil
}
