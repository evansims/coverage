package coverage

import "fmt"

type parserFunc func(data []byte) (*CoverageResult, error)

var parsers = map[string]parserFunc{
	"lcov":      parseLcov,
	"gocover":   parseGocover,
	"cobertura": parseCobertura,
	"clover":    parseClover,
	"jacoco":    parseJacoco,
}

func getParser(format string) (parserFunc, error) {
	p, ok := parsers[format]
	if !ok {
		return nil, fmt.Errorf("unknown coverage format: %q", format)
	}
	return p, nil
}
