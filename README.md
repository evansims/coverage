# Coverlint

A self-contained GitHub Action that enforces coverage thresholds on pull requests. Parses coverage reports, compares against configurable thresholds, and reports results as GitHub Actions annotations and job summaries.

No external services. No GitHub API tokens. No PR comments. Just pass/fail.

## Supported Formats

| Format           | Flag        | Typical Producer                                     |
| ---------------- | ----------- | ---------------------------------------------------- |
| LCOV             | `lcov`      | `cargo llvm-cov`, `c8`, `istanbul`, `jest`, `vitest` |
| Go cover profile | `gocover`   | `go test -coverprofile`                              |
| Cobertura XML    | `cobertura` | `pytest-cov`, `istanbul`, `cargo tarpaulin`          |
| Clover XML       | `clover`    | `phpunit`, some JS tools                             |
| JaCoCo XML       | `jacoco`    | Gradle/Maven JaCoCo plugin                           |

## Usage

Add to your workflow after your test step:

```yaml
- uses: evansims/coverlint@v1
  with:
    path: cover.out
    format: gocover
    threshold-line: 80
```

### Inputs

| Input                | Default | Required | Description                                          |
| -------------------- | ------- | -------- | ---------------------------------------------------- |
| `path`               |         | yes      | Path to coverage report file                         |
| `format`             |         | yes      | One of: `lcov`, `gocover`, `cobertura`, `clover`, `jacoco` |
| `name`               | format  | no       | Display name for annotations                         |
| `threshold-line`     |         | no*      | Minimum line coverage percentage (0-100)             |
| `threshold-branch`   |         | no*      | Minimum branch coverage percentage (0-100)           |
| `threshold-function` |         | no*      | Minimum function coverage percentage (0-100)         |
| `working-directory`  | `.`     | no       | Working directory for resolving relative paths       |
| `fail-on-error`      | `true`  | no       | Fail the action when thresholds are not met          |

*At least one threshold is required.

### Outputs

| Output    | Description                              |
| --------- | ---------------------------------------- |
| `passed`  | `true` or `false`                        |
| `results` | JSON array of per-entry coverage results |

If a threshold is configured but the coverage format doesn't report that metric (e.g., `threshold-branch` with `gocover`), the threshold is skipped and a notice annotation is emitted.

## Example Workflow

```yaml
name: Coverage
on: [pull_request]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod

      - run: go test -coverprofile=cover.out ./...

      - uses: evansims/coverlint@v1
        with:
          path: cover.out
          format: gocover
          threshold-line: 80
```

### Multiple Reports

Use multiple steps to check different coverage reports:

```yaml
- uses: evansims/coverlint@v1
  with:
    path: cover.out
    format: gocover
    name: api
    threshold-line: 80

- uses: evansims/coverlint@v1
  with:
    path: coverage/lcov.info
    format: lcov
    name: frontend
    threshold-line: 85
    threshold-branch: 70
    threshold-function: 80
```

## Contributing

```bash
git clone https://github.com/evansims/coverlint.git
cd coverlint
go test ./...
```

### Development

The project uses standard Go tooling:

- `go test ./...` runs all tests
- `go test -race -cover ./...` runs tests with race detection and coverage
- `go vet ./...` runs static analysis
- `go build ./cmd/coverlint` builds the binary

### Making Changes

1. Fork the repo and create a feature branch
2. Write tests for your changes
3. Run `go test ./...` and `go vet ./...`
4. Submit a pull request

### Releases

Releases are automated via GoReleaser. Pushing a version tag (e.g., `v1.0.0`) triggers cross-compilation and GitHub Release creation.

## License

Dual-licensed under [Apache 2.0](LICENSE-APACHE) and [MIT](LICENSE-MIT). Choose whichever you prefer.
