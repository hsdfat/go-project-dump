package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Report is the fully-analyzed set of projects ready for rendering.
type Report struct {
	Projects    []*ProjectInfo
	GeneratedAt string
}

func (r *Report) totalFiles() int {
	n := 0
	for _, p := range r.Projects {
		n += p.TotalFiles
	}
	return n
}

func (r *Report) processedFiles() int {
	n := 0
	for _, p := range r.Projects {
		n += p.ProcessedFiles
	}
	return n
}

func (r *Report) totalSize() int64 {
	var n int64
	for _, p := range r.Projects {
		n += p.Size
	}
	return n
}

func (r *Report) estimatedTokens() int {
	n := 0
	for _, p := range r.Projects {
		for _, f := range p.Files {
			n += estimateTokens(f.Content)
		}
	}
	return n
}

func (r *Report) primaryLanguage() string {
	best := combineTechnologies(r.Projects)
	if len(best) == 0 {
		return "Unknown"
	}
	return best[0].Name
}

// combineTechnologies merges per-project technologies, keeping the highest
// confidence and the union of related files for each named technology.
func combineTechnologies(projects []*ProjectInfo) []Technology {
	byName := map[string]*Technology{}
	for _, p := range projects {
		for _, t := range p.Technologies {
			existing, ok := byName[t.Name]
			if !ok {
				clone := t
				byName[t.Name] = &clone
				continue
			}
			if t.Confidence > existing.Confidence {
				existing.Confidence = t.Confidence
			}
			existing.Files = append(existing.Files, t.Files...)
		}
	}
	out := make([]Technology, 0, len(byName))
	for _, t := range byName {
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Confidence != out[j].Confidence {
			return out[i].Confidence > out[j].Confidence
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// renderReport dispatches to the requested output format.
func renderReport(r *Report, opts Options) string {
	switch strings.ToLower(opts.Format) {
	case "xml":
		return renderXML(r, opts)
	case "text", "txt", "plain":
		return renderText(r, opts)
	default:
		return renderMarkdown(r, opts)
	}
}

// ---------------------------------------------------------------------------
// Markdown
// ---------------------------------------------------------------------------

func renderMarkdown(r *Report, opts Options) string {
	var out strings.Builder
	multi := len(r.Projects) > 1

	if multi {
		out.WriteString("# go-project-dump Multi-Project Analysis\n\n")
	} else {
		out.WriteString("# go-project-dump Analysis\n\n")
	}
	out.WriteString(fmt.Sprintf("**Generated on:** %s\n\n", r.GeneratedAt))
	out.WriteString("**Analyzed Projects:**\n")
	for _, p := range r.Projects {
		out.WriteString(fmt.Sprintf("- %s\n", p.Path))
	}
	out.WriteString("\n")

	out.WriteString("## Project Summary\n\n")
	if multi {
		out.WriteString(fmt.Sprintf("- **Number of Projects:** %d\n", len(r.Projects)))
	}
	out.WriteString(fmt.Sprintf("- **Primary Language:** %s\n", r.primaryLanguage()))
	out.WriteString(fmt.Sprintf("- **Total Files:** %s\n", humanCount(r.totalFiles())))
	out.WriteString(fmt.Sprintf("- **Processed Files:** %s\n", humanCount(r.processedFiles())))
	out.WriteString(fmt.Sprintf("- **Processed Size:** %s\n", humanSize(r.totalSize())))
	out.WriteString(fmt.Sprintf("- **Estimated Tokens:** ~%s\n\n", humanCount(r.estimatedTokens())))

	if multi {
		out.WriteString("### Individual Project Details\n\n")
		for _, p := range r.Projects {
			out.WriteString(fmt.Sprintf("#### %s\n", p.Name))
			out.WriteString(fmt.Sprintf("- **Path:** %s\n", p.Path))
			out.WriteString(fmt.Sprintf("- **Primary Language:** %s\n", p.Language))
			out.WriteString(fmt.Sprintf("- **Files:** %s total, %s processed\n",
				humanCount(p.TotalFiles), humanCount(p.ProcessedFiles)))
			out.WriteString(fmt.Sprintf("- **Size:** %s\n", humanSize(p.Size)))
			out.WriteString(fmt.Sprintf("- **Top Technologies:** %s\n\n", topTechSummary(p.Technologies, 3)))
		}
	}

	if techs := combineTechnologies(r.Projects); len(techs) > 0 {
		out.WriteString("## Detected Technologies\n\n")
		for _, tech := range techs {
			out.WriteString(fmt.Sprintf("### %s (%.1f%% confidence)\n", tech.Name, tech.Confidence*100))
			out.WriteString(fmt.Sprintf("*%s*\n\n", tech.Description))
			if len(tech.Files) > 0 {
				out.WriteString("**Related files:**\n")
				for _, file := range tech.Files[:min(5, len(tech.Files))] {
					out.WriteString(fmt.Sprintf("- %s\n", file))
				}
				if len(tech.Files) > 5 {
					out.WriteString(fmt.Sprintf("- ... and %d more files\n", len(tech.Files)-5))
				}
				out.WriteString("\n")
			}
		}
	}

	if opts.StatsOnly {
		return out.String()
	}

	for _, p := range r.Projects {
		if multi {
			out.WriteString(fmt.Sprintf("## Project: %s\n\n", p.Name))
		}
		out.WriteString("### Directory Structure\n\n```\n")
		out.WriteString(generateDirectoryTree(p.Files))
		out.WriteString("```\n\n")

		out.WriteString("### Source Code\n\n")
		for _, dir := range sortedDirs(p.Files) {
			files := filesInDir(p.Files, dir)
			if dir != "root" {
				out.WriteString(fmt.Sprintf("#### %s/\n\n", dir))
			}
			for _, file := range files {
				out.WriteString(fmt.Sprintf("##### %s\n", file.Path))
				out.WriteString(fmt.Sprintf("*Language: %s | Size: %d bytes*\n\n", file.Language, file.Size))
				out.WriteString(fmt.Sprintf("```%s\n", getSyntaxLanguage(file.Language)))
				out.WriteString(file.Content)
				if !strings.HasSuffix(file.Content, "\n") {
					out.WriteString("\n")
				}
				out.WriteString("```\n\n")
			}
		}
	}

	return out.String()
}

// ---------------------------------------------------------------------------
// XML (LLM-optimized)
// ---------------------------------------------------------------------------

func renderXML(r *Report, opts Options) string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("<project_dump generated=\"%s\">\n", xmlAttr(r.GeneratedAt)))

	out.WriteString("  <summary>\n")
	out.WriteString(fmt.Sprintf("    <projects>%d</projects>\n", len(r.Projects)))
	out.WriteString(fmt.Sprintf("    <primary_language>%s</primary_language>\n", xmlText(r.primaryLanguage())))
	out.WriteString(fmt.Sprintf("    <total_files>%d</total_files>\n", r.totalFiles()))
	out.WriteString(fmt.Sprintf("    <processed_files>%d</processed_files>\n", r.processedFiles()))
	out.WriteString(fmt.Sprintf("    <size_bytes>%d</size_bytes>\n", r.totalSize()))
	out.WriteString(fmt.Sprintf("    <estimated_tokens>%d</estimated_tokens>\n", r.estimatedTokens()))
	out.WriteString("  </summary>\n")

	if techs := combineTechnologies(r.Projects); len(techs) > 0 {
		out.WriteString("  <technologies>\n")
		for _, t := range techs {
			out.WriteString(fmt.Sprintf("    <technology name=\"%s\" confidence=\"%.2f\">%s</technology>\n",
				xmlAttr(t.Name), t.Confidence, xmlText(t.Description)))
		}
		out.WriteString("  </technologies>\n")
	}

	if opts.StatsOnly {
		out.WriteString("</project_dump>\n")
		return out.String()
	}

	out.WriteString("  <files>\n")
	for _, p := range r.Projects {
		for _, f := range p.Files {
			out.WriteString(fmt.Sprintf("    <file path=\"%s\" project=\"%s\" language=\"%s\" size=\"%d\">\n",
				xmlAttr(f.Path), xmlAttr(f.Project), xmlAttr(f.Language), f.Size))
			out.WriteString(cdata(f.Content))
			out.WriteString("\n    </file>\n")
		}
	}
	out.WriteString("  </files>\n")
	out.WriteString("</project_dump>\n")
	return out.String()
}

// ---------------------------------------------------------------------------
// Plain text
// ---------------------------------------------------------------------------

func renderText(r *Report, opts Options) string {
	const bar = "================================================================"
	const sep = "----------------------------------------------------------------"
	var out strings.Builder

	out.WriteString(bar + "\n")
	out.WriteString("go-project-dump — generated " + r.GeneratedAt + "\n")
	out.WriteString(bar + "\n\n")
	out.WriteString("Projects:          " + fmt.Sprintf("%d\n", len(r.Projects)))
	for _, p := range r.Projects {
		out.WriteString("  - " + p.Path + "\n")
	}
	out.WriteString("Primary language:  " + r.primaryLanguage() + "\n")
	out.WriteString("Total files:       " + humanCount(r.totalFiles()) + "\n")
	out.WriteString("Processed files:   " + humanCount(r.processedFiles()) + "\n")
	out.WriteString("Processed size:    " + humanSize(r.totalSize()) + "\n")
	out.WriteString("Estimated tokens:  ~" + humanCount(r.estimatedTokens()) + "\n\n")

	if opts.StatsOnly {
		return out.String()
	}

	for _, p := range r.Projects {
		out.WriteString(bar + "\n")
		out.WriteString("Directory structure: " + p.Name + "\n")
		out.WriteString(bar + "\n")
		out.WriteString(generateDirectoryTree(p.Files))
		out.WriteString("\n")
	}

	for _, p := range r.Projects {
		for _, f := range p.Files {
			out.WriteString(bar + "\n")
			out.WriteString("FILE: " + f.Path + "  (" + f.Project + ")\n")
			out.WriteString(sep + "\n")
			out.WriteString(f.Content)
			if !strings.HasSuffix(f.Content, "\n") {
				out.WriteString("\n")
			}
			out.WriteString("\n")
		}
	}
	return out.String()
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

func topTechSummary(techs []Technology, n int) string {
	if len(techs) == 0 {
		return "n/a"
	}
	parts := make([]string, 0, n)
	for i, t := range techs {
		if i >= n {
			break
		}
		parts = append(parts, fmt.Sprintf("%s (%.1f%%)", t.Name, t.Confidence*100))
	}
	return strings.Join(parts, ", ")
}

func sortedDirs(files []FileInfo) []string {
	seen := map[string]bool{}
	var dirs []string
	for _, f := range files {
		dir := filepath.Dir(f.Path)
		if dir == "." {
			dir = "root"
		}
		if !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs)
	return dirs
}

func filesInDir(files []FileInfo, dir string) []FileInfo {
	var out []FileInfo
	for _, f := range files {
		d := filepath.Dir(f.Path)
		if d == "." {
			d = "root"
		}
		if d == dir {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func generateDirectoryTree(files []FileInfo) string {
	type treeNode struct {
		name     string
		children map[string]*treeNode
	}
	root := &treeNode{children: map[string]*treeNode{}}

	for _, file := range files {
		parts := strings.Split(file.Path, "/")
		current := root
		for _, part := range parts {
			if part == "" {
				continue
			}
			if current.children[part] == nil {
				current.children[part] = &treeNode{name: part, children: map[string]*treeNode{}}
			}
			current = current.children[part]
		}
	}

	var result strings.Builder
	var printTree func(node *treeNode, prefix string, isLast bool)
	printTree = func(node *treeNode, prefix string, isLast bool) {
		if node.name != "" {
			connector := "├── "
			if isLast {
				connector = "└── "
			}
			result.WriteString(prefix + connector + node.name + "\n")
		}

		names := make([]string, 0, len(node.children))
		for name := range node.children {
			names = append(names, name)
		}
		sort.Strings(names)

		for i, name := range names {
			newPrefix := prefix
			if node.name != "" {
				if isLast {
					newPrefix += "    "
				} else {
					newPrefix += "│   "
				}
			}
			printTree(node.children[name], newPrefix, i == len(names)-1)
		}
	}
	printTree(root, "", true)
	return result.String()
}

func xmlAttr(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&apos;")
	return replacer.Replace(s)
}

func xmlText(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(s)
}

// cdata wraps content in a CDATA section, splitting any embedded "]]>" so the
// section stays well-formed.
func cdata(s string) string {
	s = strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>")
	return "<![CDATA[" + s + "]]>"
}
