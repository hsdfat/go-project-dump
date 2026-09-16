package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// version is overridable at build time with:
//
//	go build -ldflags "-X main.version=v1.2.3"
var version = "0.2.0"

// stringList is a repeatable string flag (e.g. --exclude a --exclude b).
type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	for _, part := range strings.Split(v, ",") {
		if part = strings.TrimSpace(part); part != "" {
			*s = append(*s, part)
		}
	}
	return nil
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(argv []string) int {
	fs := flag.NewFlagSet("go-project-dump", flag.ContinueOnError)

	var (
		output       string
		format       string
		maxSize      string
		noGitignore  bool
		statsOnly    bool
		showVersion  bool
		excludes     stringList
		includes     stringList
		excludeFiles stringList
	)

	fs.StringVar(&output, "output", "", "write output to a file (default: stdout)")
	fs.StringVar(&output, "o", "", "shorthand for --output")
	fs.StringVar(&format, "format", "markdown", "output format: markdown | xml | text")
	fs.StringVar(&format, "f", "markdown", "shorthand for --format")
	fs.StringVar(&maxSize, "max-size", "1MB", "skip files larger than this (e.g. 500KB, 2MB)")
	fs.BoolVar(&noGitignore, "no-gitignore", false, "do not honour each project's .gitignore")
	fs.BoolVar(&statsOnly, "stats", false, "print only the summary, without source code")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")
	fs.Var(&excludes, "exclude", "extra ignore glob (repeatable, e.g. --exclude '*.test.js')")
	fs.Var(&includes, "include", "only include files matching this glob (repeatable)")
	fs.Var(&excludeFiles, "exclude-file", "read extra ignore globs from a file, one per line, gitignore-style (repeatable)")

	fs.Usage = func() { usage(fs) }

	// Parse iteratively so flags may appear before, after, or between paths
	// (stdlib flag otherwise stops at the first positional argument).
	var paths []string
	rest := argv
	for {
		if err := fs.Parse(rest); err != nil {
			return 2 // flag package already printed the error
		}
		rest = fs.Args()
		if len(rest) == 0 {
			break
		}
		paths = append(paths, rest[0])
		rest = rest[1:]
	}

	if showVersion {
		fmt.Printf("go-project-dump %s\n", version)
		return 0
	}

	if len(paths) == 0 {
		usage(fs)
		return 1
	}

	maxBytes, err := parseSize(maxSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid --max-size %q: %v\n", maxSize, err)
		return 1
	}

	for _, ef := range excludeFiles {
		lines, err := readLines(ef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: reading --exclude-file %q: %v\n", ef, err)
			return 1
		}
		excludes = append(excludes, lines...)
	}

	switch strings.ToLower(format) {
	case "markdown", "md", "xml", "text", "txt", "plain":
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown --format %q (want markdown, xml, or text)\n", format)
		return 1
	}

	opts := Options{
		Paths:        paths,
		Output:       output,
		Format:       format,
		MaxFileSize:  maxBytes,
		UseGitignore: !noGitignore,
		Excludes:     excludes,
		Includes:     includes,
		StatsOnly:    statsOnly,
	}

	detector := NewTechnologyDetector()
	report := &Report{GeneratedAt: time.Now().Format("2006-01-02 15:04:05")}

	for _, path := range opts.Paths {
		fmt.Fprintf(os.Stderr, "Analyzing project at: %s\n", path)
		info, err := analyzeProject(path, detector, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing %q: %v\n", path, err)
			return 1
		}
		report.Projects = append(report.Projects, info)
	}

	out := renderReport(report, opts)

	if opts.Output != "" {
		if err := os.WriteFile(opts.Output, []byte(out), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "Output written to: %s (%s, ~%s tokens)\n",
			opts.Output, humanSize(int64(len(out))), humanCount(report.estimatedTokens()))
	} else {
		fmt.Print(out)
	}
	return 0
}

// parseSize parses a human-readable byte size such as "1MB", "512KB", or a
// plain byte count. A value of 0 (or "0") disables the size limit.
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "KB"):
		mult, s = 1024, strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "MB"):
		mult, s = 1024*1024, strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "GB"):
		mult, s = 1024*1024*1024, strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}
	s = strings.TrimSpace(s)
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, fmt.Errorf("negative size")
	}
	return int64(n * float64(mult)), nil
}

func usage(fs *flag.FlagSet) {
	out := fs.Output()
	fmt.Fprintf(out, `go-project-dump %s — compile a project into a single LLM-friendly file.

Usage:
  go-project-dump [flags] <project-path> [more-paths...]

Flags:
  -o, --output <file>    write output to a file (default: stdout)
  -f, --format <fmt>     markdown (default) | xml | text
      --max-size <size>  skip files larger than this (default 1MB; e.g. 500KB, 2MB, 0=off)
      --no-gitignore     do not honour each project's .gitignore
      --exclude <glob>   extra ignore glob, repeatable (e.g. --exclude '*.min.js')
      --exclude-file <path>
                         read extra ignore globs from a file, one per line,
                         gitignore-style (repeatable)
      --include <glob>   only include files matching this glob, repeatable
      --stats            print only the summary (files, size, tokens)
      --version          print version and exit

Examples:
  go-project-dump .
  go-project-dump ./frontend ./backend -o dump.md
  go-project-dump . --format xml -o context.xml
  go-project-dump . --include '*.go' --exclude '*_test.go'
  go-project-dump . --exclude-file .dumpignore
  go-project-dump . --stats
`, version)
}
