# Release Process

This document describes how releases are automated for beads-tui.

## How It Works

When you push a version tag (e.g., `v1.2.3`), GitHub Actions automatically:

1. Builds binaries for 4 platforms:
   - macOS Intel (`darwin-amd64`)
   - macOS Apple Silicon (`darwin-arm64`)
   - Linux x86_64 (`linux-amd64`)
   - Linux ARM64 (`linux-arm64`)

2. Creates a GitHub Release with all binaries attached

3. Auto-generates release notes from commits since the last tag

## Creating a Release

```bash
# Tag the current commit
git tag -a v1.2.3 -m "Release v1.2.3"

# Push the tag to trigger the workflow
git push origin v1.2.3
```

The release will appear at: https://github.com/andynu/beads-tui/releases

## Authentication Setup

The release workflow requires a Personal Access Token (PAT) because `GITHUB_TOKEN` has a limitation where it cannot create releases on non-HEAD commits ([details](https://github.com/orgs/community/discussions/121022)).

### Token Configuration

- **Secret Name:** `RELEASE_TOKEN`
- **Location:** https://github.com/andynu/beads-tui/settings/secrets/actions
- **Required Scope:** `repo` (full control of private repositories)
- **Current Expiration:** ~90 days from 2025-11-25 (expires around 2025-02-23)

### Renewing the Token

When the token expires, release workflows will fail. To renew:

1. **Generate a new PAT:**
   - Go to https://github.com/settings/tokens
   - Click "Generate new token" → "Generate new token (classic)"
   - Name: `beads-tui releases` (or similar)
   - Expiration: Choose desired duration
   - Scopes: Select `repo`
   - Click "Generate token" and copy it immediately

2. **Update the repository secret:**
   - Go to https://github.com/andynu/beads-tui/settings/secrets/actions
   - Click on `RELEASE_TOKEN`
   - Click "Update secret"
   - Paste the new token
   - Click "Update secret"

3. **Test the workflow:**
   ```bash
   git tag -a v0.0.0-test -m "Test release workflow"
   git push origin v0.0.0-test
   # Verify the release is created, then delete it
   gh release delete v0.0.0-test --yes
   git tag -d v0.0.0-test
   git push origin :refs/tags/v0.0.0-test
   ```

## Workflow Files

- **CI:** `.github/workflows/ci.yml` - Runs tests and linting on push/PR
- **Release:** `.github/workflows/release.yml` - Builds and publishes releases on tag push

## Troubleshooting

### "author_id does not have push access" error

The PAT has expired or is missing. Follow the renewal steps above.

### Release created but no binaries

Check the workflow logs at https://github.com/andynu/beads-tui/actions for build failures.

### Tests fail in CI

Integration tests that require `bd` CLI are skipped in CI (using `-short` flag). If other tests fail, fix them before releasing.
