## Why

Users need a way to inspect their portfolio constellation — which securities they hold, how many shares, and optionally what those positions are worth — without opening the Portfolio Performance GUI. Currently `pp portfolios` only lists portfolios and their transactions; there is no command to aggregate holdings.

## What Changes

- Add `pp portfolios holdings <portfolio-name>` subcommand that computes net share positions per security from BUY/SELL/DELIVERY_IN/DELIVERY_OUT transactions.
- Display per-holding: Security name, ISIN/ticker (if present), number of shares, latest known price and current value (derived from the security's `prices` list in the XML if available), currency, and note.
- Support `--output` flag (table/json/csv/tsv) consistent with other commands.
- Support `--format` flag to select output format.

## Capabilities

### New Capabilities
- `portfolio-holdings`: Compute and display net share positions (holdings) for a named portfolio, with optional current-value column when price data is available in the XML.

### Modified Capabilities
<!-- No existing spec-level requirement changes -->

## Impact

- `commands/portfolios.go` — add `holdings` subcommand.
- `internal/model/` — may need a `Holding` type aggregating security + shares + value.
- `internal/xml/decoder.go` — ensure security `prices` elements are decoded (latest price per security).
- No new dependencies required.
