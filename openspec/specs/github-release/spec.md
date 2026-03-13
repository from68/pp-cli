## ADDED Requirements

### Requirement: Release workflow triggers on version tags
The GitHub Actions release workflow SHALL trigger on pushes to tags matching `v*` and on no other events.

#### Scenario: Version tag push triggers release
- **WHEN** a tag matching `v*` (e.g., `v0.1.0`) is pushed to the repository
- **THEN** the release workflow SHALL start

#### Scenario: Regular commit does not trigger release
- **WHEN** a commit is pushed to any branch without a matching tag
- **THEN** the release workflow SHALL NOT run

### Requirement: Tests pass before binaries are built
The release job SHALL run `go test ./...` before compiling any binaries, and SHALL abort if tests fail.

#### Scenario: Tests fail
- **WHEN** `go test ./...` exits non-zero
- **THEN** the workflow SHALL fail and no binaries SHALL be built or published

### Requirement: Cross-platform binary matrix
The workflow SHALL build binaries for the following targets:

| Platform | GOOS | GOARCH | Asset name |
|---|---|---|---|
| Linux x86-64 | linux | amd64 | `pp-linux-amd64` |
| Linux ARM64 | linux | arm64 | `pp-linux-arm64` |
| macOS Apple Silicon | darwin | arm64 | `pp-darwin-arm64` |
| Windows x86-64 | windows | amd64 | `pp-windows-amd64.exe` |

#### Scenario: All platform binaries produced
- **WHEN** the release workflow completes successfully
- **THEN** all four platform binaries SHALL be attached to the GitHub Release

### Requirement: Version injected into each binary
Each binary SHALL be built with `-ldflags "-X main.Version=<tag>"` where `<tag>` is the triggering git tag (e.g., `v0.1.0`).

#### Scenario: Binary reports correct version
- **WHEN** a released binary is run with `--version`
- **THEN** it SHALL output the tag that triggered its build

### Requirement: GitHub Release created with auto-generated notes
The workflow SHALL create a GitHub Release using `gh release create` (or equivalent) with `--generate-notes` to auto-populate release notes from commit messages.

#### Scenario: Release created on tag push
- **WHEN** the workflow completes
- **THEN** a GitHub Release SHALL exist for the tag with all four binaries attached and auto-generated release notes
