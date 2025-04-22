package mdmeta

import (
	"github.com/urfave/cli/v3"
)

// NewCommand creates a new mdmeta command
func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "mdmeta",
		Aliases: []string{"mm"},
		Usage:   "Update markdown file metadata based on frontmatter",
		Description: `This command scans markdown files and updates their file system
metadata (creation time and modification time) using values from the frontmatter.`,
		Action: handleMdMeta,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "directory",

				Usage: "Directory containing markdown files",
				Value: "./",
			},
			&cli.StringFlag{
				Name:  "created",
				Usage: "Frontmatter attribute for creation date",
				Value: "date",
			},
			&cli.StringFlag{
				Name:  "modified",
				Usage: "Frontmatter attribute for modification date",
				Value: "updated",
			},
			&cli.BoolFlag{
				Name:  "recursive",
				Usage: "Process directories recursively",
			},
		},
	}
}
