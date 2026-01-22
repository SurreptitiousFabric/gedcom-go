# Codex CLI Guidance

This repository is `gedcom-go`, a pure Go library for parsing, validating, and encoding GEDCOM files.

## Quick start
- Go 1.21+ required.
- `make setup-dev-env` downloads deps, installs tools, and sets hooks.
- Hooks: pre-commit runs gofmt/go vet/golangci-lint/tests; pre-push checks coverage >=85%.

## Common commands
| Command | Description |
| --- | --- |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make fmt` | Format code |
| `make vet` | Run go vet |
| `make lint` | Run staticcheck |
| `make bench` | Run benchmarks |

## Architecture
Packages:
- `gedcom/` core data types (Document, Individual, Family, Source, etc.)
- `decoder/` high-level decoding with version detection
- `encoder/` GEDCOM writing with line ending control
- `parser/` low-level line parsing with detailed errors
- `validator/` semantic validation rules
- `charset/` encoding (UTF-8, ANSEL) with BOM detection
- `version/` GEDCOM version detection (5.5, 5.5.1, 7.0)

Data flow:
```
GEDCOM file -> charset.NewReader() -> parser.Parse() -> decoder.buildDocument() -> gedcom.Document
```

## Key types and lookup
- `gedcom.Document` holds `Header`, `Records`, and `XRefMap` for O(1) lookups.
- Cross-reference resolution uses XRef strings (for example, `@I1@`).

Example:
```go
family := doc.GetFamily("@F1@")
husband := doc.GetIndividual(family.Husband)
```

## Test data and coverage
- Sample GEDCOM files live in `testdata/`.
- Coverage requirements: per-package >=85%, critical paths at 100%. See `docs/TESTING.md`.

## Documentation structure
- `README.md` overview and quick start
- `FEATURES.md` implemented feature list
- `IDEAS.md` unvetted ideas
- `docs/` implementation references and ADRs
- `specs/` feature specs and plans

## Governance
- The constitution at `.specify/memory/constitution.md` is authoritative.
- If guidance conflicts, the constitution wins.

## Git conventions
- Use conventional commits for library changes (see `CONTRIBUTING.md`).
- PR titles must NOT use conventional commit format.
- Rebase feature branches on `main` and use merge commits (no squash).

## Downstream consumer
- This library is used by `github.com/cacack/my-family` via a `replace` directive.
- If that repo is available, run its tests after changes: `go test ./...`.
