package main

import (
	"fmt"
	"os"

	"github.com/evansims/coverlint/internal/coverage"
)

func main() {
	if err := coverage.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "::error::%s\n", err)
		os.Exit(1)
	}
}
