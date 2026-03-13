# pp-cli

A Go CLI for querying [Portfolio Performance](https://www.portfolio-performance.info/) XML files from the terminal — no GUI required.

## Installation

Download the latest binary for your platform from [GitHub Releases](https://github.com/from68/pp-cli/releases):

| Platform | Binary |
|---|---|
| Linux (amd64) | `pp-linux-amd64` |
| Linux (arm64) | `pp-linux-arm64` |
| macOS (Apple Silicon) | `pp-darwin-arm64` |
| Windows (amd64) | `pp-windows-amd64.exe` |

```bash
# Linux / macOS example
chmod +x pp-linux-amd64
mv pp-linux-amd64 /usr/local/bin/pp
```

### Build from source

Requires Go 1.21+.

```bash
go install github.com/from68/pp-cli/cmd/pp@latest
```

## Usage

All commands require a Portfolio Performance XML file via `-f`:

```bash
pp -f portfolio.xml <command> [flags]
```

### Commands

```bash
# Show file summary (accounts, securities, portfolios counts)
pp -f portfolio.xml info

# List all securities
pp -f portfolio.xml securities

# List all accounts and balances
pp -f portfolio.xml accounts

# List all portfolios and holdings
pp -f portfolio.xml portfolios

# Show all transactions sorted by date
pp -f portfolio.xml transactions

# Validate file integrity
pp -f portfolio.xml validate

# Check version
pp --version
```

### Output formats

Use `-o` to select output format (`table`, `json`, `csv`, `tsv`):

```bash
pp -f portfolio.xml securities -o json
pp -f portfolio.xml transactions -o csv > transactions.csv
```

## License

[MIT](LICENSE)
