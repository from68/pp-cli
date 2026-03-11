## ADDED Requirements

### Requirement: Display portfolio file summary
The `pp info` command SHALL print a summary of the portfolio file including: schema version, base currency, number of securities, number of accounts, number of portfolios, and the date range of transactions.

#### Scenario: Info output for valid file
- **WHEN** `pp info --file portfolio.xml` is run on a valid file
- **THEN** the output shows version, base currency, security count, account count, portfolio count, and the earliest and latest transaction dates

#### Scenario: File with no transactions
- **WHEN** the portfolio file contains no transactions
- **THEN** the date range is shown as "—" or "no transactions"
