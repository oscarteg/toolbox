package mdmeta

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/urfave/cli/v3"
)

// handleDelete is the CLI handler for the delete subcommand
func handleDelete(ctx context.Context, cmd *cli.Command) error {
	dir := cmd.String("directory")
	recursive := cmd.Bool("recursive")
	verbose := cmd.Root().Bool("verbose")
	dryRun := cmd.Bool("dry-run")

	if dryRun {
		fmt.Println("DRY RUN: No changes will be made")
		fmt.Println()
	}

	fmt.Printf("Removing frontmatter from markdown files in: %s\n", dir)
	fmt.Printf("Recursive mode: %t\n\n", recursive)

	stats := Stats{}

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() {
			if path != dir && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			stats.Skipped++
			return nil
		}

		stats.Processed++

		updated, err := deleteFrontmatter(path, verbose, dryRun)
		if err != nil {
			stats.Failed++
			fmt.Fprintf(os.Stderr, "Error processing %s: %s\n", d.Name(), err)
			return nil
		}

		if updated {
			stats.Updated++
		}

		return nil
	}

	if err := filepath.WalkDir(dir, walkFn); err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("- Processed: %d markdown files\n", stats.Processed)
	fmt.Printf("- Updated:   %d files\n", stats.Updated)
	fmt.Printf("- Failed:    %d files\n", stats.Failed)
	fmt.Printf("- Skipped:   %d non-markdown files\n", stats.Skipped)

	return nil
}

// deleteFrontmatter removes frontmatter from a markdown file
func deleteFrontmatter(filePath string, verbose, dryRun bool) (bool, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	content := string(file)

	matter := map[string]any{}
	rest, err := frontmatter.Parse(strings.NewReader(content), &matter)
	if err != nil {
		if verbose {
			fmt.Printf("No valid frontmatter found in: %s\n", filepath.Base(filePath))
		}
		return false, nil
	}

	// If no frontmatter was found (empty matter map), skip
	if len(matter) == 0 {
		if verbose {
			fmt.Printf("No frontmatter to remove in: %s\n", filepath.Base(filePath))
		}
		return false, nil
	}

	if dryRun {
		fmt.Printf("Would remove frontmatter from '%s'\n", filepath.Base(filePath))
		return true, nil
	}

	if err := os.WriteFile(filePath, rest, 0644); err != nil {
		return false, err
	}

	fmt.Printf("Removed frontmatter from '%s'\n", filepath.Base(filePath))
	return true, nil
}
