## 1. Version Infrastructure

- [x] 1.1 Add `var Version = "dev"` to `cmd/pp/main.go`
- [x] 1.2 Set `rootCmd.Version = Version` on the Cobra root command in `cmd/pp/main.go`
- [x] 1.3 Verify `pp --version` and `pp -v` print the version string locally

## 2. GitHub Actions Release Workflow

- [x] 2.1 Create `.github/workflows/release.yml` with trigger `on: push: tags: ['v*']`
- [x] 2.2 Add `permissions: contents: write` to the workflow
- [x] 2.3 Add `go test ./...` step that runs before the build matrix
- [x] 2.4 Add build matrix step for all four targets (linux/amd64, linux/arm64, darwin/arm64, windows/amd64) with `-ldflags "-X main.Version=${{ github.ref_name }}"`
- [x] 2.5 Add step to create GitHub Release with `gh release create` using `--generate-notes` and upload all four binaries

## 3. Create a README & License

- [x] Create an MIT license
- [x] Create a README file with a description, sample commands and installation instructions
