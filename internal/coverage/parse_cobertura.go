package coverage

import (
	"encoding/xml"
	"fmt"
)

type coberturaCoverage struct {
	XMLName         xml.Name          `xml:"coverage"`
	LinesCovered    int64             `xml:"lines-covered,attr"`
	LinesValid      int64             `xml:"lines-valid,attr"`
	BranchesCovered int64             `xml:"branches-covered,attr"`
	BranchesValid   int64             `xml:"branches-valid,attr"`
	Packages        []coberturaPackge `xml:"packages>package"`
}

type coberturaPackge struct {
	Classes []coberturaClass `xml:"classes>class"`
}

type coberturaClass struct {
	Filename string            `xml:"filename,attr"`
	Methods  []coberturaMethod `xml:"methods>method"`
	Lines    []coberturaLine   `xml:"lines>line"`
}

type coberturaMethod struct {
	Lines []coberturaLine `xml:"lines>line"`
}

type coberturaLine struct {
	Hits int64 `xml:"hits,attr"`
}

func parseCobertura(data []byte) (*CoverageResult, error) {
	var cov coberturaCoverage
	if err := xml.Unmarshal(data, &cov); err != nil {
		return nil, fmt.Errorf("parsing cobertura XML: %w", err)
	}

	if cov.LinesValid == 0 {
		return nil, fmt.Errorf("cobertura: no line data found (is this a valid cobertura report?)")
	}

	result := &CoverageResult{
		Line: &Metric{Hit: cov.LinesCovered, Total: cov.LinesValid},
	}

	if cov.BranchesValid > 0 {
		result.Branch = &Metric{Hit: cov.BranchesCovered, Total: cov.BranchesValid}
	}

	// Extract function coverage and per-file metrics
	var totalMethods, coveredMethods int64
	fileMetrics := map[string]*FileCoverage{}

	for _, pkg := range cov.Packages {
		for _, cls := range pkg.Classes {
			fc, ok := fileMetrics[cls.Filename]
			if !ok {
				fc = &FileCoverage{
					Path: cls.Filename,
					Line: &Metric{},
				}
				fileMetrics[cls.Filename] = fc
			}

			// Count lines per file from class-level line elements
			for _, line := range cls.Lines {
				fc.Line.Total++
				if line.Hits > 0 {
					fc.Line.Hit++
				}
			}

			// Count methods
			var classMethods, classMethodsCovered int64
			for _, method := range cls.Methods {
				totalMethods++
				classMethods++
				for _, line := range method.Lines {
					if line.Hits > 0 {
						coveredMethods++
						classMethodsCovered++
						break
					}
				}
			}
			if classMethods > 0 {
				if fc.Function == nil {
					fc.Function = &Metric{}
				}
				fc.Function.Total += classMethods
				fc.Function.Hit += classMethodsCovered
			}
		}
	}

	if totalMethods > 0 {
		result.Function = &Metric{Hit: coveredMethods, Total: totalMethods}
	}

	for _, fc := range fileMetrics {
		result.Files = append(result.Files, *fc)
	}

	return result, nil
}
