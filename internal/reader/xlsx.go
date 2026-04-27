// ===================
// © AngelaMos | 2026
// xlsx.go
// ===================

package reader

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

func readXLSXPath(path, sheetName string) (*Table, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open xlsx %s: %w", path, err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("xlsx %s has no sheets", path)
	}

	chosen := sheetName
	if chosen == "" {
		chosen = sheets[0]
	} else {
		found := false
		for _, s := range sheets {
			if s == chosen {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("sheet %q not found in %s (available: %s)",
				chosen, path, strings.Join(sheets, ", "))
		}
	}

	raw, err := f.GetRows(chosen)
	if err != nil {
		return nil, fmt.Errorf("read sheet %q: %w", chosen, err)
	}

	for len(raw) > 0 && isEmptyRow(raw[len(raw)-1]) {
		raw = raw[:len(raw)-1]
	}
	if len(raw) == 0 {
		return &Table{}, nil
	}

	headers := raw[0]
	headerCount := len(headers)
	rows := make([][]string, 0, len(raw)-1)
	for _, row := range raw[1:] {
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

func isEmptyRow(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}
