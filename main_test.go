package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"1MB", 1024 * 1024, false},
		{"512KB", 512 * 1024, false},
		{"2GB", 2 * 1024 * 1024 * 1024, false},
		{"1024", 1024, false},
		{"1024B", 1024, false},
		{"1.5MB", int64(1.5 * 1024 * 1024), false},
		{"0", 0, false},
		{" 2mb ", 2 * 1024 * 1024, false},
		{"", 0, true},
		{"abc", 0, true},
		{"-5MB", 0, true},
	}
	for _, tt := range tests {
		got, err := parseSize(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseSize(%q) expected error, got %d", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseSize(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseSize(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestStringListFlag(t *testing.T) {
	var s stringList
	_ = s.Set("*.go")
	_ = s.Set("a.js,b.js")   // comma-splitting
	_ = s.Set("  , spaced ") // trims and drops blanks
	want := []string{"*.go", "a.js", "b.js", "spaced"}
	if len(s) != len(want) {
		t.Fatalf("got %v, want %v", s, want)
	}
	for i := range want {
		if s[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, s[i], want[i])
		}
	}
}

func TestRunVersion(t *testing.T) {
	if code := run([]string{"--version"}); code != 0 {
		t.Errorf("run(--version) = %d, want 0", code)
	}
}

func TestRunNoArgs(t *testing.T) {
	if code := run(nil); code != 1 {
		t.Errorf("run() with no paths = %d, want 1", code)
	}
}

func TestRunBadFormat(t *testing.T) {
	dir := t.TempDir()
	if code := run([]string{"--format", "yaml", dir}); code != 1 {
		t.Errorf("run(--format yaml) = %d, want 1", code)
	}
}

func TestRunExcludeFile(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "keep.go"), "package main\n")
	writeTestFile(t, filepath.Join(dir, "gen.pb.go"), "package main\n")
	writeTestFile(t, filepath.Join(dir, "vendor_extra", "dep.go"), "package vendor_extra\n")

	excludeFile := filepath.Join(dir, ".dumpignore")
	writeTestFile(t, excludeFile, "# generated code\n*.pb.go\nvendor_extra/\n")

	out := filepath.Join(dir, "dump.md")
	if code := run([]string{"--exclude-file", excludeFile, "-o", out, dir}); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}

	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	got := string(content)
	if !strings.Contains(got, "keep.go") {
		t.Error("expected keep.go to be included")
	}
	if strings.Contains(got, "gen.pb.go") {
		t.Error("expected gen.pb.go to be excluded by *.pb.go rule")
	}
	if strings.Contains(got, "dep.go") {
		t.Error("expected vendor_extra/dep.go to be excluded by vendor_extra/ rule")
	}
}

func TestRunExcludeFileMissing(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "keep.go"), "package main\n")

	missing := filepath.Join(dir, "does-not-exist.ignore")
	if code := run([]string{"--exclude-file", missing, dir}); code != 1 {
		t.Errorf("run(--exclude-file missing) = %d, want 1", code)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
