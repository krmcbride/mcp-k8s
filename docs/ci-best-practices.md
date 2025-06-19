# GitHub Actions CI Best Practices

## Workflow Trigger Patterns

**Current Setup (Recommended):**

```yaml
on:
  push:
    branches: [master, main]
  pull_request:
    branches: [master, main]
```

**Benefits:**

- **PR Validation**: Catches issues before merge via required status checks
- **Post-merge Verification**: Ensures main branch health after merge
- **Branch Protection**: Enables "Require status checks to pass before merging"

**Alternative Patterns:**

- **PR-only**: `pull_request` only - saves CI resources but no post-merge validation
- **Push-only**: `push` only - validates main branch but not PRs before merge
- **Scheduled**: Add `schedule: cron: '0 2 * * *'` for nightly builds

## Matrix Build Strategy

**Current Cross-Platform Matrix:**

```yaml
strategy:
  matrix:
    os: [linux, darwin, windows]
    arch: [amd64, arm64]
    exclude:
      - os: windows
        arch: arm64
```

**Rationale:**

- **linux/amd64**: Primary development and CI platform
- **darwin/amd64+arm64**: macOS Intel and Apple Silicon support
- **windows/amd64**: Windows compatibility
- **Exclude windows/arm64**: Uncommon deployment target

## Job Dependencies and Gates

**ci-success Job Pattern:**

- Use `needs: [test, lint, build]` to wait for all required jobs
- Use `if: always()` to run even if some jobs fail
- Check individual job results with `needs.jobname.result`
- Provides single status check for branch protection rules

**Best Practices:**

- **Fail Fast**: Set `fail-fast: false` in matrix to see all platform failures
- **Caching**: Use `cache: true` in `setup-go` for faster dependency downloads
- **Permissions**: Use minimal `contents: read` permissions
- **Timeouts**: Add `timeout-minutes: 10` to prevent stuck jobs

## CI Target Usage

**Integration with Makefile:**

- `make test-ci`: Runs tests with coverage output
- `make format-ci`: Checks formatting without modifying files
- `make lint-ci`: Runs linters (delegates to `make lint`)
- `make build-ci`: Cross-platform build validation

**Environment Variables:**

- `GOOS` and `GOARCH`: Set by matrix for cross-compilation
- `GOFLAGS="-mod=vendor"`: Automatically set via Makefile export

