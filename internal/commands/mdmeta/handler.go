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

// Command returns the mdmeta command
func Command() *cli.Command {
	return &cli.Command{
		Name:    "mdmeta",
		Aliases: []string{"mm"},
		Usage:   "Update markdown file metadata based on frontmatter",
		Description: `This command scans markdown files and updates their file system
metadata (creation time and modification time) using values from the frontmatter.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "directory",
				Aliases: []string{"d"},
				Usage:   "Directory containing markdown files",
				Value:   "./",
			},
			&cli.StringFlag{
				Name:    "created",
				Aliases: []string{"c"},
				Usage:   "Frontmatter attribute for creation date",
				Value:   "date",
			},
			&cli.StringFlag{
				Name:    "modified",
				Aliases: []string{"m"},
				Usage:   "Frontmatter attribute for modification date",
				Value:   "updated",
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"r"},
				Usage:   "Process directories recursively",
				Value:   true,
			},
		},
		Action: handleMdMeta,
	}
}

// Stats holds the processing statistics
type Stats struct {
	Processed int
	Skipped   int
	Updated   int
	Failed    int
}

// handleMdMeta handles the mdmeta command
func handleMdMeta(ctx context.Context, cmd *cli.Command) error {
	dir := cmd.String("directory")
	createdAttr := cmd.String("created")
	modifiedAttr := cmd.String("modified")
	recursive := cmd.Bool("recursive")
	verbose := cmd.Root().Bool("verbose")

	fmt.Printf("Processing markdown files in: %s\n", dir)
	fmt.Printf("Using frontmatter attributes:\n")
	fmt.Printf("  - Creation date: %s\n", createdAttr)
	fmt.Printf("  - Modification date: %s\n", modifiedAttr)
	fmt.Printf("Recursive mode: %t\n\n", recursive)

	stats := Stats{}

	// Walk the directory
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories unless in recursive mode
		if d.IsDir() {
			if path != dir && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		// Process only markdown files
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			stats.Skipped++
			return nil
		}

		stats.Processed++

		// Process the file
		updated, err := processMarkdownFile(path, createdAttr, modifiedAttr, verbose)
		if err != nil {
			stats.Failed++
			fmt.Fprintf(os.Stderr, "❌ Error processing %s: %s\n", d.Name(), err)
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

	// Print summary
	fmt.Printf("\nSummary:\n")
	fmt.Printf("- Processed: %d markdown files\n", stats.Processed)
	fmt.Printf("- Updated:   %d files\n", stats.Updated)
	fmt.Printf("- Failed:    %d files\n", stats.Failed)
	fmt.Printf("- Skipped:   %d non-markdown files\n", stats.Skipped)

	return nil
}

// processMarkdownFile processes a single markdown file
func processMarkdownFile(filePath, createdAttr, modifiedAttr string, verbose bool) (bool, error) {
	// Open the file
	file, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	// Parse frontmatter
	var metadata map[string]interface{}
	_, err = frontmatter.Parse(strings.NewReader(string(file)), &metadata)
	if err != nil {
		if verbose {
			fmt.Printf("⚠️ No valid frontmatter found in: %s\n", filepath.Base(filePath))
		}
		return false, nil
	}

	// Check we have at least one date attribute
	if _, hasCreated := metadata[createdAttr]; !hasCreated {
		if _, hasModified := metadata[modifiedAttr]; !hasModified {
			if verbose {
				fmt.Printf("⚠️ No date attributes found in: %s\n", filepath.Base(filePath))
			}
			return false, nil
		}
	}

	// Extract dates
	var createdTime, modifiedTime time.Time
	var createdOk, modifiedOk bool

	if createdStr, ok := getStringValue(metadata, createdAttr); ok {
		if t, err := parseDate(createdStr); err == nil {
			createdTime = t
			createdOk = true
		} else if verbose {
			fmt.Printf("⚠️ Invalid %s format in %s: %s\n",
				createdAttr, filepath.Base(filePath), createdStr)
		}
	}

	if modifiedStr, ok := getStringValue(metadata, modifiedAttr); ok {
		if t, err := parseDate(modifiedStr); err == nil {
			modifiedTime = t
			modifiedOk = true
		} else if verbose {
			fmt.Printf("⚠️ Invalid %s format in %s: %s\n",
				modifiedAttr, filepath.Base(filePath), modifiedStr)
		}
	}

	// If neither date is valid, skip
	if !createdOk && !modifiedOk {
		if verbose {
			fmt.Printf("⚠️ No valid dates found in: %s\n", filepath.Base(filePath))
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

	// Update file times
	if err := os.Chtimes(filePath, accessTime, modifyTime); err != nil {
		return false, err
	}

	// Print success message
	fmt.Printf("✅ Updated metadata of '%s':\n", filepath.Base(filePath))
	if createdOk {
		fmt.Printf("   - %s: %s\n", createdAttr, createdTime.Format(time.RFC3339))
	}
	if modifiedOk {
		fmt.Printf("   - %s: %s\n", modifiedAttr, modifiedTime.Format(time.RFC3339))
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
