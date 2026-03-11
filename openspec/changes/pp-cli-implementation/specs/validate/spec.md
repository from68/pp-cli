## ADDED Requirements

### Requirement: Validate all references resolve
The `pp validate` command SHALL check that every `reference` attribute in the XML resolves to an object in the model. It SHALL report each unresolved reference.

#### Scenario: All references resolve
- **WHEN** the file has no dangling references
- **THEN** the command prints "OK" and exits with code 0

#### Scenario: Dangling reference detected
- **WHEN** one or more `reference` attributes cannot be resolved
- **THEN** the command prints each unresolved reference path and exits with a non-zero code

### Requirement: Validate buy/sell cross-entry consistency
The `pp validate` command SHALL check that each `<buysell>` cross-entry has matching amounts on both the account and portfolio transaction sides.

#### Scenario: Consistent buy/sell entry
- **WHEN** both sides of a buysell entry have matching gross amounts
- **THEN** no validation error is reported for that entry

#### Scenario: Inconsistent buy/sell entry
- **WHEN** the account and portfolio sides of a buysell entry have mismatched amounts
- **THEN** the command reports the inconsistency with transaction details

### Requirement: Validate non-negative amounts
The `pp validate` command SHALL verify that all transaction amounts and share quantities are non-negative.

#### Scenario: Negative amount detected
- **WHEN** a transaction has a negative amount
- **THEN** the command reports the offending transaction

### Requirement: Validate summary output
After validation, the command SHALL print a summary: total checks run, errors found, warnings.

#### Scenario: Validation summary
- **WHEN** validation completes
- **THEN** the output shows "Checks: N, Errors: E, Warnings: W"
