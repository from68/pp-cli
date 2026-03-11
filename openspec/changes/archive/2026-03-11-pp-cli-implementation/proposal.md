## Why

[Portfolio Performance](https://www.portfolio-performance.info/) is a popular open-source portfolio tracker that stores data in XML files with no official CLI or scripting interface. Power users need a way to query, inspect, and eventually manipulate their portfolio data from the terminal — for automation, reporting, and integration with other tools.

## What Changes

- Introduce a new Go CLI binary (`pp`) that reads Portfolio Performance XML files
- Parse the plain-XML format including XStream relative XPath reference resolution
- Provide `info`, `securities`, `accounts`, `portfolios`, `transactions`, and `validate` subcommands
- Support multiple output formats: table, JSON, CSV, TSV
- Handle file format detection (plain XML, ZIP-compressed, AES-encrypted)
- Implement integer-safe arithmetic for money (minor units) and shares (×10⁸)

## Capabilities

### New Capabilities

- `xml-loader`: Detect and load Portfolio Performance files (plain XML; ZIP and AES as later phases)
- `xml-decoder`: Decode XML into typed Go model structs with XPath reference resolution
- `securities`: List and show securities with price history
- `accounts`: List accounts and view account transactions with filtering
- `portfolios`: List portfolios and view portfolio transactions with filtering
- `transactions`: Unified cross-account/portfolio transaction view with filtering
- `validate`: Validate file integrity — references resolve, cross-entries consistent, amounts non-negative
- `info`: Display file summary (version, currency, counts, date range)

### Modified Capabilities

## Impact

- New Go module (`github.com/yourname/pp-cli`) with no existing code — greenfield
- Runtime dependencies: `github.com/spf13/cobra`, `github.com/olekukonko/tablewriter`, `golang.org/x/crypto/pbkdf2` (Phase 4)
- All file I/O is read-only in Phase 1–3; write-back added in Phase 5
- No external services or network calls required
