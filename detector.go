package main

import (
	"path/filepath"
	"sort"
	"strings"
)

// Technology represents a detected technology in a project.
type Technology struct {
	Name        string
	Files       []string
	Confidence  float64
	Description string
}

// TechnologyDetector scores files against a set of technology fingerprints.
type TechnologyDetector struct {
	patterns map[string]TechPattern
}

// TechPattern describes how to recognise a single technology.
type TechPattern struct {
	Files       []string
	Extensions  []string
	Keywords    []string
	Description string
}

// NewTechnologyDetector returns a detector seeded with the built-in patterns.
func NewTechnologyDetector() *TechnologyDetector {
	patterns := map[string]TechPattern{
		"JavaScript": {
			Files:       []string{"package.json", "package-lock.json", "yarn.lock"},
			Extensions:  []string{".js", ".mjs", ".jsx"},
			Keywords:    []string{"require(", "import ", "export ", "module.exports"},
			Description: "JavaScript runtime and ecosystem",
		},
		"TypeScript": {
			Files:       []string{"tsconfig.json", "tslint.json"},
			Extensions:  []string{".ts", ".tsx"},
			Keywords:    []string{"interface ", "type ", ": string", ": number"},
			Description: "TypeScript - JavaScript with static typing",
		},
		"React": {
			Files:       []string{},
			Extensions:  []string{".jsx", ".tsx"},
			Keywords:    []string{"React.", "useState", "useEffect", "jsx"},
			Description: "React JavaScript library for building user interfaces",
		},
		"Node.js": {
			Files:       []string{"package.json"},
			Extensions:  []string{".js"},
			Keywords:    []string{"require('", "module.exports", "process.env"},
			Description: "Node.js JavaScript runtime",
		},
		"Python": {
			Files:       []string{"requirements.txt", "setup.py", "pyproject.toml", "Pipfile"},
			Extensions:  []string{".py", ".pyw"},
			Keywords:    []string{"def ", "import ", "from ", "__init__"},
			Description: "Python programming language",
		},
		"Go": {
			Files:       []string{"go.mod", "go.sum"},
			Extensions:  []string{".go"},
			Keywords:    []string{"package ", "func ", "import ", "type "},
			Description: "Go programming language",
		},
		"Java": {
			Files:       []string{"pom.xml", "build.gradle", "gradle.properties"},
			Extensions:  []string{".java"},
			Keywords:    []string{"public class", "import java", "package "},
			Description: "Java programming language",
		},
		"C++": {
			Files:       []string{"CMakeLists.txt", "Makefile"},
			Extensions:  []string{".cpp", ".cc", ".cxx", ".h", ".hpp"},
			Keywords:    []string{"#include", "using namespace", "std::"},
			Description: "C++ programming language",
		},
		"C": {
			Files:       []string{"Makefile"},
			Extensions:  []string{".c", ".h"},
			Keywords:    []string{"#include", "int main", "printf"},
			Description: "C programming language",
		},
		"Rust": {
			Files:       []string{"Cargo.toml", "Cargo.lock"},
			Extensions:  []string{".rs"},
			Keywords:    []string{"fn ", "use ", "mod ", "pub "},
			Description: "Rust systems programming language",
		},
		"PHP": {
			Files:       []string{"composer.json", "composer.lock"},
			Extensions:  []string{".php"},
			Keywords:    []string{"<?php", "function ", "$_GET", "$_POST"},
			Description: "PHP server-side scripting language",
		},
		"Ruby": {
			Files:       []string{"Gemfile", "Gemfile.lock"},
			Extensions:  []string{".rb"},
			Keywords:    []string{"def ", "class ", "require ", "end"},
			Description: "Ruby programming language",
		},
		"CSS": {
			Files:       []string{},
			Extensions:  []string{".css", ".scss", ".sass", ".less"},
			Keywords:    []string{"{", "}", ":", ";", "@media"},
			Description: "Cascading Style Sheets",
		},
		"HTML": {
			Files:       []string{},
			Extensions:  []string{".html", ".htm"},
			Keywords:    []string{"<html", "<body", "<div", "<!DOCTYPE"},
			Description: "HyperText Markup Language",
		},
		"Docker": {
			Files:       []string{"Dockerfile", "docker-compose.yml", "docker-compose.yaml", ".dockerignore"},
			Extensions:  []string{},
			Keywords:    []string{"FROM ", "RUN ", "COPY ", "CMD "},
			Description: "Docker containerization platform",
		},
		"Kubernetes": {
			Files:       []string{},
			Extensions:  []string{".yaml", ".yml"},
			Keywords:    []string{"apiVersion:", "kind:", "metadata:", "spec:"},
			Description: "Kubernetes container orchestration",
		},
	}

	return &TechnologyDetector{patterns: patterns}
}

// DetectTechnologies scores the given files and returns the detected
// technologies sorted by descending confidence.
func (td *TechnologyDetector) DetectTechnologies(files []FileInfo) []Technology {
	var technologies []Technology
	techScores := make(map[string]float64)
	techFiles := make(map[string][]string)

	for _, file := range files {
		fileName := filepath.Base(file.Path)
		ext := strings.ToLower(filepath.Ext(file.Path))
		content := strings.ToLower(file.Content)

		for techName, pattern := range td.patterns {
			score := 0.0

			fileMatch := false
			for _, expectedFile := range pattern.Files {
				if fileName == expectedFile {
					fileMatch = true
					break
				}
			}
			if fileMatch {
				score += 3.0
			}

			extMatch := false
			for _, expectedExt := range pattern.Extensions {
				if ext == expectedExt {
					extMatch = true
					break
				}
			}
			if extMatch {
				score += 2.0
			}

			// Only count keywords when the file is plausibly this technology
			// (matching extension or marker file). Otherwise generic tokens
			// like "{", ":", or "import " would let, say, CSS or Python score
			// against every Go source file.
			if extMatch || fileMatch {
				keywordMatches := 0
				for _, keyword := range pattern.Keywords {
					if strings.Contains(content, strings.ToLower(keyword)) {
						keywordMatches++
					}
				}
				score += float64(keywordMatches) * 0.5
			}

			if score > 0 {
				techScores[techName] += score
				techFiles[techName] = append(techFiles[techName], file.Path)
			}
		}
	}

	for techName, score := range techScores {
		if score > 0.5 { // minimum threshold
			confidence := score / 10.0 // normalise to a 0-1 range
			if confidence > 1.0 {
				confidence = 1.0
			}
			technologies = append(technologies, Technology{
				Name:        techName,
				Files:       techFiles[techName],
				Confidence:  confidence,
				Description: td.patterns[techName].Description,
			})
		}
	}

	sort.Slice(technologies, func(i, j int) bool {
		if technologies[i].Confidence != technologies[j].Confidence {
			return technologies[i].Confidence > technologies[j].Confidence
		}
		return technologies[i].Name < technologies[j].Name
	})

	return technologies
}
