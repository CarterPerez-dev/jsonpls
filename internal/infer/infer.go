// ===================
// © AngelaMos | 2026
// infer.go
// ===================

package infer

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ColumnType int

const (
	TypeString ColumnType = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeDate
	TypeIDString
)

const jsMaxSafeInt = int64(9007199254740991)

var dateLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"01/02/2006 15:04:05",
	"01/02/2006 15:04",
	"01/02/2006",
	"1/2/2006 15:04",
	"1/2/2006",
	"2006/01/02 15:04:05",
	"2006/01/02",
	"02 Jan 2006",
	"Jan 2, 2006",
}

var idHeaderPattern = regexp.MustCompile(`(?i)(^|[_\W])id($|[_\W])`)

func InferColumns(headers []string, rows [][]string) []ColumnType {
	out := make([]ColumnType, len(headers))
	for c, h := range headers {
		col := make([]string, len(rows))
		for r, row := range rows {
			if c < len(row) {
				col[r] = row[c]
			}
		}
		out[c] = inferOne(h, col)
	}
	return out
}

func inferOne(header string, values []string) ColumnType {
	idHinted := idHeaderPattern.MatchString(header)

	allBool, allInt, allFloat, allDate := true, true, true, true
	intsOverflowJS := false
	sawAnyValue := false
	sawDecimal := false

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		sawAnyValue = true

		if allBool && !looksBool(v) {
			allBool = false
		}
		if allInt {
			if i, ok := parseInt(v); ok {
				if i > jsMaxSafeInt || i < -jsMaxSafeInt {
					intsOverflowJS = true
				}
			} else {
				allInt = false
			}
		}
		if allFloat {
			if f, ok := parseFloat(v); ok {
				if strings.ContainsAny(v, ".eE") {
					sawDecimal = true
				}
				_ = f
			} else {
				allFloat = false
			}
		}
		if allDate && !looksDate(v) {
			allDate = false
		}

		if !allBool && !allInt && !allFloat && !allDate {
			break
		}
	}

	if !sawAnyValue {
		return TypeString
	}

	switch {
	case allInt && idHinted:
		return TypeIDString
	case allInt && intsOverflowJS:
		return TypeIDString
	case allInt:
		return TypeInt
	case allBool:
		return TypeBool
	case allFloat && sawDecimal:
		return TypeFloat
	case allDate:
		return TypeDate
	default:
		return TypeString
	}
}

func Cast(value string, t ColumnType) any {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	switch t {
	case TypeInt:
		if i, ok := parseInt(v); ok {
			return i
		}
	case TypeFloat:
		if f, ok := parseFloat(v); ok {
			return f
		}
	case TypeBool:
		if b, ok := parseBool(v); ok {
			return b
		}
	case TypeDate:
		if ts, ok := parseDate(v); ok {
			return ts.UTC().Format(time.RFC3339)
		}
	case TypeIDString:
		return v
	}
	return value
}

func looksBool(v string) bool {
	_, ok := parseBool(v)
	return ok
}

func looksDate(v string) bool {
	_, ok := parseDate(v)
	return ok
}

func parseInt(v string) (int64, bool) {
	v = strings.ReplaceAll(v, ",", "")
	i, err := strconv.ParseInt(v, 10, 64)
	return i, err == nil
}

func parseFloat(v string) (float64, bool) {
	v = strings.ReplaceAll(v, ",", "")
	f, err := strconv.ParseFloat(v, 64)
	return f, err == nil
}

func parseBool(v string) (bool, bool) {
	switch strings.ToLower(v) {
	case "true", "t", "yes", "y":
		return true, true
	case "false", "f", "no", "n":
		return false, true
	}
	return false, false
}

func parseDate(v string) (time.Time, bool) {
	for _, layout := range dateLayouts {
		if ts, err := time.Parse(layout, v); err == nil {
			return ts, true
		}
	}
	return time.Time{}, false
}
