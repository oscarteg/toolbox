package linkworktrees

import (
	"github.com/urfave/cli/v3"
)

// NewCommand creates a new linkworktrees command
func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "linkworktrees",
		Aliases: []string{"lw"},
		Usage:   "Symlink files from links folder to all git worktrees",
		Description: `This command symlinks files from a 'links' directory to all git worktrees.
Special directories like .claude are linked as entire directories, while other files
are linked individually with their directory structure preserved.

EXAMPLES:
  toolbox linkworktrees                   # Link files from ./links to all worktrees
  toolbox lw -d config                    # Link files from ./config directory
  toolbox linkworktrees --links-dir=dots  # Link files from ./dots directory
  toolbox lw --dry-run                    # Preview what would be linked

The command will:
1. Find all git worktrees in the current repository
2. Create symbolic links from the source directory to each worktree
3. Preserve directory structure for individual files
4. Link special directories (like .claude) as complete directories`,
		Action: handleLinkWorktrees,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "links-dir",
				Aliases: []string{"d"},
				Usage:   "Directory containing files to link",
				Value:   "links",
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "Show what would be done without making changes",
				Value:   false,
			},
		},
	}
}
