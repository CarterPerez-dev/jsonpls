# jsonpls

> CSV/XLSX → clean, AI-friendly JSON. One Go binary. Install once, use anywhere.

CSV is fine for spreadsheets but it's the wrong shape for everything else in 2026 — LLMs, APIs, `jq` pipelines, anything that wants typed data. `jsonpls` does the obvious conversion with the smart defaults you'd want by hand: real types, normalized keys, ID-safe big numbers, datetimes as ISO strings.

## Install

```bash
go install github.com/CarterPerez-dev/jsonpls@latest
```

Requires Go 1.24+.

## Usage

```bash
jsonpls report.csv                       # → report.json next to report.csv
jsonpls *.csv -o ~/data/                 # batch convert into a folder
jsonpls book.xlsx --sheet "Q1"           # pick an XLSX sheet
jsonpls data.csv --jsonl                 # one JSON object per line
cat data.csv | jsonpls - > data.json     # stdin → stdout
```

```
Flags:
  -o, --out DIR        output directory (default: alongside input)
      --jsonl          one JSON object per line, no enclosing array
      --raw            disable type inference; everything stays string
      --compact        minified JSON (default is pretty)
      --keep-headers   keep original column headers (default: snake_case)
      --sheet NAME     XLSX sheet name (default: first sheet)
      --stdout         write to stdout instead of a file
  -h, --help
  -v, --version
```

## What you get by default

- **Real types**: integers stay integers, floats stay floats, `true`/`yes` becomes `true`, empty cells become `null`.
- **ISO datetimes**: common date formats (US `MM/DD/YYYY`, ISO, etc.) are normalized to RFC3339.
- **Snake-cased keys**: `"Publish time"` → `"publish_time"`. Use `--keep-headers` to opt out.
- **ID precision**: columns named like `*_id` and any integer larger than JavaScript's `MAX_SAFE_INTEGER` (`2^53-1`) are kept as strings, so you don't silently lose precision on Stripe/Instagram/whatever IDs.
- **Pretty by default**, `--compact` for one-line.

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

## License

MIT
