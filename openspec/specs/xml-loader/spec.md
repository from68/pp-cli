## ADDED Requirements

### Requirement: Detect plain XML files
The loader SHALL detect plain XML Portfolio Performance files by checking if the first byte is `<`.

#### Scenario: Plain XML file is detected and loaded
- **WHEN** the file begins with `<`
- **THEN** the loader reads the file as-is and passes the byte stream to the XML decoder

### Requirement: Detect ZIP-compressed files
The loader SHALL detect ZIP-compressed files by checking for the magic bytes `PK\x03\x04` at the start.

#### Scenario: ZIP file detection
- **WHEN** the file begins with `PK\x03\x04`
- **THEN** the loader returns an error indicating ZIP support is not yet implemented (Phase 4)

### Requirement: Detect AES-encrypted files
The loader SHALL detect AES-encrypted files by checking for the 9-byte prefix `PORTFOLIO` at the start.

#### Scenario: AES file detection
- **WHEN** the file begins with `PORTFOLIO`
- **THEN** the loader returns an error indicating AES support is not yet implemented (Phase 4)

### Requirement: Return error for unknown format
The loader SHALL return a descriptive error for files that match none of the known magic bytes.

#### Scenario: Unknown file format
- **WHEN** the file does not match any known magic bytes
- **THEN** the loader returns an error describing the unrecognised format

### Requirement: Accept file path via flag
The CLI SHALL require a `--file` / `-f` flag pointing to the portfolio file.

#### Scenario: Missing file flag
- **WHEN** a command is run without `--file`
- **THEN** the CLI prints a usage error and exits with a non-zero code

#### Scenario: File not found
- **WHEN** `--file` points to a non-existent path
- **THEN** the CLI prints a clear error message and exits with a non-zero code
