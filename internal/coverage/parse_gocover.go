package coverage

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func parseGocover(data []byte) (*CoverageResult, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var totalStmts, coveredStmts int64
	var hasBlocks bool

	// Per-file tracking
	fileStmts := map[string]int64{}
	fileCovered := map[string]int64{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}

		// Format: file:start.col,end.col stmts count
		lastSpace := strings.LastIndex(line, " ")
		if lastSpace < 0 {
			continue
		}
		countStr := line[lastSpace+1:]
		rest := line[:lastSpace]

		secondLastSpace := strings.LastIndex(rest, " ")
		if secondLastSpace < 0 {
			continue
		}
		stmtsStr := rest[secondLastSpace+1:]
		blockRef := rest[:secondLastSpace]

		stmts, err := strconv.ParseInt(stmtsStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing statement count %q: %w", stmtsStr, err)
		}

		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing execution count %q: %w", countStr, err)
		}

		// Extract file path (everything before the colon+position)
		colonIdx := strings.LastIndex(blockRef, ":")
		filePath := blockRef
		if colonIdx > 0 {
			filePath = blockRef[:colonIdx]
		}

		totalStmts += stmts
		fileStmts[filePath] += stmts
		if count > 0 {
			coveredStmts += stmts
			fileCovered[filePath] += stmts
		}
		hasBlocks = true
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading gocover data: %w", err)
	}

	if !hasBlocks {
		return nil, fmt.Errorf("gocover: no coverage blocks found")
	}

	var files []FileCoverage
	for path, total := range fileStmts {
		files = append(files, FileCoverage{
			Path: path,
			Line: &Metric{Hit: fileCovered[path], Total: total},
		})
	}

	return &CoverageResult{
		Line:  &Metric{Hit: coveredStmts, Total: totalStmts},
		Files: files,
	}, nil
}
