package main

import "testing"

func TestGitignoreMatch(t *testing.T) {
	lines := []string{
		"# a comment",
		"",
		"*.log",          // basename glob, any depth
		"build/",         // directory only
		"/root-only.txt", // anchored to root
		"src/generated",  // anchored path
		"docs/**/tmp",    // ** wildcard
		"!important.log", // negation
		"node_modules",   // bare name, any depth
	}
	g := NewGitignore(lines)

	tests := []struct {
		path  string
		isDir bool
		want  bool
	}{
		{"app.log", false, true},
		{"sub/dir/app.log", false, true},
		{"important.log", false, false},       // negated
		{"sub/important.log", false, false},   // negated at depth
		{"build", true, true},                 // dir-only matches the dir
		{"build/output.js", false, true},      // and everything under it
		{"src/build", false, false},           // build/ is dir-only; this is a file named build
		{"root-only.txt", false, true},        // anchored, at root
		{"sub/root-only.txt", false, false},   // anchored, not at root
		{"src/generated", true, true},         // anchored path
		{"src/generated/x.go", false, true},   // under anchored path
		{"other/src/generated", false, false}, // anchored means from root
		{"docs/a/b/tmp", true, true},          // ** matches nested
		{"docs/tmp", true, true},              // ** matches zero dirs
		{"node_modules/react/index.js", false, true},
		{"packages/x/node_modules/y.js", false, true}, // bare name at depth
		{"readme.md", false, false},                   // matches nothing
	}

	for _, tt := range tests {
		if got := g.Match(tt.path, tt.isDir); got != tt.want {
			t.Errorf("Match(%q, isDir=%v) = %v, want %v", tt.path, tt.isDir, got, tt.want)
		}
	}
}

func TestGitignoreEmpty(t *testing.T) {
	g := NewGitignore([]string{"# only a comment", "", "   "})
	if !g.Empty() {
		t.Error("expected empty matcher for comment/blank-only input")
	}
	if g.Match("anything.go", false) {
		t.Error("empty matcher should not match anything")
	}
}

func TestGitignoreNegationOrder(t *testing.T) {
	// Later rules win: ignore all .env, then re-include the example.
	g := NewGitignore([]string{"*.env", "!.env.example"})
	if !g.Match("prod.env", false) {
		t.Error("prod.env should be ignored")
	}
	if g.Match(".env.example", false) {
		t.Error(".env.example should be re-included by the negation")
	}
}
