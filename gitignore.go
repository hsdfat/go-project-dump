package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Gitignore matches paths against a set of .gitignore-style rules.
//
// It supports the most common subset of the gitignore spec:
//   - blank lines and "#" comments are skipped
//   - negation with a leading "!"
//   - directory-only rules with a trailing "/"
//   - anchoring: a pattern containing a non-trailing "/" (or a leading "/")
//     is matched against the whole path relative to the ignore root; a pattern
//     without a slash is matched against the basename at any depth
//   - the "*", "?" and "**" wildcards
//
// The last rule that matches a path wins, mirroring git's own behaviour. It is
// intentionally dependency-free; a handful of rarely-used spec corners (e.g.
// character classes like "[a-z]") are treated literally.
type Gitignore struct {
	rules []gitignoreRule
}

type gitignoreRule struct {
	re      *regexp.Regexp
	negate  bool
	dirOnly bool
}

// NewGitignore builds a matcher from raw .gitignore lines.
func NewGitignore(lines []string) *Gitignore {
	g := &Gitignore{}
	for _, line := range lines {
		rule, ok := compileRule(line)
		if ok {
			g.rules = append(g.rules, rule)
		}
	}
	return g
}

// LoadGitignore reads patterns from the given file. A missing file yields an
// empty (no-op) matcher and no error.
func LoadGitignore(path string) (*Gitignore, error) {
	lines, err := readLines(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Gitignore{}, nil
		}
		return nil, err
	}
	return NewGitignore(lines), nil
}

// readLines reads a file into one string per line, using a generous scanner
// buffer so long lines don't get truncated.
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// Empty reports whether the matcher holds no rules.
func (g *Gitignore) Empty() bool { return len(g.rules) == 0 }

// Match reports whether relPath (slash-separated, relative to the ignore root)
// is ignored. isDir must be true when relPath refers to a directory.
//
// The last matching rule wins. A directory-only rule (trailing "/") matches the
// directory itself only when isDir is true, but always matches files nested
// beneath a directory it covers.
func (g *Gitignore) Match(relPath string, isDir bool) bool {
	relPath = filepath.ToSlash(relPath)
	ignored := false
	for _, rule := range g.rules {
		matched := false
		if rule.re.MatchString(relPath) {
			matched = !rule.dirOnly || isDir
		}
		if !matched && rule.dirOnly && !isDir && matchesAncestor(rule.re, relPath) {
			matched = true // a file beneath an ignored directory
		}
		if matched {
			ignored = !rule.negate
		}
	}
	return ignored
}

// matchesAncestor reports whether any ancestor directory of relPath matches re.
func matchesAncestor(re *regexp.Regexp, relPath string) bool {
	parts := strings.Split(relPath, "/")
	for i := 1; i < len(parts); i++ {
		if re.MatchString(strings.Join(parts[:i], "/")) {
			return true
		}
	}
	return false
}

func compileRule(line string) (gitignoreRule, bool) {
	// Trim a trailing (unescaped) space run and skip comments / blanks.
	line = strings.TrimRight(line, " ")
	if line == "" || strings.HasPrefix(line, "#") {
		return gitignoreRule{}, false
	}

	negate := false
	if strings.HasPrefix(line, "!") {
		negate = true
		line = line[1:]
	}

	dirOnly := strings.HasSuffix(line, "/")
	line = strings.TrimSuffix(line, "/")

	anchored := strings.HasPrefix(line, "/")
	core := strings.TrimPrefix(line, "/")
	if strings.Contains(core, "/") {
		anchored = true
	}
	if core == "" {
		return gitignoreRule{}, false
	}

	return gitignoreRule{
		re:      regexp.MustCompile(globToRegex(core, anchored)),
		negate:  negate,
		dirOnly: dirOnly,
	}, true
}

// globToRegex converts a gitignore glob into an anchored regular expression
// matching a full, slash-separated relative path.
func globToRegex(core string, anchored bool) string {
	var sb strings.Builder
	if anchored {
		sb.WriteString("^")
	} else {
		// Match at the start or after any directory separator.
		sb.WriteString("(?:^|/)")
	}

	for i := 0; i < len(core); {
		c := core[i]
		switch c {
		case '*':
			if i+1 < len(core) && core[i+1] == '*' {
				if i+2 < len(core) && core[i+2] == '/' {
					sb.WriteString("(?:.*/)?") // "**/" -> any number of leading dirs
					i += 3
					continue
				}
				sb.WriteString(".*") // trailing/standalone "**"
				i += 2
				continue
			}
			sb.WriteString("[^/]*") // "*" -> anything within a path segment
			i++
		case '?':
			sb.WriteString("[^/]")
			i++
		default:
			sb.WriteString(regexp.QuoteMeta(string(c)))
			i++
		}
	}

	// Match the end of the path or a directory boundary, so a rule for a
	// directory also matches everything beneath it.
	sb.WriteString("(?:/|$)")
	return sb.String()
}
