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

func handleLinkWorktrees(ctx context.Context, cmd *cli.Command) error {
	linksDir := cmd.String("links-dir")

	// Check if links directory exists
	if _, err := os.Stat(linksDir); os.IsNotExist(err) {
		return fmt.Errorf("error: '%s' directory not found in current directory", linksDir)
	}

	// Get all worktree paths (excluding the main worktree)
	worktrees, err := getWorktrees()
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
	for _, specialDir := range specialDirs {
		specialDirPath := filepath.Join(linksDir, specialDir)
		if _, err := os.Stat(specialDirPath); err == nil {
			sourcePath, err := filepath.Abs(specialDirPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for %s: %w", specialDirPath, err)
			}

			fmt.Printf("Linking directory %s to all worktrees...\n", specialDir)

			for _, worktree := range worktrees {
				targetPath := filepath.Join(worktree, specialDir)

				// Remove existing directory/symlink if it exists
				if _, err := os.Lstat(targetPath); err == nil {
					if err := os.RemoveAll(targetPath); err != nil {
						return fmt.Errorf("failed to remove existing %s: %w", targetPath, err)
					}
				}

				// Create symlink to the entire directory
				if err := os.Symlink(sourcePath, targetPath); err != nil {
					return fmt.Errorf("failed to create symlink %s -> %s: %w", sourcePath, targetPath, err)
				}
				fmt.Printf("  → %s\n", targetPath)
			}
			fmt.Println()
		}
	}

	// Find all files in links directory recursively, excluding special directories
	linkFiles, err := findLinkFiles(linksDir, specialDirs)
	if err != nil {
		return fmt.Errorf("failed to find link files: %w", err)
	}

	if len(linkFiles) == 0 {
		fmt.Printf("No files found in '%s' directory (excluding special directories)\n", linksDir)
		return nil
	}

	// Process each file
	for _, file := range linkFiles {
		// Get relative path from links directory
		relPath, err := filepath.Rel(linksDir, file)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", file, err)
		}

		sourcePath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}

		fmt.Printf("Linking %s to all worktrees...\n", relPath)

		// Link to each worktree
		for _, worktree := range worktrees {
			targetPath := filepath.Join(worktree, relPath)
			targetDir := filepath.Dir(targetPath)

			// Create target directory if it doesn't exist
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}

			// Remove existing file/symlink if it exists
			if _, err := os.Lstat(targetPath); err == nil {
				if err := os.Remove(targetPath); err != nil {
					return fmt.Errorf("failed to remove existing %s: %w", targetPath, err)
				}
			}

			// Create symlink
			if err := os.Symlink(sourcePath, targetPath); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", sourcePath, targetPath, err)
			}
			fmt.Printf("  → %s\n", targetPath)
		}
		fmt.Println()
	}

	fmt.Println("Symlink operation completed")
	return nil
}

func getWorktrees() ([]string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute git worktree list: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var worktrees []string
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