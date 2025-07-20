package mdmeta

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
)

// deleteHandler removes frontmatter from a markdown file
func deleteHandler(filePath string, verbose bool) (bool, error) {
	// Read the file
	file, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	content := string(file)

	// Check if the file has frontmatter
	matter := map[string]interface{}{}
	rest, err := frontmatter.Parse(strings.NewReader(content), &matter)
	if err != nil {
		if verbose {
			fmt.Printf("⚠️ No valid frontmatter found in: %s\n", filepath.Base(filePath))
		}
		return false, nil
	}

	// Get content without frontmatter
	contentWithoutFrontmatter := rest

	// Write the content back to the file
	err = os.WriteFile(filePath, contentWithoutFrontmatter, 0644)
	if err != nil {
		return false, err
	}

	fmt.Printf("✅ Removed frontmatter from '%s'\n", filepath.Base(filePath))
	return true, nil
}
