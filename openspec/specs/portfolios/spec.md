## ADDED Requirements

### Requirement: List portfolios
The `pp portfolios list` command SHALL display all portfolios with columns: Name, Reference Account.

#### Scenario: Portfolio list output
- **WHEN** `pp portfolios list` is run
- **THEN** each portfolio is shown with its name and linked reference account name

### Requirement: View portfolio transactions
The `pp portfolios transactions <portfolio-name>` command SHALL display portfolio transactions for the named portfolio.

#### Scenario: All transactions shown by default
- **WHEN** `pp portfolios transactions "My Depot"` is run
- **THEN** all transactions are shown sorted by date ascending

#### Scenario: Filter by type
- **WHEN** `--type BUY,SELL` is provided
- **THEN** only BUY and SELL transactions are shown

#### Scenario: Filter by date range
- **WHEN** `--from` and `--to` flags are provided
- **THEN** only transactions within the inclusive date range are shown

#### Scenario: Unknown portfolio name
- **WHEN** the provided portfolio name does not match any portfolio
- **THEN** the command exits with a non-zero code and a clear error message

### Requirement: Portfolio transaction output columns
Portfolio transaction output SHALL include columns: Date, Type, Shares, Amount, Currency, Security, Note.

#### Scenario: Shares displayed as decimal
- **WHEN** a portfolio transaction has shares value `50000000000`
- **THEN** the Shares column displays `500.00000000` (divided by 10⁸)
