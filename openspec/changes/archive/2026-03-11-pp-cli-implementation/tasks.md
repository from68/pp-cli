## 1. Project Scaffolding

- [x] 1.1 Initialize Go module (`go mod init github.com/yourname/pp-cli`)
- [x] 1.2 Add dependencies: cobra, tablewriter, testify (`go get`)
- [x] 1.3 Create directory structure: `cmd/pp/`, `internal/model/`, `internal/xml/`, `internal/format/`, `commands/`
- [x] 1.4 Create `cmd/pp/main.go` entry point that calls the root Cobra command

## 2. Model Layer

- [x] 2.1 Implement `internal/model/money.go` — `Money` struct with `int64` Amount and `Value() float64` helper
- [x] 2.2 Implement `internal/model/shares.go` — `Shares` type (`int64`) with `Value() float64` helper
- [x] 2.3 Implement `internal/model/security.go` — `Security`, `SecurityPrice`, `SecurityEvent` structs
- [x] 2.4 Implement `internal/model/account.go` — `Account`, `AccountTransaction`, `TxUnit` structs
- [x] 2.5 Implement `internal/model/portfolio.go` — `Portfolio`, `PortfolioTransaction` structs
- [x] 2.6 Implement `internal/model/client.go` — `Client` root struct with all nested collections
- [x] 2.7 Implement `internal/model/taxonomy.go` — `Taxonomy`, `Classification`, `Assignment` structs
- [x] 2.8 Implement `internal/model/settings.go` — `Settings`, `Bookmark`, `AttributeType` structs
- [x] 2.9 Write unit tests for `Money.Value()` and `Shares.Value()` formatting helpers

## 3. XML Loader

- [x] 3.1 Implement `internal/xml/loader.go` — file format detection via magic bytes (plain XML, ZIP stub, AES stub)
- [x] 3.2 Return `ErrUnsupportedFormat` for ZIP and AES formats with descriptive messages
- [x] 3.3 Return clear error when `--file` path does not exist
- [x] 3.4 Write unit tests for format detection with byte-level fixtures

## 4. XML Decoder — Pass 1 (Struct Parsing)

- [x] 4.1 Implement `internal/xml/decoder.go` Pass 1: decode full XML into model structs using `encoding/xml`
- [x] 4.2 Handle `<events>` polymorphism (`<event>` and `<dividendEvent>`) via `xml.RawMessage` + custom `UnmarshalXML`
- [x] 4.3 Handle `<crossEntry>` polymorphism via `class` attribute dispatch
- [x] 4.4 Store raw `reference` attribute strings in `SecurityRef` fields (do not resolve yet)
- [x] 4.5 Write golden-file tests: small hand-crafted XML fixtures → expected struct snapshots

## 5. XML Decoder — Pass 2 (Reference Resolution)

- [x] 5.1 Implement `internal/xml/reference_resolver.go` — build `xpath_string → *object` map from decoded tree
- [x] 5.2 Implement XPath relative path traversal (`../` segments, 1-based `[n]` index)
- [x] 5.3 Implement ID-mode detection (scan first 100 bytes for ` id=`) and `id → *object` map
- [x] 5.4 Walk all `AccountTransaction` and `PortfolioTransaction` records and set `Security` pointers
- [x] 5.5 Log warnings (not errors) for unresolvable references; leave pointer nil
- [x] 5.6 Write reference-resolution tests: fixture with XPath references, verify all pointers non-nil
- [x] 5.7 Write reference-resolution tests: fixture with ID-mode references

## 6. Output Format Layer

- [x] 6.1 Implement `internal/format/table.go` — ASCII table output using tablewriter; CSV and TSV as delimiter variants
- [x] 6.2 Implement `internal/format/json.go` — JSON array output using `encoding/json`
- [x] 6.3 Define `OutputFormat` enum (table, json, csv, tsv) and wire `--output` / `-o` global flag in root command

## 7. Root Command and Global Flags

- [x] 7.1 Implement `commands/root.go` — root Cobra command, `--file` / `-f` (required) and `--output` / `-o` flags
- [x] 7.2 Implement pre-run hook that loads and decodes the file, storing `*Client` in context for subcommands

## 8. `pp info` Command

- [x] 8.1 Implement `commands/info.go` — display version, base currency, security/account/portfolio counts, transaction date range
- [x] 8.2 Show "—" for date range when no transactions exist

## 9. `pp securities` Commands

- [x] 9.1 Implement `commands/securities.go` — `securities list` with Name/ISIN/Ticker/Currency/Price Count/Latest Price/Latest Date columns
- [x] 9.2 Add `--retired` flag to include retired securities in the list
- [x] 9.3 Show "—" for price columns when a security has no prices
- [x] 9.4 Implement `securities show <name-or-uuid>` with full detail and price history table
- [x] 9.5 Return non-zero exit code with clear error when security not found

## 10. `pp accounts` Commands

- [x] 10.1 Implement `commands/accounts.go` — `accounts list` with Name/Currency/Balance columns
- [x] 10.2 Compute balance by summing DEPOSIT/TRANSFER_IN and subtracting REMOVAL/TRANSFER_OUT (integer arithmetic)
- [x] 10.3 Implement `accounts transactions <name>` with Date/Type/Amount/Currency/Shares/Security/Note columns
- [x] 10.4 Add `--from` / `--to` date filters (ISO 8601)
- [x] 10.5 Add `--type` filter (single type)
- [x] 10.6 Return non-zero exit code with clear error when account not found

## 11. `pp portfolios` Commands

- [x] 11.1 Implement `commands/portfolios.go` — `portfolios list` with Name/Reference Account columns
- [x] 11.2 Implement `portfolios transactions <name>` with Date/Type/Shares/Amount/Currency/Security/Note columns
- [x] 11.3 Add `--from` / `--to` date filters
- [x] 11.4 Add `--type` filter (comma-separated types)
- [x] 11.5 Display shares as decimal (divide by 10⁸, e.g., `500.00000000`)
- [x] 11.6 Return non-zero exit code with clear error when portfolio not found

## 12. `pp transactions` Command

- [x] 12.1 Implement `commands/transactions.go` — merge and sort all account + portfolio transactions by date
- [x] 12.2 Output columns: Date/Source/Type/Amount/Currency/Shares/Security/Note
- [x] 12.3 Add `--from` / `--to` date filters
- [x] 12.4 Add `--type` filter (comma-separated)
- [x] 12.5 Add `--security <name-or-isin>` filter

## 13. `pp validate` Command

- [x] 13.1 Implement `commands/validate.go` — check all references resolved (report unresolved)
- [x] 13.2 Check buy/sell cross-entry amount consistency
- [x] 13.3 Check all transaction amounts and share quantities are non-negative
- [x] 13.4 Print summary: "Checks: N, Errors: E, Warnings: W"; exit non-zero if errors > 0

## 14. Integration and Fixture Testing

- [x] 14.1 Create `testdata/` with a minimal hand-crafted Portfolio Performance XML fixture covering all transaction types
- [x] 14.2 Create a fixture with XPath references for end-to-end reference-resolution testing
- [x] 14.3 Add integration test: run each CLI command against fixture file and assert output
- [x] 14.4 Add round-trip test stub (decode only; encode phase deferred to Phase 5)
