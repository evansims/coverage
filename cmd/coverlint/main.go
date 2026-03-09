package main

import (
	"errors"
	"os"

	"github.com/evansims/coverlint/internal/coverage"
)

func main() {
	if err := coverage.Run(); err != nil {
		coverage.EmitAnnotation("error", err.Error())

		var thresholdErr *coverage.ThresholdError
		if errors.As(err, &thresholdErr) {
			os.Exit(1)
		}
		os.Exit(2)
	}
}
