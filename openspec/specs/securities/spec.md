## ADDED Requirements

### Requirement: List securities
The `pp securities list` command SHALL display all active securities in a tabular format with columns: Name, ISIN, Ticker, Currency, Price Count, Latest Price, Latest Date.

#### Scenario: Default list excludes retired securities
- **WHEN** `pp securities list` is run
- **THEN** only non-retired securities are shown

#### Scenario: Include retired securities with flag
- **WHEN** `pp securities list --retired` is run
- **THEN** both active and retired securities are shown

#### Scenario: Output format flag
- **WHEN** `pp securities list --output json` is run
- **THEN** the output is a JSON array of security objects

### Requirement: Filter securities with no prices
The list command SHALL show "—" for Latest Price and Latest Date when a security has no price entries.

#### Scenario: Security with no prices
- **WHEN** a security has an empty `<prices>` block
- **THEN** the Latest Price and Latest Date columns show "—"

### Requirement: Show security detail
The `pp securities show <name-or-uuid>` command SHALL display full detail for a single security including all price history.

#### Scenario: Show by name
- **WHEN** `pp securities show "Vanguard FTSE All-World"` is run
- **THEN** the output includes all fields and the full price history table

#### Scenario: Show by UUID
- **WHEN** a valid UUID is supplied to `pp securities show`
- **THEN** the matching security is displayed

#### Scenario: Unknown security
- **WHEN** the name or UUID does not match any security
- **THEN** the command exits with a non-zero code and a clear error message
