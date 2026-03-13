## Why

The `pp` CLI has no automated build or distribution mechanism — binaries must be built manually. Adding a CI-based release pipeline enables users to install versioned binaries directly from GitHub Releases without requiring a local Go toolchain.

## What Changes

- Add `.github/workflows/release.yml` — triggers on `v*` tags, cross-compiles for all target platforms, and publishes a GitHub Release with binary assets and auto-generated release notes
- Add `Version` variable in `cmd/pp/main.go` injected at build time via `ldflags`
- Wire `--version` flag on the Cobra root command using the injected version

## Capabilities

### New Capabilities

- `cli-versioning`: Expose build-time version string via `--version` flag
- `github-release`: Automated cross-platform binary builds published to GitHub Releases on version tag push

### Modified Capabilities

## Impact

- New file: `.github/workflows/release.yml`
- Modified: `cmd/pp/main.go` — add `Version` var and root command version wiring
- No changes to existing commands, internal packages, or test fixtures
- No new Go dependencies
