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
    min-coverage: 80
```

Releases use [immutable tags](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases). [Pin actions by commit SHA](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions#using-third-party-actions) and use [Dependabot](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/about-dependabot-version-updates) to keep them current.

### Inputs

| Input               | Description                                                                                            |
| ------------------- | ------------------------------------------------------------------------------------------------------ |
| `format`            | Coverage format(s), one per line or comma-separated. Auto-detected if omitted                          |
| `path`              | Path(s) to coverage files, one per line or comma-separated. Supports globs. Auto-discovered if omitted |
| `min-coverage`      | Minimum weighted coverage score (0-100), computed from line, branch, and function coverage             |
| `min-line`          | Minimum line coverage (0-100), checked independently of the weighted score                             |
| `min-branch`        | Minimum branch coverage (0-100), checked independently                                                 |
| `min-function`      | Minimum function coverage (0-100), checked independently                                               |
| `weight-line`       | Relative weight for line coverage in score (default: `50`)                                             |
| `weight-branch`     | Relative weight for branch coverage in score (default: `30`)                                           |
| `weight-function`   | Relative weight for function coverage in score (default: `20`)                                         |
| `working-directory` | Working directory for resolving relative paths (default: `.`)                                          |
| `fail-on-error`     | Fail the action when minimums are not met (default: `true`)                                            |
| `suggestions`       | Show top coverage improvement opportunities in job summary (default: `true`)                           |

`min-coverage` checks a weighted score computed from line, branch, and function coverage (default weights: 50/30/20). If a metric isn't reported by your format (e.g. `gocover` doesn't report branch), its weight redistributes proportionally. Use `min-line`, `min-branch`, or `min-function` to enforce limits on individual metrics — they're checked independently of the score.

Without minimums, coverlint reports coverage without failing — useful for tracking trends. If you set a minimum that your coverage format doesn't support (e.g. `min-branch` with `gocover`), it's skipped with a notice.

### Auto-Detection and Discovery

You don't need to specify `format` or `path` — coverlint can figure both out. It tries each parser until one succeeds, and looks for reports in common locations:

| Format      | Searched Paths                                                                                  |
| ----------- | ----------------------------------------------------------------------------------------------- |
| `lcov`      | `coverage/lcov.info`, `lcov.info`, `coverage.lcov`                                              |
| `gocover`   | `cover.out`, `coverage.out`, `c.out`                                                            |
| `cobertura` | `coverage.xml`, `cobertura.xml`, `cobertura-coverage.xml`                                       |
| `clover`    | `coverage.xml`, `clover.xml`                                                                    |
| `jacoco`    | `build/reports/jacoco/test/jacocoTestReport.xml`, `target/site/jacoco/jacoco.xml`, `jacoco.xml` |

Setting `format` explicitly is still a good idea — it's faster and avoids guesswork when files share names (e.g. `coverage.xml` could be Cobertura or Clover).

### Outputs

| Output       | Description                                                      |
| ------------ | ---------------------------------------------------------------- |
| `passed`     | Whether all minimums were met (`true` or `false`)                |
| `results`    | Coverage data as JSON (see below)                                |
| `badge-svg`  | Ready-to-use SVG coverage badge                                  |
| `badge-json` | Coverage badge as [shields.io](https://shields.io) endpoint JSON |

The `results` JSON has one entry per format, each with a weighted `score` and available metrics. Multi-format runs include a `Total`:

```json
[
  { "name": "gocover", "score": 85.0, "line": 85.0, "passed": true },
  {
    "name": "lcov",
    "score": 77.4,
    "line": 78.3,
    "branch": 65.2,
    "function": 90.1,
    "passed": true
  },
  {
    "name": "Total",
    "score": 79.2,
    "line": 81.1,
    "branch": 65.2,
    "function": 90.1,
    "passed": true
  }
]
```

Use `fromJSON()` to read values in later steps:

```yaml
- run: echo "Line coverage is ${{ fromJSON(steps.coverage.outputs.results)[0].line }}%"
```

## Examples

### Quick Reference

| Language              | Test Command                                         | Format      | Path                                             |
| --------------------- | ---------------------------------------------------- | ----------- | ------------------------------------------------ |
| Go                    | `go test -coverprofile=cover.out ./...`              | `gocover`   | `cover.out`                                      |
| Rust                  | `cargo llvm-cov --lcov --output-path lcov.info`      | `lcov`      | `lcov.info`                                      |
| TypeScript/JavaScript | `npx vitest run --coverage --coverage.reporter=lcov` | `lcov`      | `coverage/lcov.info`                             |
| Python                | `pytest --cov --cov-report=xml:coverage.xml`         | `cobertura` | `coverage.xml`                                   |
| PHP                   | `vendor/bin/phpunit --coverage-clover=coverage.xml`  | `clover`    | `coverage.xml`                                   |
| Java (Gradle)         | `./gradlew test jacocoTestReport`                    | `jacoco`    | `build/reports/jacoco/test/jacocoTestReport.xml` |

```yaml
- run: go test -coverprofile=cover.out ./...

- uses: evansims/coverlint@v1
  with:
    format: gocover
    min-coverage: 80
```

### Monorepo

Combine coverage from multiple languages in one step — the job summary breaks down each format with a combined total:

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
    min-coverage: 80
```

### Different Minimums Per Metric

Use `min-coverage` for the overall bar and `min-*` for individual metrics that need their own limits:

```yaml
- uses: evansims/coverlint@v1
  with:
    format: lcov
    min-coverage: 80
    min-branch: 60 # fails if branch drops below 60%, even if the overall score passes
```

### Custom Score Weights

The coverage score weights line (50), branch (30), and function (20) by default. Weights are relative — adjust them to match what matters to your project:

```yaml
- uses: evansims/coverlint@v1
  with:
    format: lcov
    min-coverage: 80
    weight-line: 100 # only line coverage counts
    weight-branch: 0
    weight-function: 0
```

### Different Minimums Per Area

Use separate steps when parts of your project need different bars:

```yaml
- uses: evansims/coverlint@v1
  with:
    format: gocover
    path: cover.out
    min-coverage: 80

- uses: evansims/coverlint@v1
  with:
    format: lcov
    path: coverage/lcov.info
    min-coverage: 85
```

## Coverage Badges

Show live coverage in your README — no external services or secrets needed.

The workflow below uses two jobs so that only the badge job gets write access. The test job runs on every push and PR with read-only permissions; the badge job only runs on `main`:

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
          min-coverage: 80

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

Add to your README:

```markdown
![Coverage](https://raw.githubusercontent.com/OWNER/REPO/badges/coverage.svg)
```

Prefer [shields.io](https://shields.io) styling? Use `badge-json` instead:

```markdown
![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/OWNER/REPO/badges/coverage.json)
```

## Contributing

Clone and run the tests — standard Go tooling, nothing extra needed:

```bash
git clone https://github.com/evansims/coverlint.git && cd coverlint
go test -race -cover ./...
go vet ./...
```

## License

Dual-licensed under [Apache 2.0](LICENSE-APACHE) and [MIT](LICENSE-MIT). Choose whichever you prefer.
