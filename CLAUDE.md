# pp-cli

Go CLI (`pp`) for querying Portfolio Performance XML files without a GUI.
Module: `github.com/from68/pp-cli`

## Development Methodology

This project uses [OpenSpec](https://github.com/Fission-AI/OpenSpec/).

Active changes are tracked in `openspec/changes/`. Do not consider `openspec/changes/archive/`.

Get the status via running

```bash
openspec list
```

**The active specification is maintained in `openspec/specs/`.**

## Build & Test

```bash
go build ./...          # build all packages
go build -o /tmp/pp ./cmd/pp  # build binary
go test ./...           # run all tests
```

All `go get`, `go mod tidy`, and `go build`/`go test` commands require `dangerouslyDisableSandbox: true` (network + module cache writes).

## Key Architecture

- `internal/model/` — data types; `Money` = int64 minor units, `Shares` = int64 ×10⁸
- `internal/xml/` — loader (magic-byte detection) + two-pass decoder + XPath reference resolver
- `internal/format/` — table/JSON/CSV/TSV output; uses tablewriter v1.1.3 (new API, not v0.0.5)
- `commands/` — Cobra subcommands wired in `init()`, file loaded in `PersistentPreRunE`

## tablewriter v1.1.3 API

```go
t := tablewriter.NewTable(w)
t.Header([]string{"Col1"})
t.Append([]string{"val"})
t.Render()
```

## PP XML Quirks

- Dates: `"2006-01-02T15:04"`, `"2006-01-02"`, or Unix milliseconds
- XPath security refs like `../../../../../securities/security[3]` — strip all `../` prefixes, then look up in absolute path map
- ID-mode auto-detected by checking first 100 bytes for ` id=`
- Test fixtures: `testdata/minimal.xml`, `testdata/xpath_refs.xml`
