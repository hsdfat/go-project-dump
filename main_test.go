package main

import "testing"

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
