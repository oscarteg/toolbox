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
are linked individually with their directory structure preserved.`,
		Action: handleLinkWorktrees,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "links-dir",
				Aliases: []string{"d"},
				Usage:   "Directory containing files to link",
				Value:   "links",
			},
		},
	}
}