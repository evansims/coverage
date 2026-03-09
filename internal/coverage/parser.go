package coverage

import (
	"bytes"
	"fmt"
)

type parserFunc func(data []byte) (*CoverageResult, error)

var parsers = map[string]parserFunc{
	"lcov":      parseLcov,
	"gocover":   parseGocover,
	"cobertura": parseCobertura,
	"clover":    parseClover,
	"jacoco":    parseJacoco,
}

// rejectXMLEntities checks for ENTITY declarations that could trigger
// entity expansion (billion laughs) attacks. DOCTYPE declarations without
// entities are allowed since some tools (e.g., JaCoCo) include them in
// standard output.
//
// Scans the full file content. The 50MB file size limit makes this negligible
// in performance terms, and scanning the full content prevents bypasses via
// large padding before the DOCTYPE.
func rejectXMLEntities(data []byte) error {
	if bytes.Contains(data, []byte("<!ENTITY")) {
		return fmt.Errorf("XML contains ENTITY declarations, which are not allowed in coverage reports")
	}
	return nil
}

func getParser(format string) (parserFunc, error) {
	p, ok := parsers[format]
	if !ok {
		return nil, fmt.Errorf("unknown coverage format: %q", format)
	}
	return p, nil
}
