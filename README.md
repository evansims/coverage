# Coverlint

![Coverage](https://raw.githubusercontent.com/evansims/coverlint/badges/coverage.svg)

A self-contained GitHub Action that parses coverage reports, enforces thresholds, and reports results as GitHub Actions annotations and job summaries. No external services, secrets, or other headaches — just pass/fail.

## Supported Formats

| Format           | Flag        | Typical Producer                                     |
| ---------------- | ----------- | ---------------------------------------------------- |
| LCOV             | `lcov`      | `cargo llvm-cov`, `c8`, `istanbul`, `jest`, `vitest` |
| Go cover profile | `gocover`   | `go test -coverprofile`                              |
| Cobertura XML    | `cobertura` | `pytest-cov`, `istanbul`, `cargo tarpaulin`          |
| Clover XML       | `clover`    | `phpunit`, some JS tools                             |
| JaCoCo XML       | `jacoco`    | Gradle/Maven JaCoCo plugin                           |

## Usage

```yaml
- uses: evansims/coverlint@v1
  with:
    format: gocover # recommended; auto-detected if omitted
    threshold-line: 80
```

For [security hardening](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions#using-third-party-actions), pin to the commit SHA from each release (e.g. `evansims/coverlint@COMMIT_SHA # v1.0.0`) instead of a mutable tag.

### Inputs

| Input                | Default | Description                                                                                            |
| -------------------- | ------- | ------------------------------------------------------------------------------------------------------ |
| `format`             |         | Coverage format(s), one per line or comma-separated. Auto-detected if omitted                          |
| `path`               |         | Path(s) to coverage files, one per line or comma-separated. Supports globs. Auto-discovered if omitted |
| `threshold-line`     |         | Minimum line coverage percentage (0-100)                                                               |
| `threshold-branch`   |         | Minimum branch coverage percentage (0-100)                                                             |
| `threshold-function` |         | Minimum function coverage percentage (0-100)                                                           |
| `working-directory`  | `.`     | Working directory for resolving relative paths                                                         |
| `fail-on-error`      | `true`  | Fail the action when thresholds are not met                                                            |
| `suggestions`        | `true`  | Show top coverage improvement opportunities in job summary                                             |

When no thresholds are configured, coverlint reports metrics without enforcing minimums — useful for tracking trends. If a threshold is set but the format doesn't report that metric (e.g. `threshold-branch` with `gocover`), it's skipped with a notice annotation.

### Auto-Detection and Discovery

When `format` is omitted, coverlint tries each parser in priority order (gocover, lcov, jacoco, cobertura, clover) until one succeeds. When `path` is also omitted, it searches common default locations:

| Format      | Searched Paths                                                                                  |
| ----------- | ----------------------------------------------------------------------------------------------- |
| `lcov`      | `coverage/lcov.info`, `lcov.info`, `coverage.lcov`                                              |
| `gocover`   | `cover.out`, `coverage.out`, `c.out`                                                            |
| `cobertura` | `coverage.xml`, `cobertura.xml`, `cobertura-coverage.xml`                                       |
| `clover`    | `coverage.xml`, `clover.xml`                                                                    |
| `jacoco`    | `build/reports/jacoco/test/jacocoTestReport.xml`, `target/site/jacoco/jacoco.xml`, `jacoco.xml` |

Specifying `format` explicitly is recommended — it's faster and avoids ambiguity when files could match multiple formats (e.g. `coverage.xml` could be Cobertura or Clover).

### Outputs

| Output       | Description                                      |
| ------------ | ------------------------------------------------ |
| `passed`     | `true` or `false`                                |
| `results`    | JSON array of per-entry coverage results         |
| `badge-svg`  | SVG badge showing line coverage percentage       |
| `badge-json` | Shields.io endpoint JSON for line coverage badge |

The `results` output contains per-format entries (with a `Total` row for multi-format runs). Fields like `branch` and `function` are omitted when the format doesn't report them:

```json
[
  { "name": "gocover", "line": 85.0, "passed": true },
  { "name": "lcov", "line": 78.3, "branch": 65.2, "function": 90.1, "passed": true },
  { "name": "Total", "line": 81.1, "branch": 65.2, "function": 90.1, "passed": true }
]
```

Access values in subsequent steps with `fromJSON()`:

```yaml
- run: echo "Line coverage is ${{ fromJSON(steps.coverage.outputs.results)[0].line }}%"
```

## Examples

### Per-Language Quick Reference

| Language              | Test Command                                          | Format      | Path                                                   |
| --------------------- | ----------------------------------------------------- | ----------- | ------------------------------------------------------ |
| Go                    | `go test -coverprofile=cover.out ./...`                | `gocover`   | `cover.out`                                            |
| Rust                  | `cargo llvm-cov --lcov --output-path lcov.info`       | `lcov`      | `lcov.info`                                            |
| TypeScript/JavaScript | `npx vitest run --coverage --coverage.reporter=lcov`  | `lcov`      | `coverage/lcov.info`                                   |
| Python                | `pytest --cov --cov-report=xml:coverage.xml`          | `cobertura` | `coverage.xml`                                         |
| PHP                   | `vendor/bin/phpunit --coverage-clover=coverage.xml`   | `clover`    | `coverage.xml`                                         |
| Java (Gradle)         | `./gradlew test jacocoTestReport`                     | `jacoco`    | `build/reports/jacoco/test/jacocoTestReport.xml`       |

Full example with thresholds:

```yaml
- run: go test -coverprofile=cover.out ./...

- uses: evansims/coverlint@v1
  with:
    format: gocover
    threshold-line: 80
    threshold-branch: 70
```

### Monorepo (Multiple Formats)

Combine coverage from different languages in a single step:

```yaml
- uses: evansims/coverlint@v1
  with:
    format: |
      gocover
      lcov
      cobertura
    path: |
      go-service/cover.out
      node-service/coverage/lcov.info
      python-service/coverage.xml
    threshold-line: 80
```

### Multiple Independent Checks

Use separate steps for different thresholds per project area:

```yaml
- uses: evansims/coverlint@v1
  with:
    format: gocover
    path: cover.out
    threshold-line: 80

- uses: evansims/coverlint@v1
  with:
    format: lcov
    path: coverage/lcov.info
    threshold-line: 85
    threshold-branch: 70
```

## Coverage Badges

Coverlint generates badge outputs for live coverage indicators in your README. No external services or secrets required.

Use a two-job workflow so only the badge job gets `contents: write` (principle of least privilege):

```yaml
on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    outputs:
      badge-svg: ${{ steps.coverage.outputs.badge-svg }}
    steps:
      - uses: actions/checkout@v6

      # ... your test steps ...

      - uses: evansims/coverlint@v1
        id: coverage
        with:
          format: gocover
          threshold-line: 80

  update-badges:
    needs: test
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v6

      - name: Push coverage badge
        env:
          BADGE_SVG: ${{ needs.test.outputs.badge-svg }}
        run: |
          tmpdir=$(mktemp -d)
          printf '%s' "$BADGE_SVG" > "$tmpdir/coverage.svg"

          git config user.name "github-actions[bot]"
          git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

          if git ls-remote --exit-code origin badges &>/dev/null; then
            git fetch origin badges
            git checkout badges
          else
            git checkout --orphan badges
            git rm -rf . 2>/dev/null || true
          fi

          cp "$tmpdir/coverage.svg" .
          git add coverage.svg
          git diff --cached --quiet && exit 0
          git commit -m "Update coverage badge"
          git push origin badges
```

Then reference it in your README:

```markdown
![Coverage](https://raw.githubusercontent.com/OWNER/REPO/badges/coverage.svg)
```

For [shields.io](https://shields.io) styling, use `badge-json` output instead:

```markdown
![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/OWNER/REPO/badges/coverage.json)
```

## Contributing

```bash
git clone https://github.com/evansims/coverlint.git
cd coverlint
go test ./...
```

Standard Go tooling: `go test ./...`, `go vet ./...`, `go build ./cmd/coverlint`. Run `go test -race -cover ./...` for race detection. Releases are automated via GoReleaser on version tags.

## License

Dual-licensed under [Apache 2.0](LICENSE-APACHE) and [MIT](LICENSE-MIT). Choose whichever you prefer.
