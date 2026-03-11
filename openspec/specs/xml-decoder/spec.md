## ADDED Requirements

### Requirement: Decode Client root element
The decoder SHALL parse the `<client>` root element into a `Client` struct containing version, base currency, securities, accounts, portfolios, taxonomies, and settings.

#### Scenario: Valid client XML decoded
- **WHEN** a valid Portfolio Performance XML file is decoded
- **THEN** the resulting `Client` struct has non-zero version and a non-empty base currency

### Requirement: Decode Securities
The decoder SHALL parse each `<security>` element into a `Security` struct with UUID, name, currency, ISIN, ticker, feed, price history, retired flag, and updatedAt timestamp.

#### Scenario: Security with price history decoded
- **WHEN** a security element contains a `<prices>` block with multiple `<price t="…" v="…">` entries
- **THEN** `Security.Prices` contains one `SecurityPrice` per entry with the correct date and integer value

### Requirement: Decode Accounts and Account Transactions
The decoder SHALL parse each `<account>` element with its nested `<account-transaction>` elements.

#### Scenario: Account transaction decoded
- **WHEN** an account contains transactions
- **THEN** each `AccountTransaction` has the correct type, amount (in minor units), date, and raw security reference string

### Requirement: Decode Portfolios and Portfolio Transactions
The decoder SHALL parse each `<portfolio>` element with its nested `<portfolio-transaction>` elements.

#### Scenario: Portfolio transaction decoded
- **WHEN** a portfolio contains transactions
- **THEN** each `PortfolioTransaction` has the correct type, shares (×10⁸ integer), amount, and raw security reference string

### Requirement: Resolve XPath relative references
The decoder SHALL perform a two-pass decode: first building an xpath-position map, then resolving all `reference` attributes to typed pointers.

#### Scenario: XPath reference resolved
- **WHEN** a transaction contains `<security reference="../../../../../securities/security[3]"/>`
- **THEN** `Transaction.Security` points to the third security in the securities list (1-based index)

#### Scenario: Unresolvable reference logged as warning
- **WHEN** a `reference` attribute cannot be resolved in the position map
- **THEN** the decoder logs a warning and leaves the pointer nil rather than returning a fatal error

### Requirement: Detect and handle ID-reference mode
The decoder SHALL auto-detect ID-reference mode by checking for ` id=` within the first 100 bytes of the XML.

#### Scenario: ID-mode reference resolved
- **WHEN** the file uses `id` attributes and `reference="…"` values that are UUIDs/IDs
- **THEN** the decoder builds an `id → *object` map and resolves all references correctly

### Requirement: Handle polymorphic event elements
The decoder SHALL handle `<events>` blocks that contain both `<event>` and `<dividendEvent>` child elements.

#### Scenario: Mixed event types decoded
- **WHEN** a security's `<events>` contains both `<event>` and `<dividendEvent>` children
- **THEN** both types are captured without error

### Requirement: Money amounts stored as integer minor units
The decoder SHALL store all monetary amounts as `int64` minor units (e.g., `100` = €1.00). No float64 conversion SHALL occur during decoding.

#### Scenario: Amount stored as integer
- **WHEN** XML contains `<amount>12345</amount>`
- **THEN** `AccountTransaction.Amount` equals `12345` as `int64`

### Requirement: Shares stored as integer ×10⁸
The decoder SHALL store share quantities as `int64` multiples of 10⁸ (e.g., `50000000000` = 500 shares).

#### Scenario: Shares stored as integer
- **WHEN** XML contains `<shares>50000000000</shares>`
- **THEN** `PortfolioTransaction.Shares` equals `50000000000` as `int64`
