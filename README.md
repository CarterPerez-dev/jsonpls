```regex
     ██╗███████╗ ██████╗ ███╗   ██╗██████╗ ██╗     ███████╗
     ██║██╔════╝██╔═══██╗████╗  ██║██╔══██╗██║     ██╔════╝
     ██║███████╗██║   ██║██╔██╗ ██║██████╔╝██║     ███████╗
██   ██║╚════██║██║   ██║██║╚██╗██║██╔═══╝ ██║     ╚════██║
╚█████╔╝███████║╚██████╔╝██║ ╚████║██║     ███████╗███████║
 ╚════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝     ╚══════╝╚══════╝
```
### *[Go Docs](https://pkg.go.dev/github.com/CarterPerez-dev/jsonpls)*

[![Personal Tool](https://img.shields.io/badge/Personal-Tool%20%23420-red?style=flat&logo=github)](https://github.com/CarterPerez-dev/jsonpls)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![License: AGPLv3](https://img.shields.io/badge/License-AGPL_v3-purple.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/CarterPerez-dev/jsonpls)](https://goreportcard.com/report/github.com/CarterPerez-dev/jsonpls)

> CSV/XLSX → clean, AI-friendly JSON. One Go binary. Install once, use anywhere.

CSV is fine for spreadsheets but it's the wrong shape for everything else in 2026 — LLMs, APIs, `jq` pipelines, anything that wants typed data. `jsonpls` does the obvious conversion with the smart defaults you'd want by hand: real types, normalized keys, ID-safe big numbers, datetimes as ISO strings.

## Highlights

- **One command, any spreadsheet** — point it at a CSV or XLSX, get JSON back.
- **Smart type inference** per column: integers, floats, booleans, ISO datetimes, nulls. Use `--raw` to opt out.
- **ID-precision safe** — id-hinted columns and any integer larger than JavaScript's `MAX_SAFE_INTEGER` (`2^53-1`) stay as strings, so 17-digit Stripe / Instagram / Meta IDs round-trip without silent precision loss.
- **Snake-cased keys by default** — `"Publish time"` becomes `publish_time`. Use `--keep-headers` to keep originals.
- **Pretty array, JSONL, or compact** — same data, three shapes. JSONL is great for streaming into LLMs and analytics pipelines.
- **Stdin / stdout via `-`** — pipes anywhere a CSV would go.
- **BOM-aware**, multi-line quoted CSV fields, XLSX sheet picking, batch globs.
- **Single static binary**, zero runtime dependencies.

## Installation

### Quickest (curl)

```bash
curl -fsSL https://raw.githubusercontent.com/CarterPerez-dev/jsonpls/main/install.sh | bash
```

The installer detects your OS / architecture, downloads the matching pre-built binary from the latest GitHub release (when available), falls back to `go install` if no binary exists for your platform, and adds `~/.jsonpls/bin` to your `PATH` for `bash`, `zsh`, or `fish`.

Override with environment variables:

```bash
JSONPLS_INSTALL_DIR=$HOME/.local/bin \
JSONPLS_VERSION=v0.1.0 \
  bash <(curl -fsSL https://raw.githubusercontent.com/CarterPerez-dev/jsonpls/main/install.sh)
```

### `go install`

```bash
go install github.com/CarterPerez-dev/jsonpls@latest
```

### Build from source

```bash
git clone https://github.com/CarterPerez-dev/jsonpls.git
cd jsonpls
go build -o jsonpls .
```

## Quick start

```bash
# Convert one file (writes report.json next to report.csv)
jsonpls report.csv

# Batch convert into a folder
jsonpls *.csv -o ~/data/

# XLSX with a specific sheet
jsonpls book.xlsx --sheet "Q1"

# JSONL for streaming into LLMs / analytics
jsonpls data.csv --jsonl

# Disable type inference (everything stays string)
jsonpls data.csv --raw

# Pipe through stdin / stdout
cat data.csv | jsonpls - > data.json
```

## Flags

| Flag | Description |
|------|-------------|
| `-o`, `--out DIR` | Output directory (default: same dir as input) |
| `--jsonl` | One JSON object per line, no enclosing array |
| `--raw` | Disable type inference; everything stays string |
| `--compact` | Minified JSON (default is pretty, 2-space indent) |
| `--keep-headers` | Keep original column headers verbatim (default: snake_case) |
| `--sheet NAME` | XLSX sheet name (default: first sheet) |
| `--stdout` | Write to stdout instead of a file |
| `-h`, `--help` | Show help |
| `-v`, `--version` | Show version |

Flags work in any position — `jsonpls file.csv --jsonl` is equivalent to `jsonpls --jsonl file.csv`.

## Example

Input row from an Instagram analytics export:

```csv
"Post ID","Account ID",Views,"Publish time","Data comment"
17960906015933057,17841472780952814,108270,"04/24/2026 06:05",
```

Output:

```json
[
  {
    "post_id": "17960906015933057",
    "account_id": "17841472780952814",
    "views": 108270,
    "publish_time": "2026-04-24T06:05:00Z",
    "data_comment": null
  }
]
```

Notice `post_id` and `account_id` are **strings** — those 17-digit IDs would round-trip incorrectly if cast to numbers in any JS-based consumer (and most things eventually touch JS).

## How type inference works

Per column, `jsonpls` scans every non-empty value and picks the narrowest type that fits all of them:

```
1. null         empty / whitespace cells (always)
2. bool         every value is true / false / yes / no / t / f
3. int64        every value is an integer (commas tolerated as thousands)
                AND no value exceeds 2^53-1 (JS-safe range)
4. id-string    integer column whose header looks like *_id, OR any
                integer that exceeds the JS-safe range
5. float64      every value parses as a float, and at least one has a
                decimal point or exponent
6. datetime     every value matches a known date layout
                (RFC3339, ISO, US MM/DD/YYYY, etc.) -> normalized to RFC3339
7. string       fallback when none of the above hold
```

Mixed columns (`100`, `abc`, `200`) fall through to `string` so you never lose data. Run with `--raw` to skip the whole pipeline and emit strings only.

## Header normalization

By default, headers are normalized to snake_case for clean JSON keys:

| Original | Normalized |
|----------|------------|
| `Post ID` | `post_id` |
| `Publish time` | `publish_time` |
| `Duration (sec)` | `duration_sec` |
| `Account username` | `account_username` |

Rules: trim → lowercase → replace runs of non-alphanumerics with `_` → strip leading/trailing `_` → disambiguate duplicates with `_2`, `_3`, ... Use `--keep-headers` to skip normalization (collisions still get suffixed).

## File support

- `.csv` — UTF-8, BOM-aware, quoted multi-line fields, lazy quoting tolerated.
- `.tsv`, `.txt` — treated as CSV (Go's `encoding/csv` autodetects the delimiter when reasonable; TSV with literal tabs may need conversion first — issue a PR if this matters).
- `.xlsx`, `.xlsm` — first sheet by default, override with `--sheet NAME`. Values only; formulas are read as their cached results.
- `-` — read CSV from stdin.

## Project structure

```
main.go                       Entry point
internal/
  app/app.go                  Flag parsing + orchestration
  reader/
    reader.go                 Format dispatch, Table type
    csv.go                    CSV reader (BOM, multi-line)
    xlsx.go                   XLSX reader (excelize)
  infer/
    infer.go                  Per-column type inference + casting
    infer_test.go             Table-driven tests
  writer/
    writer.go                 JSON / JSONL output, header normalization
install.sh                    Curl-installable bootstrapper
```

Single external dependency: [`github.com/xuri/excelize/v2`](https://github.com/xuri/excelize) for XLSX reading.

## Development

Requires Go 1.24+.

```bash
go test ./...                 # Run unit tests
go vet ./...                  # Static analysis
go build -o jsonpls .         # Build local binary
```

## License

AGPL 3.0
