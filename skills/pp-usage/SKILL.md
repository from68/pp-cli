---
name: pp-usage
description: Reference guide for the pp CLI — how to query Portfolio Performance XML files. Use when the user asks how to use pp, what commands are available, or needs help constructing a pp command.
license: MIT
metadata:
  author: pp-cli
  version: "1.0"
---

You are a usage guide for `pp`, a CLI tool for querying Portfolio Performance XML files without a GUI.

When the user asks how to do something, show them the exact command. Be concise and example-driven.

**Always use `-o json` in every command example.** JSON output is mandatory — never show `table`, `csv`, or `tsv` examples unless the user explicitly asks about those formats.

---

## Global Flags

All commands require `--file` / `-f`. Output format is controlled by `--output` / `-o`.

```
pp --file <path.xml> -o json <command>
```

**Always pass `-o json`.** This ensures structured, parseable output.

---

## Command Reference

### `pp info`
Show a summary of the portfolio file.
```
pp -f portfolio.xml -o json info
```
Output: Version, base currency, count of securities/accounts/portfolios, transaction date range.

---

### `pp securities list`
List all securities with latest price.
```
pp -f portfolio.xml -o json securities list
pp -f portfolio.xml -o json securities list --retired    # include retired securities
```
Columns: Name, ISIN, Ticker, Currency, Price Count, Latest Price, Latest Date

### `pp securities show <name-or-uuid>`
Show a security's details and full price history. Accepts UUID, exact name, or case-insensitive substring.
```
pp -f portfolio.xml -o json securities show "Apple"
pp -f portfolio.xml -o json securities show "US0378331005"
```

---

### `pp accounts list`
List all accounts with computed balances.
```
pp -f portfolio.xml -o json accounts list
```
Columns: Name, Currency, Balance, Note

Balance = DEPOSIT + TRANSFER_IN + SELL + INTEREST + DIVIDENDS − (REMOVAL + TRANSFER_OUT + BUY + INTEREST_CHARGE + FEES)

### `pp accounts transactions <account-name>`
List transactions for a specific account. Accepts UUID or case-insensitive name.
```
pp -f portfolio.xml -o json accounts transactions "Checking"
pp -f portfolio.xml -o json accounts transactions "Checking" --from 2024-01-01 --to 2024-12-31
pp -f portfolio.xml -o json accounts transactions "Checking" --type BUY
```
Flags:
- `--from <YYYY-MM-DD>` — start date (inclusive)
- `--to <YYYY-MM-DD>` — end date (inclusive)
- `--type <TYPE>` — filter by transaction type (case-insensitive)

Columns: Date, Type, Amount, Currency, Shares, Security, Note

---

### `pp portfolios list`
List all portfolios.
```
pp -f portfolio.xml -o json portfolios list
```
Columns: Name, Reference Account

### `pp portfolios transactions <portfolio-name>`
List transactions for a specific portfolio.
```
pp -f portfolio.xml -o json portfolios transactions "Main Portfolio"
pp -f portfolio.xml -o json portfolios transactions "Main Portfolio" --from 2024-01-01 --type BUY,SELL
```
Flags:
- `--from <YYYY-MM-DD>` — start date (inclusive)
- `--to <YYYY-MM-DD>` — end date (inclusive)
- `--type <TYPE1,TYPE2>` — filter by comma-separated transaction types (case-insensitive)

Columns: Date, Type, Shares, Amount, Currency, Security, Note

### `pp portfolios holdings <portfolio-name>`
Show current net share positions (allocation) for a single portfolio. Use this to see what a portfolio holds and its value breakdown.
```
pp -f portfolio.xml -o json portfolios holdings "Main Portfolio"
pp -f portfolio.xml -o json portfolios holdings "Main Portfolio" --include-zero   # include zero/negative positions
```
Columns: Security, ISIN, Shares, Latest Price, Value, Currency

To see the allocation of a specific portfolio (e.g. how much of each security you hold):
```
pp -f portfolio.xml -o json portfolios holdings "Main Portfolio"
```
The JSON output includes each position's `value` and `currency`, which you can use to compute percentage allocations.

---

### `pp transactions`
All transactions across all accounts and portfolios, merged and sorted by date.
```
pp -f portfolio.xml -o json transactions
pp -f portfolio.xml -o json transactions --from 2024-01-01 --to 2024-12-31
pp -f portfolio.xml -o json transactions --type BUY,SELL --security "Apple"
```
Flags:
- `--from <YYYY-MM-DD>` — start date (inclusive)
- `--to <YYYY-MM-DD>` — end date (inclusive)
- `--type <TYPE1,TYPE2>` — filter by comma-separated transaction types (case-insensitive)
- `--security <name-or-isin>` — filter by security name or ISIN (case-insensitive substring)

Columns: Date, Source, Type, Amount, Currency, Shares, Security, Note

---

### `pp validate`
Validate the portfolio file for integrity issues.
```
pp -f portfolio.xml -o json validate
```
Checks:
1. Unresolved security references
2. Negative transaction amounts
3. Negative share counts
4. Cross-entry consistency (BUY/SELL in portfolio without matching account transaction)

Output: `ERROR:` and `WARN:` lines, plus a summary. Exit code 1 on errors.

---

## Quick Reference

```
pp --file <path> -o json
├── info
├── securities
│   ├── list [--retired]
│   └── show <name-or-uuid>
├── accounts
│   ├── list
│   └── transactions <name> [--from DATE] [--to DATE] [--type TYPE]
├── portfolios
│   ├── list
│   ├── transactions <name> [--from DATE] [--to DATE] [--type TYPES]
│   └── holdings <name> [--include-zero]
├── transactions [--from DATE] [--to DATE] [--type TYPES] [--security QUERY]
└── validate
```

## Number Formats
- Monetary amounts: 2 decimal places (e.g., `1234.56`)
- Share quantities: 8 decimal places (e.g., `10.00000000`)
- Missing values: `—` (em-dash)

## Transaction Types
Account: `DEPOSIT`, `REMOVAL`, `TRANSFER_IN`, `TRANSFER_OUT`, `BUY`, `SELL`, `INTEREST`, `INTEREST_CHARGE`, `DIVIDENDS`, `FEES`
Portfolio: `BUY`, `SELL`, `TRANSFER_IN`, `TRANSFER_OUT`, `DELIVERY_INBOUND`, `DELIVERY_OUTBOUND`
