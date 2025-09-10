package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var composerCommand = &cli.Command{
	Name: "composer",
	Commands: []*cli.Command{
		{
			Name:    "version",
			Aliases: []string{"V"},
			Usage:   "Show composer version",
			Action:  composerShowHelp,
		},
		{
			Name:   "install",
			Usage:  "Install the project dependencies from the composer.lock file if present, or fall back on the composer.json.",
			Action: composerInstall,
		},
	},
	Action: composerShowHelp,
}

func composerShowHelp(ctx context.Context, cmd *cli.Command) error {
	return cli.ShowAppHelp(cmd)
}

func composerInstall(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Installing project dependencies from the composer.lock file.")
	return nil
}
