// ===================
// © AngelaMos | 2026
// app.go
// ===================

package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CarterPerez-dev/jsonpls/internal/reader"
	"github.com/CarterPerez-dev/jsonpls/internal/writer"
)

const (
	exitOK    = 0
	exitUsage = 2
	exitFail  = 1
)

var version = "dev"

const usage = `jsonpls — convert CSV/XLSX to AI-friendly JSON

Usage:
  jsonpls FILE [FILE...] [flags]
  jsonpls -                    read CSV from stdin, write JSON to stdout
  cat foo.csv | jsonpls -      same thing

Flags:
  -o, --out DIR        write output to DIR (default: same dir as input)
      --jsonl          one JSON object per line (no enclosing array)
      --raw            disable type inference; everything stays string
      --compact        minified JSON (default is pretty, 2-space indent)
      --keep-headers   keep original column headers verbatim (default: snake_case them)
      --sheet NAME     XLSX: sheet name to read (default: first sheet)
      --stdout         write to stdout instead of a file
  -h, --help           show this help
  -v, --version        show version

Examples:
  jsonpls report.csv
  jsonpls *.csv -o ~/data/
  jsonpls book.xlsx --sheet "Q1" --jsonl
  cat data.csv | jsonpls - > data.json
`

type flags struct {
	out         string
	jsonl       bool
	raw         bool
	compact     bool
	keepHeaders bool
	sheet       string
	stdout      bool
	help        bool
	version     bool
}

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("jsonpls", flag.ContinueOnError)
	fl.SetOutput(stderr)
	fl.Usage = func() { fmt.Fprint(stderr, usage) }

	var f flags
	fl.StringVar(&f.out, "o", "", "")
	fl.StringVar(&f.out, "out", "", "")
	fl.BoolVar(&f.jsonl, "jsonl", false, "")
	fl.BoolVar(&f.raw, "raw", false, "")
	fl.BoolVar(&f.compact, "compact", false, "")
	fl.BoolVar(&f.keepHeaders, "keep-headers", false, "")
	fl.StringVar(&f.sheet, "sheet", "", "")
	fl.BoolVar(&f.stdout, "stdout", false, "")
	fl.BoolVar(&f.help, "h", false, "")
	fl.BoolVar(&f.help, "help", false, "")
	fl.BoolVar(&f.version, "v", false, "")
	fl.BoolVar(&f.version, "version", false, "")

	if err := fl.Parse(reorderFlagsFirst(args)); err != nil {
		return exitUsage
	}

	if f.help {
		fmt.Fprint(stdout, usage)
		return exitOK
	}
	if f.version {
		fmt.Fprintf(stdout, "jsonpls %s\n", version)
		return exitOK
	}

	inputs := fl.Args()
	if len(inputs) == 0 {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}

	wopts := writer.Options{
		JSONL:       f.jsonl,
		Compact:     f.compact,
		Raw:         f.raw,
		KeepHeaders: f.keepHeaders,
	}
	ropts := reader.Options{Sheet: f.sheet}

	hadFail := false
	for _, in := range inputs {
		if in == "-" {
			if err := convertStream(stdin, stdout, wopts); err != nil {
				fmt.Fprintf(stderr, "jsonpls: stdin: %v\n", err)
				hadFail = true
			}
			continue
		}

		if err := convertFile(in, f, ropts, wopts, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "jsonpls: %s: %v\n", in, err)
			hadFail = true
			continue
		}
	}

	if hadFail {
		return exitFail
	}
	return exitOK
}

func convertStream(in io.Reader, out io.Writer, wopts writer.Options) error {
	t, err := reader.ReadStream(in)
	if err != nil {
		return err
	}
	return writer.Write(out, t, wopts)
}

func convertFile(path string, f flags, ropts reader.Options, wopts writer.Options, stdout, stderr io.Writer) error {
	t, err := reader.ReadFile(path, ropts)
	if err != nil {
		return err
	}

	if f.stdout {
		return writer.Write(stdout, t, wopts)
	}

	outPath := outputPath(path, f.out, f.jsonl)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if _, err := os.Stat(outPath); err == nil {
		fmt.Fprintf(stderr, "jsonpls: overwriting %s\n", outPath)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer out.Close()

	if err := writer.Write(out, t, wopts); err != nil {
		return err
	}
	fmt.Fprintf(stderr, "jsonpls: wrote %s (%d rows)\n", outPath, len(t.Rows))
	return nil
}

func reorderFlagsFirst(args []string) []string {
	valueFlags := map[string]bool{
		"-o": true, "--o": true, "-out": true, "--out": true,
		"-sheet": true, "--sheet": true,
	}

	var flags, positional []string
	i := 0
	for i < len(args) {
		a := args[i]
		if a == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if a == "-" || !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
			i++
			continue
		}
		flags = append(flags, a)
		eq := strings.Contains(a, "=")
		key := a
		if eq {
			key = a[:strings.Index(a, "=")]
		}
		if !eq && valueFlags[key] && i+1 < len(args) {
			flags = append(flags, args[i+1])
			i += 2
			continue
		}
		i++
	}
	return append(flags, positional...)
}

func outputPath(input, outDir string, jsonl bool) string {
	base := filepath.Base(input)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	newExt := ".json"
	if jsonl {
		newExt = ".jsonl"
	}
	name := stem + newExt

	if outDir != "" {
		return filepath.Join(outDir, name)
	}
	return filepath.Join(filepath.Dir(input), name)
}
