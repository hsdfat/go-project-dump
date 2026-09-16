# go-project-dump 🚀

[![CI](https://github.com/hsdfat/go-project-dump/actions/workflows/ci.yml/badge.svg)](https://github.com/hsdfat/go-project-dump/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hsdfat/go-project-dump)](https://goreportcard.com/report/github.com/hsdfat/go-project-dump)
[![Go Reference](https://pkg.go.dev/badge/github.com/hsdfat/go-project-dump.svg)](https://pkg.go.dev/github.com/hsdfat/go-project-dump)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
![Zero dependencies](https://img.shields.io/badge/dependencies-0-brightgreen)

**Pack an entire codebase into a single file — ready to paste into an LLM.**

`go-project-dump` walks one or more project directories, skips the noise
(`.gitignore`, `node_modules`, binaries, build artifacts), detects the
technologies in use, and writes everything to one clean file with an
**estimated token count** so you know whether it fits your model's context
window. A single static Go binary, **zero external dependencies**.

Great for feeding code to Claude / ChatGPT, code reviews, documentation,
onboarding, and archiving a snapshot of a project.

## ✨ Features

- **🤖 LLM-first** — estimated token count up front, plus an `xml` output mode
  optimized for models that parse structured context well.
- **🧾 Respects `.gitignore`** — honours your project's ignore rules out of the
  box (disable with `--no-gitignore`), on top of sensible built-in excludes.
- **🗂️ Multi-project** — pass several paths to dump a whole microservices repo
  into one file with a combined summary and per-project breakdown.
- **🎛️ Precise control** — `--include` / `--exclude` globs, `--max-size`, and a
  `--stats` mode that prints just the summary.
- **🔍 Technology detection** — identifies 15+ technologies (Go, TypeScript,
  Python, React, Docker, …) with confidence scores.
- **🌳 Directory tree + syntax highlighting** in the Markdown output.
- **⚡ Fast, single binary, no dependencies.**

## 🚀 Installation

Install with Go (1.21+):

```bash
go install github.com/hsdfat/go-project-dump@latest
```

Or clone and build:

```bash
git clone https://github.com/hsdfat/go-project-dump.git
cd go-project-dump
go build -o go-project-dump .
```

## ⚡ Quick start

```bash
# Dump the current directory to stdout (Markdown)
go-project-dump .

# Write to a file
go-project-dump . -o dump.md

# Build LLM context from just your Go sources, minus tests
go-project-dump . --include '*.go' --exclude '*_test.go' -o context.md

# XML output, tuned for pasting into an LLM
go-project-dump . --format xml -o context.xml

# Two services into one combined file
go-project-dump ./frontend ./backend -o fullstack.md

# Load extra exclude globs from a file, one per line (gitignore-style)
go-project-dump . --exclude-file .dumpignore

# Just the numbers (files, size, estimated tokens) — no source
go-project-dump . --stats
```

Flags may appear **before, after, or between** paths.

## 🎛️ Command-line reference

```
Usage:
  go-project-dump [flags] <project-path> [more-paths...]

Flags:
  -o, --output <file>    write output to a file (default: stdout)
  -f, --format <fmt>     markdown (default) | xml | text
      --max-size <size>  skip files larger than this (default 1MB; e.g. 500KB, 2MB, 0=off)
      --no-gitignore     do not honour each project's .gitignore
      --exclude <glob>   extra ignore glob, repeatable (e.g. --exclude '*.min.js')
      --exclude-file <path>  read extra ignore globs from a file, one per
                         line, gitignore-style (repeatable)
      --include <glob>   only include files matching this glob, repeatable
      --stats            print only the summary (files, size, tokens)
      --version          print version and exit
```

## 📦 Output formats

| Format     | Flag             | Best for |
|------------|------------------|----------|
| `markdown` | *(default)*      | Human-readable reviews, docs, GitHub |
| `xml`      | `--format xml`   | Feeding to LLMs — file boundaries are unambiguous tags, contents wrapped in `CDATA` |
| `text`     | `--format text`  | Minimal, greppable, smallest token overhead |

## 🚫 What gets filtered

Files are included unless they match one of these:

1. **`.gitignore`** rules in each project root (unless `--no-gitignore`).
2. **Built-in directory excludes**: `node_modules`, `vendor`, `.git`, `dist`,
   `build`, `target`, `__pycache__`, `.idea`, `.vscode`, and more.
3. **Binary / media / archive extensions** (`.exe`, `.png`, `.zip`, `.pdf`, …)
   and any file containing null bytes.
4. **Size limit** — files larger than `--max-size` (default 1 MB; `0` disables).
5. **`--exclude` globs** (including any loaded via **`--exclude-file`**), then
   **`--include` globs** (if `--include` is set, only matching files are
   kept).

Unreadable files are skipped with a warning rather than aborting the run.

## 🔢 Token estimation

The summary reports an **estimated** token count using the common ~4-characters-
per-token heuristic, which tracks closely with the tokenizers used by Claude and
GPT-family models for source code. It is meant to tell you whether a dump fits in
a context window — not to be billed against.

## 🧪 Example

```console
$ go-project-dump . --include '*.go' --stats
# go-project-dump Analysis

**Generated on:** 2026-07-18 14:04:04

**Analyzed Projects:**
- .

## Project Summary

- **Primary Language:** Go
- **Total Files:** 18
- **Processed Files:** 12
- **Processed Size:** 52.38 KB
- **Estimated Tokens:** ~13,413
```

## 🔧 Supported technologies

Go, JavaScript, TypeScript, Node.js, React, Python, Java, C, C++, Rust, PHP,
Ruby, CSS (incl. SCSS/Sass/Less), HTML, Docker, and Kubernetes — with more easy
to add in [`detector.go`](detector.go).

## 🤝 Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md). Please keep
the project **dependency-free** and run `gofmt`, `go vet`, and `go test` before
opening a PR (CI enforces all three).

## 📄 License

Licensed under the [MIT License](LICENSE).
