// ===================
// © AngelaMos | 2026
// reader.go
// ===================

package reader

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type Table struct {
	Headers []string
	Rows    [][]string
}

type Options struct {
	Sheet string
}

func ReadFile(path string, opts Options) (*Table, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv", ".tsv", ".txt", "":
		return readCSVPath(path)
	case ".xlsx", ".xlsm":
		return readXLSXPath(path, opts.Sheet)
	default:
		return nil, fmt.Errorf("unsupported file extension %q (expected .csv or .xlsx)", ext)
	}
}

func ReadStream(r io.Reader) (*Table, error) {
	return readCSVStream(r)
}
