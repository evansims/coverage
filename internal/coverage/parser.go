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

// rejectXMLEntities checks the first portion of XML data for DOCTYPE or
// ENTITY declarations that could trigger entity expansion (billion laughs)
// attacks. Coverage reports never contain DTDs.
func rejectXMLEntities(data []byte) error {
	// Check a generous prefix — DTDs appear before the root element
	limit := 4096
	if len(data) < limit {
		limit = len(data)
	}
	prefix := data[:limit]
	if bytes.Contains(prefix, []byte("<!DOCTYPE")) || bytes.Contains(prefix, []byte("<!ENTITY")) {
		return fmt.Errorf("XML contains DOCTYPE or ENTITY declarations, which are not allowed in coverage reports")
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
