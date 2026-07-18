package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := map[string]string{
		".go":      "Go",
		".ts":      "TypeScript",
		".py":      "Python",
		".unknown": "Text",
		"":         "Text",
	}
	for ext, want := range tests {
		if got := detectLanguage(ext); got != want {
			t.Errorf("detectLanguage(%q) = %q, want %q", ext, got, want)
		}
	}
}

func TestIsBinaryFile(t *testing.T) {
	if isBinaryFile([]byte("plain text, no nulls")) {
		t.Error("text wrongly detected as binary")
	}
	if !isBinaryFile([]byte{'a', 0x00, 'b'}) {
		t.Error("content with a null byte should be binary")
	}
	if isBinaryFile(nil) {
		t.Error("empty content should not be binary")
	}
}

// writeFile is a small fixture helper.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func collectPaths(info *ProjectInfo) map[string]bool {
	m := map[string]bool{}
	for _, f := range info.Files {
		m[f.Path] = true
	}
	return m
}

func TestAnalyzeProjectGitignore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\nfunc main() {}\n")
	writeFile(t, root, "secret.env", "TOKEN=abc")
	writeFile(t, root, ".gitignore", "*.env\nbuild/\n")
	writeFile(t, root, "build/artifact.go", "package build")
	writeFile(t, root, "pkg/util.go", "package pkg")

	detector := NewTechnologyDetector()
	info, err := analyzeProject(root, detector, Options{MaxFileSize: 1 << 20, UseGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	got := collectPaths(info)

	if !got["main.go"] || !got["pkg/util.go"] {
		t.Errorf("expected source files included, got %v", got)
	}
	if got["secret.env"] {
		t.Error(".env should be gitignored")
	}
	if got["build/artifact.go"] {
		t.Error("files under build/ should be gitignored")
	}
	if info.Language != "Go" {
		t.Errorf("primary language = %q, want Go", info.Language)
	}
}

func TestAnalyzeProjectNoGitignore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "keep.go", "package main")
	writeFile(t, root, "secret.env", "TOKEN=abc")
	writeFile(t, root, ".gitignore", "*.env")

	info, err := analyzeProject(root, NewTechnologyDetector(),
		Options{MaxFileSize: 1 << 20, UseGitignore: false})
	if err != nil {
		t.Fatal(err)
	}
	if !collectPaths(info)["secret.env"] {
		t.Error("with --no-gitignore, .env should be included")
	}
}

func TestAnalyzeProjectIncludeExclude(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.go", "package main")
	writeFile(t, root, "a_test.go", "package main")
	writeFile(t, root, "b.js", "console.log(1)")

	info, err := analyzeProject(root, NewTechnologyDetector(), Options{
		MaxFileSize:  1 << 20,
		UseGitignore: true,
		Includes:     []string{"*.go"},
		Excludes:     []string{"*_test.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := collectPaths(info)
	if !got["a.go"] {
		t.Error("a.go should be included")
	}
	if got["a_test.go"] {
		t.Error("a_test.go should be excluded")
	}
	if got["b.js"] {
		t.Error("b.js should not match the include glob")
	}
}

func TestAnalyzeProjectMaxSize(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "small.go", "package main")
	big := make([]byte, 2048)
	for i := range big {
		big[i] = 'x'
	}
	writeFile(t, root, "big.txt", string(big))

	info, err := analyzeProject(root, NewTechnologyDetector(),
		Options{MaxFileSize: 1024, UseGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	got := collectPaths(info)
	if !got["small.go"] {
		t.Error("small file should be included")
	}
	if got["big.txt"] {
		t.Error("file over max-size should be skipped")
	}
}

func TestAnalyzeProjectSkipsBinary(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "code.go", "package main")
	writeFile(t, root, "blob.dat", "abc\x00def") // null byte -> binary

	info, err := analyzeProject(root, NewTechnologyDetector(),
		Options{MaxFileSize: 1 << 20, UseGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	if collectPaths(info)["blob.dat"] {
		t.Error("binary file should be skipped")
	}
}

func TestAnalyzeProjectMissingPath(t *testing.T) {
	_, err := analyzeProject(filepath.Join(t.TempDir(), "nope"),
		NewTechnologyDetector(), Options{MaxFileSize: 1 << 20})
	if err == nil {
		t.Error("expected error for missing path")
	}
}
