package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Options controls a single run of the analyzer/renderer.
type Options struct {
	Paths        []string
	Output       string // output file; empty means stdout
	Format       string // markdown | xml | text
	MaxFileSize  int64  // skip files larger than this many bytes
	UseGitignore bool   // honour each project's .gitignore
	Excludes     []string
	Includes     []string
	StatsOnly    bool // print only the summary, no source
}

// FileInfo represents a single included file.
type FileInfo struct {
	Path     string // path relative to the project root (slash-separated)
	Project  string // owning project name
	Size     int64
	Content  string
	Language string
}

// ProjectInfo holds the analysis of one project path.
type ProjectInfo struct {
	Name           string
	Path           string
	Technologies   []Technology
	TotalFiles     int
	ProcessedFiles int
	Size           int64 // total bytes of included content
	Language       string
	Files          []FileInfo
}

// Common file extensions to ignore.
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

// Directories to ignore even without a .gitignore.
var ignoredDirs = map[string]bool{
	"node_modules": true, ".git": true, ".svn": true, ".hg": true,
	"vendor": true, "__pycache__": true, ".idea": true, ".vscode": true,
	"build": true, "dist": true, "target": true, "bin": true, "obj": true,
	".next": true, ".nuxt": true, "coverage": true, ".nyc_output": true,
	"logs": true, "tmp": true, "temp": true,
}

// analyzeProject walks a single project path and returns its analysis.
func analyzeProject(projectPath string, detector *TechnologyDetector, opts Options) (*ProjectInfo, error) {
	rootInfo, err := os.Stat(projectPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access %q: %w", projectPath, err)
	}

	name := filepath.Base(strings.TrimRight(projectPath, string(filepath.Separator)))

	// Gather ignore/keep matchers once per project.
	ignore := &Gitignore{}
	if opts.UseGitignore && rootInfo.IsDir() {
		ignore, err = LoadGitignore(filepath.Join(projectPath, ".gitignore"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: reading .gitignore: %v\n", err)
			ignore = &Gitignore{}
		}
	}
	extraExclude := NewGitignore(opts.Excludes)
	var include *Gitignore
	if len(opts.Includes) > 0 {
		include = NewGitignore(opts.Includes)
	}

	info := &ProjectInfo{Name: name, Path: projectPath}

	// A single file path is treated as a one-file project.
	if !rootInfo.IsDir() {
		info.TotalFiles = 1
		if f, ok := readFile(projectPath, name, rootInfo, opts); ok {
			info.Files = append(info.Files, f)
			info.ProcessedFiles = 1
			info.Size = f.Size
		}
		info.Technologies = detector.DetectTechnologies(info.Files)
		info.Language = primaryLanguage(info.Technologies)
		return info, nil
	}

	walkErr := filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip unreadable entries instead of aborting the whole walk.
			fmt.Fprintf(os.Stderr, "warning: skipping %q: %v\n", path, err)
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, relErr := filepath.Rel(projectPath, path)
		if relErr != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		if d.IsDir() {
			if path == projectPath {
				return nil
			}
			if ignoredDirs[d.Name()] ||
				ignore.Match(relPath, true) ||
				extraExclude.Match(relPath, true) {
				return filepath.SkipDir
			}
			return nil
		}

		info.TotalFiles++

		ext := strings.ToLower(filepath.Ext(path))
		if ignoredExtensions[ext] {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") && !isImportantDotFile(d.Name()) {
			return nil
		}
		if ignore.Match(relPath, false) || extraExclude.Match(relPath, false) {
			return nil
		}
		if include != nil && !include.Match(relPath, false) {
			return nil
		}

		fi, ferr := d.Info()
		if ferr != nil {
			return nil
		}
		if opts.MaxFileSize > 0 && fi.Size() > opts.MaxFileSize {
			return nil
		}

		f, ok := readFileEntry(path, relPath, name, fi, opts)
		if !ok {
			return nil
		}
		info.Files = append(info.Files, f)
		info.ProcessedFiles++
		info.Size += f.Size
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	info.Technologies = detector.DetectTechnologies(info.Files)
	info.Language = primaryLanguage(info.Technologies)
	return info, nil
}

// readFile reads a standalone file path (single-file project).
func readFile(path, project string, fi os.FileInfo, opts Options) (FileInfo, bool) {
	return readFileEntry(path, filepath.Base(path), project, fi, opts)
}

// readFileEntry reads and classifies one file, returning ok=false if it should
// be skipped (unreadable, oversized, or binary).
func readFileEntry(path, relPath, project string, fi os.FileInfo, opts Options) (FileInfo, bool) {
	if opts.MaxFileSize > 0 && fi.Size() > opts.MaxFileSize {
		return FileInfo{}, false
	}
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: skipping %q: %v\n", path, err)
		return FileInfo{}, false
	}
	if isBinaryFile(content) {
		return FileInfo{}, false
	}
	return FileInfo{
		Path:     filepath.ToSlash(relPath),
		Project:  project,
		Size:     fi.Size(),
		Content:  string(content),
		Language: detectLanguage(strings.ToLower(filepath.Ext(path))),
	}, true
}

func primaryLanguage(techs []Technology) string {
	if len(techs) == 0 {
		return "Unknown"
	}
	return techs[0].Name
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
	limit := len(content)
	if limit > 512 {
		limit = 512
	}
	for _, b := range content[:limit] {
		if b == 0 { // null byte: strong signal of binary content
			return true
		}
	}
	return false
}

func detectLanguage(ext string) string {
	langMap := map[string]string{
		".js": "JavaScript", ".jsx": "JavaScript", ".mjs": "JavaScript",
		".ts": "TypeScript", ".tsx": "TypeScript",
		".py": "Python", ".pyw": "Python",
		".go":   "Go",
		".java": "Java",
		".cpp":  "C++", ".cc": "C++", ".cxx": "C++", ".hpp": "C++",
		".c": "C", ".h": "C/C++",
		".rs":  "Rust",
		".php": "PHP",
		".rb":  "Ruby",
		".css": "CSS", ".scss": "SCSS", ".sass": "Sass", ".less": "Less",
		".html": "HTML", ".htm": "HTML",
		".xml": "XML", ".json": "JSON",
		".yaml": "YAML", ".yml": "YAML",
		".md": "Markdown", ".sh": "Shell", ".bash": "Bash",
		".ps1": "PowerShell", ".sql": "SQL",
	}
	if lang, exists := langMap[ext]; exists {
		return lang
	}
	return "Text"
}

func getSyntaxLanguage(language string) string {
	syntaxMap := map[string]string{
		"JavaScript": "javascript", "TypeScript": "typescript",
		"Python": "python", "Go": "go", "Java": "java",
		"C++": "cpp", "C": "c", "C/C++": "cpp", "Rust": "rust",
		"PHP": "php", "Ruby": "ruby", "CSS": "css", "SCSS": "scss",
		"Sass": "sass", "Less": "less", "HTML": "html", "XML": "xml",
		"JSON": "json", "YAML": "yaml", "Markdown": "markdown",
		"Shell": "bash", "Bash": "bash", "PowerShell": "powershell", "SQL": "sql",
	}
	if syntax, exists := syntaxMap[language]; exists {
		return syntax
	}
	return "text"
}
