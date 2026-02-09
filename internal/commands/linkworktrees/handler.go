package linkworktrees

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

// linkOptions holds configuration for the link operation
type linkOptions struct {
	linksDir string
	dryRun   bool
	verbose  bool
}

func handleLinkWorktrees(ctx context.Context, cmd *cli.Command) error {
	opts := linkOptions{
		linksDir: cmd.String("links-dir"),
		dryRun:   cmd.Bool("dry-run"),
		verbose:  cmd.Root().Bool("verbose"),
	}

	if opts.dryRun {
		fmt.Println("DRY RUN: No changes will be made")
		fmt.Println()
	}

	// Check if links directory exists
	if _, err := os.Stat(opts.linksDir); os.IsNotExist(err) {
		return fmt.Errorf("error: '%s' directory not found in current directory", opts.linksDir)
	}

	// Get all worktree paths (excluding the main worktree)
	worktrees, err := getWorktrees(ctx)
	if err != nil {
		return fmt.Errorf("failed to get worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		fmt.Println("No additional worktrees found")
		return nil
	}

	fmt.Println("Found worktrees:")
	for _, worktree := range worktrees {
		fmt.Printf("  %s\n", worktree)
	}
	fmt.Println()

	// Handle special directory exceptions first (.claude)
	specialDirs := []string{".claude"}
	if err := linkSpecialDirs(ctx, opts, worktrees, specialDirs); err != nil {
		return err
	}

	// Find all files in links directory recursively, excluding special directories
	linkFiles, err := findLinkFiles(opts.linksDir, specialDirs)
	if err != nil {
		return fmt.Errorf("failed to find link files: %w", err)
	}

	if len(linkFiles) == 0 {
		fmt.Printf("No files found in '%s' directory (excluding special directories)\n", opts.linksDir)
		return nil
	}

	// Process each file
	if err := linkFiles_(ctx, opts, worktrees, linkFiles); err != nil {
		return err
	}

	fmt.Println("Symlink operation completed")
	return nil
}

// linkSpecialDirs links special directories (like .claude) as complete directories
func linkSpecialDirs(ctx context.Context, opts linkOptions, worktrees, specialDirs []string) error {
	for _, specialDir := range specialDirs {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		specialDirPath := filepath.Join(opts.linksDir, specialDir)
		if _, err := os.Stat(specialDirPath); err != nil {
			continue // Directory doesn't exist, skip
		}

		sourcePath, err := filepath.Abs(specialDirPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", specialDirPath, err)
		}

		fmt.Printf("Linking directory %s to all worktrees...\n", specialDir)

		for _, worktree := range worktrees {
			targetPath := filepath.Join(worktree, specialDir)

			if opts.dryRun {
				fmt.Printf("  Would link: %s -> %s\n", sourcePath, targetPath)
				continue
			}

			// Remove existing directory/symlink if it exists
			if _, err := os.Lstat(targetPath); err == nil {
				if err := os.RemoveAll(targetPath); err != nil {
					return fmt.Errorf("failed to remove existing %s: %w", targetPath, err)
				}
			}

			if err := os.Symlink(sourcePath, targetPath); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", sourcePath, targetPath, err)
			}
			fmt.Printf("  -> %s\n", targetPath)
		}
		fmt.Println()
	}
	return nil
}

// linkFiles_ links individual files to worktrees
func linkFiles_(ctx context.Context, opts linkOptions, worktrees, files []string) error {
	for _, file := range files {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(opts.linksDir, file)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", file, err)
		}

		sourcePath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}

		fmt.Printf("Linking %s to all worktrees...\n", relPath)

		for _, worktree := range worktrees {
			targetPath := filepath.Join(worktree, relPath)
			targetDir := filepath.Dir(targetPath)

			if opts.dryRun {
				fmt.Printf("  Would link: %s -> %s\n", sourcePath, targetPath)
				continue
			}

			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}

			// Remove existing file/symlink if it exists
			if _, err := os.Lstat(targetPath); err == nil {
				if err := os.Remove(targetPath); err != nil {
					return fmt.Errorf("failed to remove existing %s: %w", targetPath, err)
				}
			}

			if err := os.Symlink(sourcePath, targetPath); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", sourcePath, targetPath, err)
			}
			fmt.Printf("  -> %s\n", targetPath)
		}
		fmt.Println()
	}
	return nil
}

func getWorktrees(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute git worktree list: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	worktrees := make([]string, 0, len(lines)/4) // Pre-allocate, ~4 lines per worktree
	first := true

	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			if first {
				// Skip the first worktree (main worktree)
				first = false
				continue
			}
			worktree := strings.TrimPrefix(line, "worktree ")
			if worktree != "" {
				worktrees = append(worktrees, worktree)
			}
		}
	}

	return worktrees, nil
}

func findLinkFiles(linksDir string, excludeDirs []string) ([]string, error) {
	var files []string

	err := filepath.Walk(linksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip hidden files (files starting with .)
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Check if file is in an excluded directory
		relPath, err := filepath.Rel(linksDir, path)
		if err != nil {
			return err
		}

		for _, excludeDir := range excludeDirs {
			if strings.HasPrefix(relPath, excludeDir+string(filepath.Separator)) {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}
