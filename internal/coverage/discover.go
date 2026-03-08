package coverage

import (
	"fmt"
	"os"
	"path/filepath"
)

// defaultPaths maps coverage formats to common default output paths,
// ordered by likelihood. These are the standard output locations for
// the most popular coverage tools in each ecosystem.
var defaultPaths = map[string][]string{
	"lcov": {
		"coverage/lcov.info",
		"lcov.info",
		"coverage.lcov",
	},
	"gocover": {
		"cover.out",
		"coverage.out",
		"c.out",
	},
	"cobertura": {
		"coverage.xml",
		"cobertura.xml",
		"cobertura-coverage.xml",
	},
	"clover": {
		"coverage.xml",
		"clover.xml",
	},
	"jacoco": {
		"build/reports/jacoco/test/jacocoTestReport.xml",
		"target/site/jacoco/jacoco.xml",
		"jacoco.xml",
	},
}

// DiscoverReport searches for a coverage report file using the default paths
// for the given format. It returns the first path that exists, relative to workDir.
func DiscoverReport(format, workDir string) (string, error) {
	paths, ok := defaultPaths[format]
	if !ok {
		return "", fmt.Errorf("no default paths configured for format %q", format)
	}

	for _, p := range paths {
		full := filepath.Join(workDir, p)
		if _, err := os.Stat(full); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("auto-discovery: no %s coverage report found in %q (searched: %v)", format, workDir, paths)
}
