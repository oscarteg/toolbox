package mdmeta

import (
	"github.com/urfave/cli/v3"
)

// Common flags shared across mdmeta subcommands
var commonFlags = []cli.Flag{
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
	&cli.BoolFlag{
		Name:    "dry-run",
		Aliases: []string{"n"},
		Usage:   "Show what would be done without making changes",
		Value:   false,
	},
}

// NewCommand creates a new mdmeta command
func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "mdmeta",
		Aliases: []string{"mm"},
		Usage:   "Update markdown file metadata based on frontmatter",
		Description: `This command scans markdown files and updates their file system
metadata (creation time and modification time) using values from the frontmatter.

EXAMPLES:
  toolbox mdmeta update                           # Update metadata in current directory
  toolbox mm update -d ./posts                    # Update metadata in ./posts directory
  toolbox mdmeta update -c created -m modified    # Use custom frontmatter fields
  toolbox mm update -d ./content -r               # Process ./content recursively
  toolbox mm update --dry-run                     # Preview changes without applying

The command will:
1. Scan for markdown files in the specified directory
2. Parse frontmatter from each file
3. Update file system timestamps based on frontmatter values
4. Support custom field names for created/modified dates`,

		Commands: []*cli.Command{
			{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "Update file metadata from frontmatter dates",
				Flags:   commonFlags,
				Action:  handleMdMeta,
			},
			{
				Name:    "delete",
				Aliases: []string{"del"},
				Usage:   "Remove frontmatter from markdown files",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "directory",
						Aliases: []string{"d"},
						Usage:   "Directory containing markdown files",
						Value:   "./",
					},
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "Process directories recursively",
						Value:   true,
					},
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "Show what would be done without making changes",
						Value:   false,
					},
				},
				Action: handleDelete,
			},
		},
	}
}
