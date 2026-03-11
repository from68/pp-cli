## ADDED Requirements

### Requirement: Unified transaction view
The `pp transactions` command SHALL display a merged, chronologically sorted view of all account transactions and all portfolio transactions across all accounts and portfolios.

#### Scenario: All transactions shown by default
- **WHEN** `pp transactions` is run
- **THEN** transactions from all accounts and portfolios are shown sorted by date ascending

#### Scenario: Transactions from both accounts and portfolios included
- **WHEN** both account transactions and portfolio transactions exist in the file
- **THEN** both types appear in the unified output with a Source column indicating origin

### Requirement: Filter by date range
The `pp transactions` command SHALL support `--from` and `--to` date filters (ISO 8601 format).

#### Scenario: Date range filter applied
- **WHEN** `--from 2024-01-01 --to 2024-12-31` is provided
- **THEN** only transactions on or between those dates are shown

### Requirement: Filter by transaction type
The `pp transactions` command SHALL support `--type` flag accepting one or more comma-separated transaction types.

#### Scenario: Single type filter
- **WHEN** `--type DIVIDENDS` is provided
- **THEN** only DIVIDENDS transactions are shown

#### Scenario: Multiple type filter
- **WHEN** `--type BUY,SELL` is provided
- **THEN** only BUY and SELL transactions are shown

### Requirement: Filter by security
The `pp transactions` command SHALL support `--security <name-or-isin>` to show only transactions for a specific security.

#### Scenario: Filter by security name
- **WHEN** `--security "Vanguard FTSE All-World"` is provided
- **THEN** only transactions linked to that security are shown

#### Scenario: Filter by ISIN
- **WHEN** `--security IE00B3RBWM25` is provided
- **THEN** only transactions for the security with that ISIN are shown

### Requirement: Unified transaction output columns
Output SHALL include columns: Date, Source, Type, Amount, Currency, Shares, Security, Note.

#### Scenario: Source column identifies origin
- **WHEN** a transaction originates from an account named "Giro"
- **THEN** the Source column shows "Giro"
