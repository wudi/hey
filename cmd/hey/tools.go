package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var initCommand = &cli.Command{
	Name:   "init",
	Usage:  "Creates a composer.json file in the current directory",
	Flags:  []cli.Flag{},
	Action: initAction,
}

func initAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Creates a composer.json file in the current directory")
	return nil
}

var requireCommand = &cli.Command{
	Name:   "require",
	Usage:  "Adds required packages to your composer.json and installs them",
	Flags:  []cli.Flag{},
	Action: requireAction,
}

func requireAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Adds required packages to your composer.json and installs them")
	return nil
}

var installCommand = &cli.Command{
	Name:    "install",
	Aliases: []string{"i"},
	Usage:   "Installs your composer.json and installs them",
	Flags:   []cli.Flag{},
	Action:  installAction,
}

func installAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Installs the project dependencies from the composer.lock file if present, or falls back on the composer.json")
	return nil
}

var updateCommand = &cli.Command{
	Name:    "update",
	Aliases: []string{"u"},
	Usage:   "Updates your dependencies to the latest version according to composer.json, and updates the composer.lock file",
	Flags:   []cli.Flag{},
	Action:  updateAction,
}

func updateAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Updates your dependencies to the latest version according to composer.json, and updates the composer.lock file")
	return nil
}

var validateCommand = &cli.Command{
	Name:   "validate",
	Usage:  "Validates a composer.json file",
	Flags:  []cli.Flag{},
	Action: validateAction,
}

func validateAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Validates a composer.json file")
	return nil
}

var fpmCommand = &cli.Command{
	Name:  "fpm",
	Usage: "FastCGI process manager",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "c",
			Usage: "Look for php.ini file in this directory",
		},
		&cli.StringFlag{
			Name:    "fpm-config",
			Aliases: []string{"y"},
			Usage:   "Specify alternative path to FastCGI process manager config file.",
		},
	},
	Action: fpmAction,
}

func fpmAction(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Run PHP script with FPM")
	return nil
}
