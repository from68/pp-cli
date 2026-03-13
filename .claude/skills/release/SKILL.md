---
name: release
description: Release a new version of the CLI
---

Follow this flow:

1. Check that there are no open changes on the main branch.
2. Create a new tag with the new version number (e.g. `git tag v1.0.0`).
3. Push the tag to GitHub (e.g. `git push origin v1.0.0`).
4. GitHub Actions will automatically create a new release and upload the binaries for the new version.
