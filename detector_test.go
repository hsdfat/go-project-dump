package main

import "testing"

func TestDetectTechnologiesGatesKeywords(t *testing.T) {
	// A pure-Go set of files must not score CSS/Python/etc. off generic
	// tokens like "{", ":", or "import ".
	files := []FileInfo{
		{Path: "main.go", Language: "Go", Content: "package main\nimport \"fmt\"\nfunc main() { m := map[string]int{} ; _ = m }\n"},
		{Path: "go.mod", Language: "Text", Content: "module example\n\ngo 1.23\n"},
	}
	techs := NewTechnologyDetector().DetectTechnologies(files)
	if len(techs) == 0 {
		t.Fatal("expected at least Go to be detected")
	}
	if techs[0].Name != "Go" {
		t.Errorf("primary technology = %q, want Go", techs[0].Name)
	}
	for _, tech := range techs {
		switch tech.Name {
		case "CSS", "Python", "Java", "Ruby":
			t.Errorf("%s should not be detected in a pure-Go project", tech.Name)
		}
	}
}

func TestDetectTechnologiesCSS(t *testing.T) {
	files := []FileInfo{
		{Path: "style.css", Language: "CSS", Content: "body { color: red; }\n@media screen { a { color: blue; } }\n"},
	}
	techs := NewTechnologyDetector().DetectTechnologies(files)
	if len(techs) == 0 || techs[0].Name != "CSS" {
		t.Fatalf("expected CSS as primary tech, got %+v", techs)
	}
}
