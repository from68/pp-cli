## ADDED Requirements

### Requirement: Version variable injected at build time
The binary SHALL expose a `Version` variable in `package main` defaulting to `"dev"`, overridable at build time via `-ldflags "-X main.Version=<value>"`.

#### Scenario: Local build without ldflags
- **WHEN** the binary is built with `go build` without `-ldflags`
- **THEN** `pp --version` SHALL output a string containing `"dev"`

#### Scenario: Release build with ldflags
- **WHEN** the binary is built with `-ldflags "-X main.Version=v0.1.0"`
- **THEN** `pp --version` SHALL output a string containing `"v0.1.0"`

### Requirement: Version exposed via --version flag
The Cobra root command SHALL have its `.Version` field set to the `Version` variable, enabling the built-in `--version` and `-v` flags.

#### Scenario: --version flag
- **WHEN** user runs `pp --version`
- **THEN** the command SHALL print the version string and exit with code 0

#### Scenario: -v shorthand
- **WHEN** user runs `pp -v`
- **THEN** the command SHALL print the version string and exit with code 0
