package commands

import (
	"net/mail"

	"github.com/oscarteg/toolbox/internal/commands/linkworktrees"
	"github.com/oscarteg/toolbox/internal/commands/mdmeta"
	"github.com/urfave/cli/v3"
)

func RootCommand() *cli.Command {
	return &cli.Command{
		Name:  "toolbox",
		Usage: "Personal toolkit for automation tasks",
		Description: `A collection of utility tools for common automation tasks.

EXAMPLES:
  toolbox linkworktrees                   # Link files from ./links to all git worktrees
  toolbox linkworktrees -d config         # Link files from ./config directory
  toolbox mdmeta update                   # Update markdown metadata from frontmatter
  toolbox mdmeta update -d ./posts -r     # Process ./posts recursively

Run 'toolbox <command> --help' for more information on a specific command.`,
		Version: "0.1.0",
		Authors: []any{
			mail.Address{Name: "Oscar te Giffel", Address: "oscar@tegiffel.com"},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "Enable verbose output",
				Aliases: []string{"v"},
			},
		},
		Commands: []*cli.Command{
			linkworktrees.NewCommand(),
			mdmeta.NewCommand(),
		},
	}
}
