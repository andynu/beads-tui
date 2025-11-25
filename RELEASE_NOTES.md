# Release Process

This document describes how releases are automated for beads-tui.

## How It Works

When you push a version tag (e.g., `v1.2.3`), GitHub Actions runs [GoReleaser](https://goreleaser.com/) which:

1. Builds binaries for 4 platforms:
   - macOS Intel (`darwin_amd64`)
   - macOS Apple Silicon (`darwin_arm64`)
   - Linux x86_64 (`linux_amd64`)
   - Linux ARM64 (`linux_arm64`)

2. Creates `.tar.gz` archives for each platform

3. Generates `checksums.txt` for verification

4. Creates a GitHub Release with changelog from commits

## Creating a Release

```bash
# Tag the current commit
git tag -a v1.2.3 -m "Release v1.2.3"

# Push the tag to trigger the workflow
git push origin v1.2.3
```

The release will appear at: https://github.com/andynu/beads-tui/releases

## Configuration Files

- `.goreleaser.yaml` - GoReleaser build configuration
- `.github/workflows/release.yml` - GitHub Actions workflow

## Authentication (PAT Workaround)

Due to a [GitHub bug affecting new repositories](https://github.com/orgs/community/discussions/180369), the workflow uses a Personal Access Token instead of the default `GITHUB_TOKEN`.

### Token Configuration

- **Secret Name:** `RELEASE_TOKEN`
- **Location:** https://github.com/andynu/beads-tui/settings/secrets/actions
- **Required Scope:** `repo` (full control of private repositories)
- **Current Expiration:** ~90 days from 2025-11-25 (expires around 2025-02-23)

### When GitHub Fixes the Bug

Once GitHub resolves [Discussion #180369](https://github.com/orgs/community/discussions/180369), edit `.github/workflows/release.yml` and change:

```yaml
GITHUB_TOKEN: ${{ secrets.RELEASE_TOKEN }}
```

to:

```yaml
GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

Then the PAT will no longer be needed.

### Renewing the Token (If Still Needed)

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

## Local Testing

You can test GoReleaser locally before pushing:

```bash
# Install goreleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Validate config
goreleaser check

# Dry run (builds but doesn't publish)
goreleaser release --snapshot --clean
```

## Troubleshooting

### "author_id does not have push access" error

This is the GitHub bug. Ensure `RELEASE_TOKEN` secret is set with a valid PAT.

### Release created but missing assets

Check the workflow logs at https://github.com/andynu/beads-tui/actions for build failures.

### GoReleaser config errors

Run `goreleaser check` locally to validate `.goreleaser.yaml`.
