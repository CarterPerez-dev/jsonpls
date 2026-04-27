// ===================
// © AngelaMos | 2026
// csv.go
// ===================

package reader

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func readCSVPath(path string) (*Table, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	return readCSVStream(f)
}

func readCSVStream(r io.Reader) (*Table, error) {
	br := bufio.NewReader(r)
	if first, err := br.Peek(3); err == nil && bytes.Equal(first, utf8BOM) {
		_, _ = br.Discard(3)
	}

	cr := csv.NewReader(br)
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	cr.ReuseRecord = false

	headers, err := cr.Read()
	if err != nil {
		if err == io.EOF {
			return &Table{}, nil
		}
		return nil, fmt.Errorf("read header: %w", err)
	}

	headerCount := len(headers)
	var rows [][]string
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse row %d: %w", len(rows)+2, err)
		}

		if len(row) < headerCount {
			padded := make([]string, headerCount)
			copy(padded, row)
			row = padded
		} else if len(row) > headerCount {
			row = row[:headerCount]
		}
		rows = append(rows, row)
	}

	return &Table{Headers: headers, Rows: rows}, nil
}
