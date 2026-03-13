## ADDED Requirements

### Requirement: View portfolio holdings
The `pp portfolios holdings <portfolio-name>` command SHALL display the current net share positions for all securities held in the named portfolio, computed by replaying its transactions.

#### Scenario: Holdings shown for known portfolio
- **WHEN** `pp portfolios holdings "My Depot"` is run
- **THEN** each security with a positive net share position is listed with columns: Security, ISIN, Shares, Latest Price, Value, Currency

#### Scenario: Zero or negative positions excluded by default
- **WHEN** a security's net shares across all transactions is ≤ 0
- **THEN** that security does NOT appear in the output

#### Scenario: Include zero positions with flag
- **WHEN** `--include-zero` flag is provided
- **THEN** securities with net shares ≤ 0 are also shown

#### Scenario: Unknown portfolio name
- **WHEN** the provided portfolio name does not match any portfolio
- **THEN** the command exits with a non-zero code and a clear error message

### Requirement: Holdings share computation
Net shares per security SHALL be computed by summing signed shares across all portfolio transactions for that security.

#### Scenario: BUY increases shares
- **WHEN** a portfolio has a BUY transaction of 100 shares for security X
- **THEN** security X shows 100 net shares

#### Scenario: SELL decreases shares
- **WHEN** a portfolio has a BUY of 100 shares followed by SELL of 40 shares for security X
- **THEN** security X shows 60 net shares

#### Scenario: DELIVERY_IN increases shares
- **WHEN** a portfolio has a DELIVERY_IN transaction of 50 shares for security X
- **THEN** security X shows 50 net shares

#### Scenario: DELIVERY_OUT decreases shares
- **WHEN** a portfolio has DELIVERY_IN of 50 followed by DELIVERY_OUT of 50 for security X
- **THEN** security X shows 0 net shares (excluded from default output)

### Requirement: Holdings current value
When a security has price history, the Value column SHALL be computed as `latestPrice × netShares / 10^8`.

#### Scenario: Value shown when prices available
- **WHEN** a security has at least one entry in its price history
- **THEN** the Value column shows `latestPrice × netShares / 10^8` formatted as a decimal
- **AND** the Currency column shows the security's currency

#### Scenario: Value omitted when no prices available
- **WHEN** a security has no price history
- **THEN** the Value column displays `—`

### Requirement: Holdings output format
The `holdings` command SHALL support `--output` flag values: `table` (default), `json`, `csv`, `tsv`.

#### Scenario: Default table output
- **WHEN** no `--output` flag is provided
- **THEN** output is formatted as an ASCII table

#### Scenario: JSON output
- **WHEN** `--output json` is provided
- **THEN** output is a JSON array of holding objects with fields: security, isin, shares, latestPrice, value, currency
