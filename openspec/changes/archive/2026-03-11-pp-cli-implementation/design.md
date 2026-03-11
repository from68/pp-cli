## Context

Portfolio Performance stores data in a proprietary XML format using the Java XStream serialization library. The format uses relative XPath references (e.g., `<security reference="../../../../../securities/security[3]"/>`) instead of IDs, making standard XML parsing insufficient. This is a greenfield Go project with no existing code.

## Goals / Non-Goals

**Goals:**
- Parse Portfolio Performance plain-XML files correctly including XPath reference resolution
- Expose data via a composable Cobra CLI with `info`, `securities`, `accounts`, `portfolios`, `transactions`, and `validate` subcommands
- Support table, JSON, CSV, TSV output formats
- Use integer arithmetic throughout (no float64 for money/shares summation)
- Phase 1–3: read-only operations

**Non-Goals:**
- ZIP-compressed or AES-encrypted file support (Phase 4)
- Write-back / mutation commands (Phase 5)
- GUI or web interface
- Real-time price feeds or network requests

## Decisions

### Two-pass XML decoding for reference resolution

**Decision**: Decode in two passes — first parse the full document recording every element's XPath position, then walk all transactions resolving `reference` attributes to typed pointers.

**Rationale**: XStream relative references like `../../../../../securities/security[3]` are positional and require the entire document tree to be available before resolution. A streaming single-pass approach would require buffering unresolved references anyway.

**Alternative considered**: DOM-style parse with `etree` third-party library. Rejected to keep dependencies minimal; stdlib `encoding/xml` with a custom two-pass decoder achieves the same result.

### ID-reference mode auto-detection

**Decision**: Detect reference style by scanning the first 100 bytes for ` id=`. Use an `id → *object` map for the ID-mode and an `xpath_string → *object` map for the default XPath mode.

**Rationale**: Some Portfolio Performance versions use explicit `id` attributes. Both modes must be supported transparently.

### Integer-only arithmetic for Money and Shares

**Decision**: `Money.Amount` is `int64` (minor units). `Shares` is `int64` (×10⁸). All aggregations sum `int64` values; formatting to display strings converts at the last step.

**Rationale**: Float64 accumulation introduces rounding errors unacceptable for financial data.

### Cobra command hierarchy

**Decision**: Root command holds `--file` and `--output` global flags. Subcommands are `info`, `securities list|show`, `accounts list|transactions`, `portfolios list|transactions`, `transactions`, `validate`.

**Rationale**: Mirrors the Portfolio Performance mental model (accounts vs. portfolios) and allows each subcommand to be independently testable.

### Output format abstraction

**Decision**: `internal/format` package provides `table.go` (tablewriter) and `json.go` (stdlib). CSV/TSV are special modes of the table renderer using different delimiters. Commands receive an `OutputFormat` enum and call format helpers.

**Rationale**: Centralises output concerns; commands stay focused on data retrieval.

### Polymorphic XML elements

**Decision**: `<events>` (containing `<event>` and `<dividendEvent>`) and `<crossEntry>` (with `class` attribute) are handled via `xml.RawMessage` capture + custom `UnmarshalXML` with a switch on tag/attribute.

**Rationale**: Go's stdlib `encoding/xml` cannot unmarshal into interface slices automatically; raw capture + manual dispatch is the idiomatic workaround.

## Risks / Trade-offs

- **XPath reference depth** → The relative path traversal (multiple `../`) depends on the exact nesting structure of the XML. Any Portfolio Performance schema version change could break reference resolution. Mitigation: add integration tests with real fixture files; log unresolved references as warnings rather than hard errors.
- **Large price history** → Securities with years of daily prices produce large in-memory slices. Mitigation: price data is only loaded when needed (`securities show`); the list command shows only the latest price.
- **Round-trip fidelity** (Phase 5 risk) → Unknown XML elements not mapped to Go structs will be lost on re-encode. Mitigation: use `xml.RawMessage` passthrough for unrecognised children; defer Phase 5 until the model is stable.
- **XStream index 1-based** → Off-by-one errors in XPath index parsing will silently mis-resolve references. Mitigation: explicit unit tests for the reference resolver with multi-element fixtures.
