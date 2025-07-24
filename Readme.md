# go-project-dump 🚀

A powerful CLI tool that analyzes project technologies, filters out non-essential files, and compiles source code and directory structure into a single readable file. Perfect for code reviews, documentation, sharing projects, or AI analysis.

## ✨ Features

- **🔍 Smart Technology Detection**: Automatically identifies 15+ technologies including JavaScript, TypeScript, Python, Go, Java, React, Docker, and more
- **🗂️ Multi-Project Support**: Analyze single projects or multiple folders simultaneously
- **📁 Intelligent Filtering**: Excludes build artifacts, node_modules, binary files, and other non-essential content
- **📊 Comprehensive Analysis**: Generates detailed reports with technology confidence scores
- **🌳 Directory Visualization**: Creates beautiful tree structures of your project layout
- **📝 Markdown Output**: Produces clean, readable markdown with syntax highlighting
- **⚡ Fast & Efficient**: Written in Go for optimal performance

## 🚀 Quick Start

### Installation

Install directly using Go:

```bash
go install github.com/hsdfat/go-project-dump@latest
```

Or clone and build:

```bash
git clone https://github.com/hsdfat/go-project-dump.git
cd go-project-dump
go build -o go-project-dump main.go
```

### Basic Usage

```bash
# Analyze a single project
go-project-dump /path/to/your/project

# Save output to file
go-project-dump /path/to/your/project -o analysis.md

# Analyze multiple projects
go-project-dump /project1 /project2 /project3 -o combined-analysis.md

# Analyze microservices architecture
go-project-dump /app/frontend /app/backend /app/api -o microservices-dump.md
```

## 📖 Usage Examples

### Single Project Analysis
```bash
go-project-dump ~/my-react-app -o react-analysis.md
```

### Multiple Project Analysis
```bash
# Analyze frontend and backend together
go-project-dump ~/myapp/client ~/myapp/server -o fullstack-analysis.md

# Compare multiple similar projects
go-project-dump ~/project-v1 ~/project-v2 ~/project-v3 -o version-comparison.md
```

### Microservices Architecture
```bash
go-project-dump \
  ~/microservices/user-service \
  ~/microservices/payment-service \
  ~/microservices/notification-service \
  -o microservices-overview.md
```

## 🔧 Supported Technologies

go-project-dump can detect and analyze:

### Programming Languages
- **JavaScript** (Node.js, browser-based)
- **TypeScript** (TS/TSX files, tsconfig.json)
- **Python** (requirements.txt, setup.py, pipenv)
- **Go** (go.mod, .go files)
- **Java** (Maven, Gradle projects)
- **C/C++** (CMake, Makefiles)
- **Rust** (Cargo.toml)
- **PHP** (Composer projects)
- **Ruby** (Gemfile projects)

### Frontend Technologies
- **React** (JSX/TSX detection)
- **HTML/CSS** (including SCSS, Sass, Less)

### Infrastructure & DevOps
- **Docker** (Dockerfile, docker-compose)
- **Kubernetes** (YAML manifests)

### Configuration & Data
- **JSON, YAML, XML**
- **Markdown**
- **Shell scripts**

## 📊 Output Format

go-project-dump generates a comprehensive markdown report containing:

### 1. Project Summary
- Combined statistics (file counts, sizes, languages)
- Individual project breakdowns (for multi-project analysis)
- Primary language detection

### 2. Technology Analysis
- Detected technologies with confidence scores
- Related files for each technology
- Technology descriptions and context

### 3. Directory Structure
- Visual tree representation
- Clean, readable format
- Multi-project organization

### 4. Source Code
- Complete source code with syntax highlighting
- Organized by project and directory
- Language detection for proper highlighting

## 🎯 Use Cases

### Code Reviews
```bash
go-project-dump /path/to/feature-branch -o feature-review.md
```

### Documentation
```bash
go-project-dump /project -o project-documentation.md
```

### AI Analysis
```bash
# Perfect for feeding to AI tools for code analysis
go-project-dump /complex-project -o ai-analysis-input.md
```

### Project Comparison
```bash
go-project-dump /old-version /new-version -o migration-analysis.md
```

### Team Onboarding
```bash
go-project-dump /team-project -o onboarding-guide.md
```

## ⚙️ Command Line Options

```
Usage: go-project-dump <project-paths...> [-o output-file]

Arguments:
  project-paths    One or more paths to project directories (space-separated)

Options:
  -o output-file   Optional output file (default: stdout)

Examples:
  go-project-dump /path/to/project
  go-project-dump /path/to/project1 /path/to/project2
  go-project-dump /path/to/project -o output.md
  go-project-dump /frontend /backend -o combined.md
```

## 🚫 Filtering Logic

go-project-dump intelligently excludes:

### Directories
- `node_modules`, `vendor`, `__pycache__`
- `.git`, `.svn`, `.hg`
- `build`, `dist`, `target`, `bin`, `obj`
- `.idea`, `.vscode`
- `logs`, `tmp`, `temp`

### File Types
- **Binary files**: `.exe`, `.dll`, `.so`, `.jar`
- **Media files**: `.jpg`, `.png`, `.mp4`, `.mp3`
- **Archives**: `.zip`, `.tar`, `.gz`
- **Compiled files**: `.class`, `.o`, `.pyc`

### File Size
- Files larger than 1MB are automatically excluded
- Binary file detection prevents including non-text content

## 🛠️ Development

### Building from Source

```bash
git clone https://github.com/hsdfat/go-project-dump.git
cd go-project-dump
go mod tidy
go build -o go-project-dump main.go
```

### Running Tests

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 Example Output

Here's what a typical go-project-dump analysis looks like:

```markdown
# go-project-dump Multi-Project Analysis

**Generated on:** 2025-07-19 14:30:25
**Analyzed Projects:**
- /home/user/frontend
- /home/user/backend

## Combined Project Summary

- **Number of Projects:** 2
- **Primary Language:** TypeScript
- **Total Files:** 847
- **Processed Files:** 156
- **Combined Size:** 2.4 MB

### Individual Project Details

#### frontend
- **Path:** /home/user/frontend
- **Primary Language:** TypeScript
- **Files:** 423 total, 89 processed
- **Size:** 1.2 MB
- **Top Technologies:** TypeScript (95.2%), React (87.3%), CSS (45.1%)

#### backend
- **Path:** /home/user/backend  
- **Primary Language:** Go
- **Files:** 424 total, 67 processed
- **Size:** 1.2 MB
- **Top Technologies:** Go (92.1%), Docker (78.4%), YAML (34.2%)

## Detected Technologies (Combined)

### TypeScript (95.2% confidence)
*TypeScript - JavaScript with static typing*

**Related files:**
- frontend/src/components/App.tsx
- frontend/src/utils/api.ts
- frontend/tsconfig.json
- ... and 45 more files
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- Inspired by the need for better code analysis tools
- Thanks to all contributors who help improve this tool
---

**Made with ❤️ by developers, for developers**