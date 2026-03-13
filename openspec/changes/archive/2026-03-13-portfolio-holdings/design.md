## Context

The `pp` CLI already decodes Portfolio Performance XML into `model.Portfolio` (with `[]PortfolioTransaction`) and `model.Security` (with `[]SecurityPrice`). Holdings must be computed by aggregating transaction shares per security — there is no pre-computed "current holding" element in the PP XML format.

Existing pattern: subcommands live in `commands/portfolios.go` and follow the Cobra pattern of `portfolios <subcmd> <name>`.

## Goals / Non-Goals

**Goals:**
- Compute net share positions per security from portfolio transactions.
- Show latest known price and current value when `Security.Prices` is non-empty.
- Respect `--output` flag (table/json/csv/tsv) like all other commands.
- Filter to a named portfolio (case-insensitive partial match, consistent with other subcommands).

**Non-Goals:**
- Real-time price fetching (no network calls).
- Cost-basis / P&L calculation.
- Multi-portfolio aggregation.

## Decisions

### Holdings computation logic
Aggregate `PortfolioTransaction.Shares` by `Security.UUID`:
- BUY, DELIVERY_IN → add shares
- SELL, DELIVERY_OUT → subtract shares
- TRANSFER_IN → add; TRANSFER_OUT → subtract

Positions with net shares ≤ 0 are excluded from output by default (sold-out positions). A `--include-zero` flag can show them.

**Why this approach:** The PP XML has no snapshot balances at portfolio level; transaction replay is the canonical approach used by the PP app itself.

### Current value
Take the last entry in `Security.Prices` (already ordered by date in the XML) as the latest price. Value = `latestPrice × netShares / 10^8`. If `Prices` is empty, the Value column shows `—`.

**Why last price:** Prices are appended chronologically in the PP XML; no sorting needed.

### Note field
The holdings note is not per-security-per-portfolio in PP XML; we will omit it or use the security's own note if the model exposes it. For now the Note column is not included (PP XML doesn't store a per-holding note — notes are per transaction).

### Model addition
Add `model.Holding` struct (computed, not decoded from XML):
```go
type Holding struct {
    Security    *Security
    NetShares   Shares
    LatestPrice int64  // 0 if unknown
    Currency    string
    Value       int64  // 0 if unknown
}
```
Kept in `internal/model/` for reuse in future commands.

## Risks / Trade-offs

- **Split transactions / corporate actions**: Stock splits recorded as `DELIVERY_IN/OUT` or `TRANSFER` with special notes will affect net shares — but replaying all transactions is correct regardless of event type.
  → Mitigation: use the complete sign-adjusted logic above; no special-casing needed for correctness.

- **Currency mismatch**: If a portfolio holds securities denominated in multiple currencies, current value is per-security-currency — no FX conversion.
  → Mitigation: display currency per row; no aggregated total row.
