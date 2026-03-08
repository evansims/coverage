package coverage

import (
	"encoding/xml"
	"fmt"
)

type jacocoReport struct {
	XMLName  xml.Name        `xml:"report"`
	Counters []jacocoCounter `xml:"counter"`
	Packages []jacocoPackage `xml:"package"`
}

type jacocoPackage struct {
	Name        string            `xml:"name,attr"`
	SourceFiles []jacocoSourceFile `xml:"sourcefile"`
}

type jacocoSourceFile struct {
	Name     string          `xml:"name,attr"`
	Counters []jacocoCounter `xml:"counter"`
}

type jacocoCounter struct {
	Type    string `xml:"type,attr"`
	Missed  int64  `xml:"missed,attr"`
	Covered int64  `xml:"covered,attr"`
}

func parseJacoco(data []byte) (*CoverageResult, error) {
	var report jacocoReport
	if err := xml.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parsing jacoco XML: %w", err)
	}

	if len(report.Counters) == 0 {
		return nil, fmt.Errorf("jacoco: no counters found at report level")
	}

	result := &CoverageResult{}

	for _, c := range report.Counters {
		total := c.Missed + c.Covered
		switch c.Type {
		case "LINE":
			result.Line = &Metric{Hit: c.Covered, Total: total}
		case "BRANCH":
			if total > 0 {
				result.Branch = &Metric{Hit: c.Covered, Total: total}
			}
		case "METHOD":
			if total > 0 {
				result.Function = &Metric{Hit: c.Covered, Total: total}
			}
		}
	}

	if result.Line == nil {
		return nil, fmt.Errorf("jacoco: no LINE counter found")
	}

	// Collect per-sourcefile metrics
	for _, pkg := range report.Packages {
		for _, sf := range pkg.SourceFiles {
			fc := FileCoverage{
				Path: pkg.Name + "/" + sf.Name,
			}
			for _, c := range sf.Counters {
				total := c.Missed + c.Covered
				switch c.Type {
				case "LINE":
					fc.Line = &Metric{Hit: c.Covered, Total: total}
				case "BRANCH":
					if total > 0 {
						fc.Branch = &Metric{Hit: c.Covered, Total: total}
					}
				case "METHOD":
					if total > 0 {
						fc.Function = &Metric{Hit: c.Covered, Total: total}
					}
				}
			}
			if fc.Line != nil {
				result.Files = append(result.Files, fc)
			}
		}
	}

	return result, nil
}
