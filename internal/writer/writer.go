// ===================
// © AngelaMos | 2026
// writer.go
// ===================

package writer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/CarterPerez-dev/jsonpls/internal/infer"
	"github.com/CarterPerez-dev/jsonpls/internal/reader"
)

type Options struct {
	JSONL       bool
	Compact     bool
	Raw         bool
	KeepHeaders bool
}

func Write(w io.Writer, table *reader.Table, opts Options) error {
	if table == nil || len(table.Headers) == 0 {
		_, err := io.WriteString(w, emptyOutput(opts))
		return err
	}

	keys := table.Headers
	if !opts.KeepHeaders {
		keys = normalizeHeaders(keys)
	} else {
		keys = disambiguate(keys)
	}

	var types []infer.ColumnType
	if !opts.Raw {
		types = infer.InferColumns(table.Headers, table.Rows)
	}

	bw := bufio.NewWriter(w)
	defer bw.Flush()

	if opts.JSONL {
		return writeJSONL(bw, keys, types, table.Rows, opts.Raw)
	}
	return writeArray(bw, keys, types, table.Rows, opts)
}

func writeArray(w *bufio.Writer, keys []string, types []infer.ColumnType, rows [][]string, opts Options) error {
	if opts.Compact {
		if _, err := w.WriteString("["); err != nil {
			return err
		}
		for i, row := range rows {
			if i > 0 {
				if _, err := w.WriteString(","); err != nil {
					return err
				}
			}
			if err := encodeRow(w, keys, types, row, opts.Raw, false); err != nil {
				return err
			}
		}
		_, err := w.WriteString("]\n")
		return err
	}

	if _, err := w.WriteString("[\n"); err != nil {
		return err
	}
	for i, row := range rows {
		if i > 0 {
			if _, err := w.WriteString(",\n"); err != nil {
				return err
			}
		}
		if _, err := w.WriteString("  "); err != nil {
			return err
		}
		if err := encodeRow(w, keys, types, row, opts.Raw, true); err != nil {
			return err
		}
	}
	_, err := w.WriteString("\n]\n")
	return err
}

func writeJSONL(w *bufio.Writer, keys []string, types []infer.ColumnType, rows [][]string, raw bool) error {
	for _, row := range rows {
		if err := encodeRow(w, keys, types, row, raw, false); err != nil {
			return err
		}
		if _, err := w.WriteString("\n"); err != nil {
			return err
		}
	}
	return nil
}

func encodeRow(w *bufio.Writer, keys []string, types []infer.ColumnType, row []string, raw, indent bool) error {
	obj := orderedRow(keys, types, row, raw)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if indent {
		enc.SetIndent("  ", "  ")
	}

	buf := &strings.Builder{}
	buf.WriteString("{")
	for i, kv := range obj {
		if i > 0 {
			if indent {
				buf.WriteString(",\n    ")
			} else {
				buf.WriteString(",")
			}
		} else if indent {
			buf.WriteString("\n    ")
		}
		k, _ := json.Marshal(kv.key)
		v, err := json.Marshal(kv.val)
		if err != nil {
			return fmt.Errorf("marshal field %q: %w", kv.key, err)
		}
		buf.Write(k)
		if indent {
			buf.WriteString(": ")
		} else {
			buf.WriteString(":")
		}
		buf.Write(v)
	}
	if indent {
		buf.WriteString("\n  }")
	} else {
		buf.WriteString("}")
	}
	_, err := w.WriteString(buf.String())
	return err
}

type kv struct {
	key string
	val any
}

func orderedRow(keys []string, types []infer.ColumnType, row []string, raw bool) []kv {
	out := make([]kv, len(keys))
	for i, k := range keys {
		var raw_v string
		if i < len(row) {
			raw_v = row[i]
		}
		var v any
		if raw {
			if strings.TrimSpace(raw_v) == "" {
				v = nil
			} else {
				v = raw_v
			}
		} else {
			v = infer.Cast(raw_v, types[i])
		}
		out[i] = kv{key: k, val: v}
	}
	return out
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeHeaders(headers []string) []string {
	seen := make(map[string]int, len(headers))
	out := make([]string, len(headers))
	for i, h := range headers {
		s := strings.ToLower(strings.TrimSpace(h))
		s = nonAlnum.ReplaceAllString(s, "_")
		s = strings.Trim(s, "_")
		if s == "" {
			s = fmt.Sprintf("col_%d", i+1)
		}
		out[i] = uniquify(s, seen)
	}
	return out
}

func disambiguate(headers []string) []string {
	seen := make(map[string]int, len(headers))
	out := make([]string, len(headers))
	for i, h := range headers {
		s := strings.TrimSpace(h)
		if s == "" {
			s = fmt.Sprintf("col_%d", i+1)
		}
		out[i] = uniquify(s, seen)
	}
	return out
}

func uniquify(s string, seen map[string]int) string {
	if _, dup := seen[s]; !dup {
		seen[s] = 1
		return s
	}
	seen[s]++
	return fmt.Sprintf("%s_%d", s, seen[s])
}

func emptyOutput(opts Options) string {
	if opts.JSONL {
		return ""
	}
	if opts.Compact {
		return "[]\n"
	}
	return "[]\n"
}
