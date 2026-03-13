## Context

`pp` is a Go CLI with a single entry point at `cmd/pp/main.go`. There is no existing CI, no `.github/` directory, and no version variable. The module is hosted at `github.com/from68/pp-cli`. Releases will be distributed as raw binaries attached to GitHub Releases.

## Goals / Non-Goals

**Goals:**
- Cross-compile `pp` for Linux (amd64, arm64), macOS (arm64), and Windows (amd64) on every `v*` tag push
- Publish binaries as GitHub Release assets with auto-generated release notes
- Inject the git tag as the binary's version string at build time
- Expose version via `pp --version`

**Non-Goals:**
- GoReleaser or any third-party release tooling
- `.tar.gz` / `.zip` archives (raw binaries only, first iteration)
- Homebrew tap or package manager distribution
- PR CI / test gate workflow
- Linux (amd64) Homebrew or apt/rpm packaging

## Decisions

### 1. Raw GitHub Actions over GoReleaser

GoReleaser adds significant capability (archives, checksums, Homebrew tap, changelog) but also config surface area and an external tool dependency. For a first iteration with raw binary distribution, a single GitHub Actions workflow file is sufficient and easier to understand and modify.

Revisit GoReleaser if checksums, archives, or a Homebrew tap are needed later.

### 2. `var Version = "dev"` in `cmd/pp/main.go`

The version variable lives in `main` package (not a separate internal package) because it's injected via `-ldflags "-X main.Version=..."` which requires the full package path. `"dev"` is the fallback for local `go build` without ldflags.

Alternative considered: a `internal/version/version.go` package. Rejected — unnecessary indirection for a single string.

### 3. Cobra `rootCmd.Version` for `--version`

Cobra's root command has a built-in `.Version` field. Setting it automatically enables the `--version` / `-v` flag with no extra code. Output format: `pp version v0.1.0`.

### 4. Build matrix: Linux amd64/arm64, darwin arm64, windows amd64

Covers the primary user targets. Intel Mac (`darwin/amd64`) excluded per user preference. Can be added to the matrix later with a single line change.

### 5. Trigger: `push: tags: ['v*']`

Standard semantic versioning tag trigger. The release job only runs on tags — no workflow runs on regular commits. Tests run as part of the release job as a sanity check before building.

### 6. Binary naming: `pp-<goos>-<goarch>` (`.exe` for windows)

Simple, unambiguous naming. Users can download the correct binary for their platform without any extraction step.

## Risks / Trade-offs

- **No checksums**: Raw binaries without `sha256sums.txt` means users can't verify integrity. → Acceptable for first iteration; add via GoReleaser if needed.
- **No test-only CI**: Tests only run as part of the release workflow, not on regular commits or PRs. → Acceptable given no current PR process; add a separate `ci.yml` later.
- **`GITHUB_TOKEN` permissions**: The release workflow needs `contents: write` to create releases. This is granted via `permissions:` in the workflow, not a PAT. → Standard GitHub Actions pattern, no risk.

## Migration Plan

1. Add `Version` var to `cmd/pp/main.go` and wire `rootCmd.Version`
2. Create `.github/workflows/release.yml`
3. Tag `v0.1.0`, push — first release builds automatically
