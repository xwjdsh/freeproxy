package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	homeDir, _ := os.UserHomeDir()
	app := &cli.App{
		Usage: "A command-line free proxy manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   fmt.Sprintf("%s/.config/freeproxy/config.yml", homeDir),
				Usage:   "Set config file path",
			},
		},
		Commands: []*cli.Command{
			configCommand,
			fetchCommand,
			tidyCommand,
			summaryCommand,
			exportCommand,
			proxyCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
