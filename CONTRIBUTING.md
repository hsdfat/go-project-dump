# Contributing to go-project-dump

Thanks for your interest in improving go-project-dump! Contributions of all
sizes are welcome — bug reports, docs, new technology fingerprints, and features.

## Getting started

```bash
git clone https://github.com/hsdfat/go-project-dump.git
cd go-project-dump
go build ./...
go test ./...
```

The project has **zero external dependencies** — please keep it that way. If a
change seems to need a third-party module, open an issue first so we can discuss
whether it belongs in the standard library subset instead.

## Before opening a pull request

Please make sure the following all pass locally — CI runs the same checks:

```bash
gofmt -l .        # should print nothing
go vet ./...
go test -race ./...
```

- Keep changes focused; one logical change per PR.
- Add or update tests for any behaviour you change. New `TechPattern`s should
  come with a small test in `detector_test.go`.
- Match the existing code style (idiomatic Go, short doc comments on exported
  identifiers).

## Adding a new technology

Technology detection lives in `detector.go`. Each entry is a `TechPattern` with:

- `Files` — marker filenames (e.g. `Cargo.toml`), scored highest
- `Extensions` — file extensions (e.g. `.rs`)
- `Keywords` — tokens searched **only** in files that already match by extension
  or marker file, so keep them specific to avoid false positives
- `Description` — a one-line human description

## Reporting bugs

Open an issue with:

- what you ran (the exact command),
- what you expected,
- what happened instead (include output where possible), and
- your OS and `go version`.

## License

By contributing, you agree that your contributions will be licensed under the
project's [MIT License](LICENSE).
