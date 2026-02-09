# Toolbox

A personal collection of utility tools written in Go for common automation tasks.

## Features

- **linkworktrees**: Symlink files from a source directory to all git worktrees
- **mdmeta**: Update markdown file metadata based on frontmatter values

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/oscarteg/toolbox.git
cd toolbox

# Build the toolbox
go build -o toolbox ./cmd/toolbox

# Move to a location in your PATH (optional)
sudo mv toolbox /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/oscarteg/toolbox/cmd/toolbox@latest
```

## Usage

### Getting Help

```bash
# Show all available commands
toolbox --help

# Show help for a specific command
toolbox linkworktrees --help
toolbox mdmeta --help
```

## Commands

### linkworktrees (alias: lw)

Symlinks files from a source directory to all git worktrees in the current repository.

```bash
# Basic usage - link files from ./links directory
toolbox linkworktrees

# Use a different source directory
toolbox linkworktrees --links-dir config
toolbox lw -d config

# Examples
toolbox linkworktrees --links-dir=dots
```

**How it works:**
1. Finds all git worktrees in the current repository
2. Creates symbolic links from the source directory to each worktree
3. Preserves directory structure for individual files
4. Links special directories (like `.claude`) as complete directories

**Use cases:**
- Share configuration files across multiple git worktrees
- Maintain consistent development environment setup
- Sync IDE settings and configurations

### mdmeta (alias: mm)

Updates markdown file system metadata (creation and modification times) based on frontmatter values.

```bash
# Update metadata in current directory
toolbox mdmeta update

# Process a specific directory
toolbox mdmeta update --directory ./posts
toolbox mm update -d ./posts

# Use custom frontmatter field names
toolbox mdmeta update --created created_at --modified updated_at
toolbox mm update -c created_at -m updated_at

# Process directories recursively
toolbox mdmeta update --directory ./content --recursive
```

**How it works:**
1. Scans for markdown files in the specified directory
2. Parses YAML frontmatter from each file
3. Updates file system timestamps based on frontmatter values
4. Supports custom field names for created/modified dates

**Default frontmatter fields:**
- Creation time: `date`
- Modification time: `updated`

**Use cases:**
- Sync file timestamps with blog post dates
- Maintain accurate file metadata for content management
- Organize files by their actual creation/update dates

## Examples

### Setting up shared configurations across worktrees

```bash
# Create a links directory with shared files
mkdir links
cp .vimrc links/
cp .gitconfig links/
mkdir -p links/.vscode
cp -r .vscode/settings.json links/.vscode/

# Link to all worktrees
toolbox linkworktrees
```

### Updating blog post metadata

```bash
# For a blog with posts in ./content/posts/
toolbox mdmeta update -d ./content/posts -r

# Using custom frontmatter fields
toolbox mdmeta update -d ./content -c publishDate -m lastmod -r
```

## Global Flags

- `--verbose, -v`: Enable verbose output
- `--help, -h`: Show help
- `--version`: Show version

## Development

```bash
# Run tests
go test ./...

# Build
go build -o toolbox ./cmd/toolbox

# Run locally
go run ./cmd/toolbox <command>
```

## License

MIT License - see LICENSE file for details.
