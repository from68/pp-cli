## ADDED Requirements

### Requirement: List accounts
The `pp accounts list` command SHALL display all accounts with columns: Name, Currency, Balance (computed from transactions).

#### Scenario: Balance computed from transactions
- **WHEN** `pp accounts list` is run
- **THEN** each account's balance is the sum of DEPOSIT and TRANSFER_IN minus REMOVAL and TRANSFER_OUT amounts in minor units, formatted as a decimal

#### Scenario: Retired accounts excluded by default
- **WHEN** `pp accounts list` is run without `--no-retired` flag
- **THEN** retired accounts are included (retired flag only excludes when `--no-retired` is set globally)

### Requirement: View account transactions
The `pp accounts transactions <account-name>` command SHALL display transactions for the named account.

#### Scenario: All transactions shown by default
- **WHEN** `pp accounts transactions "My Account"` is run
- **THEN** all transactions for that account are shown sorted by date ascending

#### Scenario: Filter by date range
- **WHEN** `--from 2024-01-01 --to 2024-12-31` flags are provided
- **THEN** only transactions within the inclusive date range are shown

#### Scenario: Filter by type
- **WHEN** `--type DIVIDENDS` is provided
- **THEN** only transactions of that type are shown

#### Scenario: Unknown account name
- **WHEN** the provided account name does not match any account
- **THEN** the command exits with a non-zero code and a clear error message

### Requirement: Transaction output columns
Account transaction output SHALL include columns: Date, Type, Amount, Currency, Shares, Security, Note.

#### Scenario: Transaction with security reference
- **WHEN** a transaction has a resolved security reference
- **THEN** the Security column shows the security name

#### Scenario: Transaction without security reference
- **WHEN** a transaction has no security reference
- **THEN** the Security column shows "—"
