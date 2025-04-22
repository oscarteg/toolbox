package main

import (
	"context"
	"log"
	"os"

	commands "github.com/oscarteg/toolbox/internal/app"
)

func main() {
	cmd := commands.RootCommand()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
