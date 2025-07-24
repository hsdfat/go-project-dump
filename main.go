package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Technology represents a detected technology in the project
type Technology struct {
	Name        string
	Files       []string
	Confidence  float64
	Description string
}

// ProjectInfo holds information about the analyzed project
type ProjectInfo struct {
	Path           string
	Technologies   []Technology
	TotalFiles     int
	ProcessedFiles int
	Size           int64
	Language       string
}

// FileInfo represents information about a single file
type FileInfo struct {
	Path     string
	Size     int64
	Content  string
	Language string
}

// TechnologyDetector handles technology detection logic
type TechnologyDetector struct {
	patterns map[string]TechPattern
}

type TechPattern struct {
	Files       []string
	Extensions  []string
	Keywords    []string
	Description string
}

// Common file extensions to ignore
var ignoredExtensions = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".dylib": true,
	".zip": true, ".tar": true, ".gz": true, ".rar": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
	".mp4": true, ".avi": true, ".mov": true, ".mp3": true, ".wav": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".class": true, ".jar": true, ".war": true,
	".o": true, ".obj": true, ".lib": true, ".a": true,
	".pyc": true, ".pyo": true, ".pyd": true,
}

// Directories to ignore
var ignoredDirs = map[string]bool{
	"node_modules": true, ".git": true, ".svn": true, ".hg": true,
	"vendor": true, "__pycache__": true, ".idea": true, ".vscode": true,
	"build": true, "dist": true, "target": true, "bin": true, "obj": true,
	".next": true, ".nuxt": true, "coverage": true, ".nyc_output": true,
	"logs": true, "tmp": true, "temp": true,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: projectdump <project-path> [output-file]")
		fmt.Println("  project-path: Path to the project directory")
		fmt.Println("  output-file:  Optional output file (default: stdout)")
		os.Exit(1)
	}

	projectPath := os.Args[1]
	outputFile := ""
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	// Initialize detector
	detector := NewTechnologyDetector()

	// Analyze project
	fmt.Fprintf(os.Stderr, "Analyzing project at: %s\n", projectPath)
	projectInfo, files, err := analyzeProject(projectPath, detector)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing project: %v\n", err)
		os.Exit(1)
	}

	// Generate output
	output := generateOutput(projectInfo, files)

	// Write output
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Output written to: %s\n", outputFile)
	} else {
		fmt.Print(output)
	}
}

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

func (td *TechnologyDetector) DetectTechnologies(projectPath string, files []FileInfo) []Technology {
	var technologies []Technology
	techScores := make(map[string]float64)
	techFiles := make(map[string][]string)

	// Check each file against patterns
	for _, file := range files {
		fileName := filepath.Base(file.Path)
		ext := strings.ToLower(filepath.Ext(file.Path))
		content := strings.ToLower(file.Content)

		for techName, pattern := range td.patterns {
			score := 0.0

			// Check specific files
			for _, expectedFile := range pattern.Files {
				if fileName == expectedFile {
					score += 3.0
					break
				}
			}

			// Check extensions
			for _, expectedExt := range pattern.Extensions {
				if ext == expectedExt {
					score += 2.0
					break
				}
			}

			// Check keywords in content
			keywordMatches := 0
			for _, keyword := range pattern.Keywords {
				if strings.Contains(content, strings.ToLower(keyword)) {
					keywordMatches++
				}
			}
			score += float64(keywordMatches) * 0.5

			if score > 0 {
				techScores[techName] += score
				techFiles[techName] = append(techFiles[techName], file.Path)
			}
		}
	}

	// Convert scores to technologies
	for techName, score := range techScores {
		if score > 0.5 { // Minimum threshold
			confidence := score / 10.0 // Normalize to 0-1 range
			if confidence > 1.0 {
				confidence = 1.0
			}

			tech := Technology{
				Name:        techName,
				Files:       techFiles[techName],
				Confidence:  confidence,
				Description: td.patterns[techName].Description,
			}
			technologies = append(technologies, tech)
		}
	}

	// Sort by confidence
	sort.Slice(technologies, func(i, j int) bool {
		return technologies[i].Confidence > technologies[j].Confidence
	})

	return technologies
}

func analyzeProject(projectPath string, detector *TechnologyDetector) (*ProjectInfo, []FileInfo, error) {
	var files []FileInfo
	var totalSize int64
	totalFiles := 0

	err := filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored directories
		if d.IsDir() {
			if ignoredDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		totalFiles++
		relPath, _ := filepath.Rel(projectPath, path)

		// Skip ignored files
		ext := strings.ToLower(filepath.Ext(path))
		if ignoredExtensions[ext] {
			return nil
		}

		// Skip hidden files (except specific ones)
		if strings.HasPrefix(d.Name(), ".") && !isImportantDotFile(d.Name()) {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return err
		}

		totalSize += info.Size()

		// Skip very large files (>1MB)
		if info.Size() > 1024*1024 {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Skip binary files
		if isBinaryFile(content) {
			return nil
		}

		fileInfo := FileInfo{
			Path:     relPath,
			Size:     info.Size(),
			Content:  string(content),
			Language: detectLanguage(ext),
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// Detect technologies
	technologies := detector.DetectTechnologies(projectPath, files)

	// Determine primary language
	primaryLang := "Unknown"
	if len(technologies) > 0 {
		primaryLang = technologies[0].Name
	}

	projectInfo := &ProjectInfo{
		Path:           projectPath,
		Technologies:   technologies,
		TotalFiles:     totalFiles,
		ProcessedFiles: len(files),
		Size:           totalSize,
		Language:       primaryLang,
	}

	return projectInfo, files, nil
}

func isImportantDotFile(name string) bool {
	importantFiles := map[string]bool{
		".gitignore": true, ".dockerignore": true, ".env": true,
		".env.example": true, ".eslintrc": true, ".prettierrc": true,
		".babelrc": true, ".travis.yml": true, ".github": true,
	}
	return importantFiles[name] || importantFiles[strings.TrimSuffix(name, filepath.Ext(name))]
}

func isBinaryFile(content []byte) bool {
	if len(content) == 0 {
		return false
	}

	// Check for null bytes (common in binary files)
	for i, b := range content {
		if b == 0 {
			return true
		}
		// Only check first 512 bytes for performance
		if i > 512 {
			break
		}
	}
	return false
}

func detectLanguage(ext string) string {
	langMap := map[string]string{
		".js":   "JavaScript",
		".jsx":  "JavaScript",
		".ts":   "TypeScript",
		".tsx":  "TypeScript",
		".py":   "Python",
		".go":   "Go",
		".java": "Java",
		".cpp":  "C++",
		".cc":   "C++",
		".cxx":  "C++",
		".c":    "C",
		".h":    "C/C++",
		".hpp":  "C++",
		".rs":   "Rust",
		".php":  "PHP",
		".rb":   "Ruby",
		".css":  "CSS",
		".scss": "SCSS",
		".sass": "Sass",
		".less": "Less",
		".html": "HTML",
		".htm":  "HTML",
		".xml":  "XML",
		".json": "JSON",
		".yaml": "YAML",
		".yml":  "YAML",
		".md":   "Markdown",
		".sh":   "Shell",
		".bash": "Bash",
		".ps1":  "PowerShell",
		".sql":  "SQL",
	}

	if lang, exists := langMap[ext]; exists {
		return lang
	}
	return "Text"
}

func generateOutput(projectInfo *ProjectInfo, files []FileInfo) string {
	var output strings.Builder

	// Header
	output.WriteString("# ProjectDump Analysis\n\n")
	output.WriteString(fmt.Sprintf("**Generated on:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	output.WriteString(fmt.Sprintf("**Project Path:** %s\n\n", projectInfo.Path))

	// Summary
	output.WriteString("## Project Summary\n\n")
	output.WriteString(fmt.Sprintf("- **Primary Language:** %s\n", projectInfo.Language))
	output.WriteString(fmt.Sprintf("- **Total Files:** %d\n", projectInfo.TotalFiles))
	output.WriteString(fmt.Sprintf("- **Processed Files:** %d\n", projectInfo.ProcessedFiles))
	output.WriteString(fmt.Sprintf("- **Project Size:** %.2f KB\n\n", float64(projectInfo.Size)/1024))

	// Technologies
	if len(projectInfo.Technologies) > 0 {
		output.WriteString("## Detected Technologies\n\n")
		for _, tech := range projectInfo.Technologies {
			confidence := fmt.Sprintf("%.1f%%", tech.Confidence*100)
			output.WriteString(fmt.Sprintf("### %s (%s confidence)\n", tech.Name, confidence))
			output.WriteString(fmt.Sprintf("*%s*\n\n", tech.Description))
			if len(tech.Files) > 0 {
				output.WriteString("**Related files:**\n")
				for _, file := range tech.Files[:min(5, len(tech.Files))] {
					output.WriteString(fmt.Sprintf("- %s\n", file))
				}
				if len(tech.Files) > 5 {
					output.WriteString(fmt.Sprintf("- ... and %d more files\n", len(tech.Files)-5))
				}
				output.WriteString("\n")
			}
		}
	}

	// Directory Structure
	output.WriteString("## Directory Structure\n\n")
	output.WriteString("```\n")
	output.WriteString(generateDirectoryTree(files))
	output.WriteString("```\n\n")

	// File Contents
	output.WriteString("## Source Code\n\n")

	// Group files by directory
	filesByDir := make(map[string][]FileInfo)
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		if dir == "." {
			dir = "root"
		}
		filesByDir[dir] = append(filesByDir[dir], file)
	}

	// Sort directories
	var dirs []string
	for dir := range filesByDir {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	for _, dir := range dirs {
		if dir != "root" {
			output.WriteString(fmt.Sprintf("### %s/\n\n", dir))
		}

		// Sort files in directory
		dirFiles := filesByDir[dir]
		sort.Slice(dirFiles, func(i, j int) bool {
			return dirFiles[i].Path < dirFiles[j].Path
		})

		for _, file := range dirFiles {
			output.WriteString(fmt.Sprintf("#### %s\n", file.Path))
			output.WriteString(fmt.Sprintf("*Language: %s | Size: %d bytes*\n\n", file.Language, file.Size))

			// Determine code block language for syntax highlighting
			syntaxLang := getSyntaxLanguage(file.Language)

			output.WriteString(fmt.Sprintf("```%s\n", syntaxLang))
			output.WriteString(file.Content)
			if !strings.HasSuffix(file.Content, "\n") {
				output.WriteString("\n")
			}
			output.WriteString("```\n\n")
		}
	}

	return output.String()
}

func generateDirectoryTree(files []FileInfo) string {
	type TreeNode struct {
		Name     string
		Children map[string]*TreeNode
		IsFile   bool
	}

	root := &TreeNode{
		Name:     "",
		Children: make(map[string]*TreeNode),
		IsFile:   false,
	}

	// Build tree structure
	for _, file := range files {
		parts := strings.Split(file.Path, string(filepath.Separator))
		current := root

		for i, part := range parts {
			if part == "" {
				continue
			}

			if current.Children[part] == nil {
				current.Children[part] = &TreeNode{
					Name:     part,
					Children: make(map[string]*TreeNode),
					IsFile:   i == len(parts)-1,
				}
			}
			current = current.Children[part]
		}
	}

	// Generate tree string
	var result strings.Builder
	var printTree func(*TreeNode, string, bool)

	printTree = func(node *TreeNode, prefix string, isLast bool) {
		if node.Name != "" {
			connector := "├── "
			if isLast {
				connector = "└── "
			}
			result.WriteString(prefix + connector + node.Name + "\n")
		}

		// Sort children
		var children []string
		for name := range node.Children {
			children = append(children, name)
		}
		sort.Strings(children)

		for i, name := range children {
			child := node.Children[name]
			newPrefix := prefix
			if node.Name != "" {
				if isLast {
					newPrefix += "    "
				} else {
					newPrefix += "│   "
				}
			}
			printTree(child, newPrefix, i == len(children)-1)
		}
	}

	printTree(root, "", true)
	return result.String()
}

func getSyntaxLanguage(language string) string {
	syntaxMap := map[string]string{
		"JavaScript": "javascript",
		"TypeScript": "typescript",
		"Python":     "python",
		"Go":         "go",
		"Java":       "java",
		"C++":        "cpp",
		"C":          "c",
		"C/C++":      "cpp",
		"Rust":       "rust",
		"PHP":        "php",
		"Ruby":       "ruby",
		"CSS":        "css",
		"SCSS":       "scss",
		"Sass":       "sass",
		"Less":       "less",
		"HTML":       "html",
		"XML":        "xml",
		"JSON":       "json",
		"YAML":       "yaml",
		"Markdown":   "markdown",
		"Shell":      "bash",
		"Bash":       "bash",
		"PowerShell": "powershell",
		"SQL":        "sql",
	}

	if syntax, exists := syntaxMap[language]; exists {
		return syntax
	}
	return "text"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
