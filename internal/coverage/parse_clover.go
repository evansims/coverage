package coverage

import (
	"encoding/xml"
	"fmt"
)

type cloverCoverage struct {
	XMLName xml.Name      `xml:"coverage"`
	Project cloverProject `xml:"project"`
}

type cloverProject struct {
	Metrics  cloverMetrics  `xml:"metrics"`
	Packages []cloverPkg    `xml:"package"`
	Files    []cloverFile   `xml:"file"`
}

type cloverPkg struct {
	Files []cloverFile `xml:"file"`
}

type cloverFile struct {
	Name    string        `xml:"name,attr"`
	Path    string        `xml:"path,attr"`
	Metrics cloverMetrics `xml:"metrics"`
}

type cloverMetrics struct {
	Statements          int64 `xml:"statements,attr"`
	CoveredStatements   int64 `xml:"coveredstatements,attr"`
	Conditionals        int64 `xml:"conditionals,attr"`
	CoveredConditionals int64 `xml:"coveredconditionals,attr"`
	Methods             int64 `xml:"methods,attr"`
	CoveredMethods      int64 `xml:"coveredmethods,attr"`
}

func parseClover(data []byte) (*CoverageResult, error) {
	var cov cloverCoverage
	if err := xml.Unmarshal(data, &cov); err != nil {
		return nil, fmt.Errorf("parsing clover XML: %w", err)
	}

	m := cov.Project.Metrics

	if m.Statements == 0 {
		return nil, fmt.Errorf("clover: no statement data found (is this a valid clover report?)")
	}

	result := &CoverageResult{
		Line: &Metric{Hit: m.CoveredStatements, Total: m.Statements},
	}

	if m.Conditionals > 0 {
		result.Branch = &Metric{Hit: m.CoveredConditionals, Total: m.Conditionals}
	}

	if m.Methods > 0 {
		result.Function = &Metric{Hit: m.CoveredMethods, Total: m.Methods}
	}

	// Collect per-file metrics from packages and top-level files
	var allFiles []cloverFile
	for _, pkg := range cov.Project.Packages {
		allFiles = append(allFiles, pkg.Files...)
	}
	allFiles = append(allFiles, cov.Project.Files...)

	for _, f := range allFiles {
		name := f.Path
		if name == "" {
			name = f.Name
		}
		if name == "" || f.Metrics.Statements == 0 {
			continue
		}

		fc := FileCoverage{
			Path: name,
			Line: &Metric{Hit: f.Metrics.CoveredStatements, Total: f.Metrics.Statements},
		}
		if f.Metrics.Conditionals > 0 {
			fc.Branch = &Metric{Hit: f.Metrics.CoveredConditionals, Total: f.Metrics.Conditionals}
		}
		if f.Metrics.Methods > 0 {
			fc.Function = &Metric{Hit: f.Metrics.CoveredMethods, Total: f.Metrics.Methods}
		}
		result.Files = append(result.Files, fc)
	}

	return result, nil
}
