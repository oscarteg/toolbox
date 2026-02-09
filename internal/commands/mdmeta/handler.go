package mdmeta

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/urfave/cli/v3"
)

// Stats holds the processing statistics
type Stats struct {
	Processed int
	Skipped   int
	Updated   int
	Failed    int
}

// updateOptions holds the options for processing markdown files
type updateOptions struct {
	createdAttr  string
	modifiedAttr string
	verbose      bool
	dryRun       bool
}

// handleMdMeta handles the mdmeta update command
func handleMdMeta(ctx context.Context, cmd *cli.Command) error {
	dir := cmd.String("directory")
	recursive := cmd.Bool("recursive")
	dryRun := cmd.Bool("dry-run")

	opts := updateOptions{
		createdAttr:  cmd.String("created"),
		modifiedAttr: cmd.String("modified"),
		verbose:      cmd.Root().Bool("verbose"),
		dryRun:       dryRun,
	}

	if dryRun {
		fmt.Println("DRY RUN: No changes will be made")
		fmt.Println()
	}

	fmt.Printf("Processing markdown files in: %s\n", dir)
	fmt.Printf("Using frontmatter attributes:\n")
	fmt.Printf("  - Creation date: %s\n", opts.createdAttr)
	fmt.Printf("  - Modification date: %s\n", opts.modifiedAttr)
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

		updated, err := processMarkdownFile(path, opts)
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

// processMarkdownFile processes a single markdown file and updates its timestamps
func processMarkdownFile(filePath string, opts updateOptions) (bool, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	var metadata map[string]any
	_, err = frontmatter.Parse(strings.NewReader(string(file)), &metadata)
	if err != nil {
		if opts.verbose {
			fmt.Printf("No valid frontmatter found in: %s\n", filepath.Base(filePath))
		}
		return false, nil
	}

	// Check we have at least one date attribute
	_, hasCreated := metadata[opts.createdAttr]
	_, hasModified := metadata[opts.modifiedAttr]
	if !hasCreated && !hasModified {
		if opts.verbose {
			fmt.Printf("No date attributes found in: %s\n", filepath.Base(filePath))
		}
		return false, nil
	}

	var createdTime, modifiedTime time.Time
	var createdOk, modifiedOk bool

	if createdStr, ok := getStringValue(metadata, opts.createdAttr); ok {
		if t, err := parseDate(createdStr); err == nil {
			createdTime = t
			createdOk = true
		} else if opts.verbose {
			fmt.Printf("Invalid %s format in %s: %s\n",
				opts.createdAttr, filepath.Base(filePath), createdStr)
		}
	}

	if modifiedStr, ok := getStringValue(metadata, opts.modifiedAttr); ok {
		if t, err := parseDate(modifiedStr); err == nil {
			modifiedTime = t
			modifiedOk = true
		} else if opts.verbose {
			fmt.Printf("Invalid %s format in %s: %s\n",
				opts.modifiedAttr, filepath.Base(filePath), modifiedStr)
		}
	}

	if !createdOk && !modifiedOk {
		if opts.verbose {
			fmt.Printf("No valid dates found in: %s\n", filepath.Base(filePath))
		}
		return false, nil
	}

	// Use created time for both if modified is not available
	accessTime := modifiedTime
	if !modifiedOk {
		accessTime = createdTime
	}

	modifyTime := modifiedTime
	if !modifiedOk {
		modifyTime = createdTime
	}

	if opts.dryRun {
		fmt.Printf("Would update metadata of '%s':\n", filepath.Base(filePath))
		if createdOk {
			fmt.Printf("   - %s: %s\n", opts.createdAttr, createdTime.Format(time.RFC3339))
		}
		if modifiedOk {
			fmt.Printf("   - %s: %s\n", opts.modifiedAttr, modifiedTime.Format(time.RFC3339))
		}
		return true, nil
	}

	if err := os.Chtimes(filePath, accessTime, modifyTime); err != nil {
		return false, err
	}

	fmt.Printf("Updated metadata of '%s':\n", filepath.Base(filePath))
	if createdOk {
		fmt.Printf("   - %s: %s\n", opts.createdAttr, createdTime.Format(time.RFC3339))
	}
	if modifiedOk {
		fmt.Printf("   - %s: %s\n", opts.modifiedAttr, modifiedTime.Format(time.RFC3339))
	}

	return true, nil
}

// getStringValue safely extracts a string value from metadata
func getStringValue(metadata map[string]interface{}, key string) (string, bool) {
	value, exists := metadata[key]
	if !exists {
		return "", false
	}

	// Handle different types
	switch v := value.(type) {
	case string:
		return v, true
	case time.Time:
		return v.Format(time.RFC3339), true
	case fmt.Stringer:
		return v.String(), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// parseDate attempts to parse a date string in various formats
func parseDate(dateStr string) (time.Time, error) {
	// Try different date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		time.RFC3339,
		"January 2, 2006",
		"Jan 2, 2006",
		"2/1/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}
