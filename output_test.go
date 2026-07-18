package main

import (
	"strings"
	"testing"
)

func sampleReport() *Report {
	return &Report{
		GeneratedAt: "2026-07-18 10:00:00",
		Projects: []*ProjectInfo{
			{
				Name: "svc", Path: "/tmp/svc", Language: "Go",
				TotalFiles: 2, ProcessedFiles: 1, Size: 20,
				Technologies: []Technology{
					{Name: "Go", Confidence: 0.9, Description: "Go programming language", Files: []string{"main.go"}},
				},
				Files: []FileInfo{
					{Path: "main.go", Project: "svc", Size: 20, Content: "package main\n", Language: "Go"},
				},
			},
		},
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
	}
	for _, tt := range tests {
		if got := humanSize(tt.in); got != tt.want {
			t.Errorf("humanSize(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGenerateDirectoryTree(t *testing.T) {
	files := []FileInfo{
		{Path: "main.go"},
		{Path: "pkg/util.go"},
		{Path: "pkg/util_test.go"},
	}
	tree := generateDirectoryTree(files)
	for _, want := range []string{"main.go", "pkg", "util.go", "util_test.go"} {
		if !strings.Contains(tree, want) {
			t.Errorf("tree missing %q:\n%s", want, tree)
		}
	}
}

func TestRenderMarkdown(t *testing.T) {
	out := renderMarkdown(sampleReport(), Options{Format: "markdown"})
	for _, want := range []string{
		"# go-project-dump Analysis",
		"**Primary Language:** Go",
		"Estimated Tokens:",
		"```go",
		"package main",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("markdown output missing %q", want)
		}
	}
}

func TestRenderMarkdownStatsOnly(t *testing.T) {
	out := renderMarkdown(sampleReport(), Options{Format: "markdown", StatsOnly: true})
	if strings.Contains(out, "package main") {
		t.Error("stats-only output should not contain source code")
	}
	if !strings.Contains(out, "Estimated Tokens:") {
		t.Error("stats-only output should still contain the summary")
	}
}

func TestRenderXML(t *testing.T) {
	out := renderXML(sampleReport(), Options{Format: "xml"})
	for _, want := range []string{
		"<project_dump generated=\"2026-07-18 10:00:00\">",
		"<estimated_tokens>",
		"<file path=\"main.go\"",
		"<![CDATA[package main",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("xml output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderXMLEscaping(t *testing.T) {
	r := sampleReport()
	r.Projects[0].Files[0].Content = "if a < b && c > d { }"
	r.Projects[0].Files[0].Path = "a&b.go"
	out := renderXML(r, Options{Format: "xml"})
	if !strings.Contains(out, "path=\"a&amp;b.go\"") {
		t.Error("xml attribute not escaped")
	}
	// Content lives inside CDATA, so raw < and > are preserved verbatim.
	if !strings.Contains(out, "if a < b && c > d") {
		t.Error("CDATA content should be preserved verbatim")
	}
}

func TestCdata(t *testing.T) {
	// Plain content is wrapped verbatim.
	if got, want := cdata("hello"), "<![CDATA[hello]]>"; got != want {
		t.Errorf("cdata(plain) = %q, want %q", got, want)
	}
	// An embedded "]]>" is split into two sections so the XML stays well-formed.
	if got, want := cdata("a]]>b"), "<![CDATA[a]]]]><![CDATA[>b]]>"; got != want {
		t.Errorf("cdata(with terminator) = %q, want %q", got, want)
	}
}

func TestRenderText(t *testing.T) {
	out := renderText(sampleReport(), Options{Format: "text"})
	for _, want := range []string{"FILE: main.go", "Estimated tokens:", "package main"} {
		if !strings.Contains(out, want) {
			t.Errorf("text output missing %q", want)
		}
	}
}

func TestMultiProjectSummary(t *testing.T) {
	r := sampleReport()
	r.Projects = append(r.Projects, &ProjectInfo{
		Name: "web", Path: "/tmp/web", Language: "TypeScript",
		TotalFiles: 3, ProcessedFiles: 2, Size: 40,
		Technologies: []Technology{{Name: "TypeScript", Confidence: 0.8, Description: "TS"}},
		Files: []FileInfo{
			{Path: "app.ts", Project: "web", Size: 40, Content: "export const x = 1\n", Language: "TypeScript"},
		},
	})
	out := renderMarkdown(r, Options{Format: "markdown"})
	for _, want := range []string{
		"Multi-Project Analysis",
		"**Number of Projects:** 2",
		"Individual Project Details",
		"#### svc",
		"#### web",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("multi-project markdown missing %q", want)
		}
	}
	if r.totalFiles() != 5 || r.processedFiles() != 3 {
		t.Errorf("aggregates wrong: total=%d processed=%d", r.totalFiles(), r.processedFiles())
	}
}

func TestCombineTechnologies(t *testing.T) {
	projects := []*ProjectInfo{
		{Technologies: []Technology{{Name: "Go", Confidence: 0.5, Files: []string{"a.go"}}}},
		{Technologies: []Technology{{Name: "Go", Confidence: 0.9, Files: []string{"b.go"}}}},
	}
	got := combineTechnologies(projects)
	if len(got) != 1 {
		t.Fatalf("expected 1 merged technology, got %d", len(got))
	}
	if got[0].Confidence != 0.9 {
		t.Errorf("merged confidence = %v, want 0.9 (max)", got[0].Confidence)
	}
	if len(got[0].Files) != 2 {
		t.Errorf("merged files = %v, want union of 2", got[0].Files)
	}
}
