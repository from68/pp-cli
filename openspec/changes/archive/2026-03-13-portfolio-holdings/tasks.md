## 1. Model

- [x] 1.1 Add `Holding` struct to `internal/model/` with fields: `Security *Security`, `NetShares Shares`, `LatestPrice int64`, `Currency string`, `Value int64`

## 2. Holdings Computation

- [x] 2.1 Implement `ComputeHoldings(portfolio *Portfolio) []Holding` function in `internal/model/` that aggregates net shares per security from portfolio transactions (BUY/DELIVERY_IN add; SELL/DELIVERY_OUT subtract; TRANSFER_IN add; TRANSFER_OUT subtract)
- [x] 2.2 Populate `LatestPrice` and `Value` from `Security.Prices` last entry (0 if empty)
- [x] 2.3 Write unit tests for `ComputeHoldings` covering: BUY only, BUY+SELL, DELIVERY_IN, DELIVERY_OUT, zero position excluded, mixed transaction types

## 3. Command

- [x] 3.1 Add `holdingsCmd` Cobra command in `commands/portfolios.go` with `--output` and `--include-zero` flags
- [x] 3.2 Implement portfolio name lookup (case-insensitive, error on no match) reusing existing pattern
- [x] 3.3 Call `ComputeHoldings`, filter out zero positions unless `--include-zero` is set
- [x] 3.4 Wire `holdingsCmd` as subcommand of `portfoliosCmd` in `init()`

## 4. Output Formatting

- [x] 4.1 Implement table output with columns: Security, ISIN, Shares, Latest Price, Value, Currency (show `—` for Value/Latest Price when 0)
- [x] 4.2 Implement JSON output with fields: security, isin, shares, latestPrice, value, currency
- [x] 4.3 Implement CSV and TSV output (reuse existing format helpers)

## 5. Tests & Fixtures

- [x] 5.1 Update `testdata/minimal.xml` if needed to include security prices for test coverage
- [x] 5.2 Add integration test for `pp portfolios holdings` covering: table output, `--output json`, `--include-zero`, unknown portfolio error
