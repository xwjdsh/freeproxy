package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/xwjdsh/freeproxy"
	"github.com/xwjdsh/freeproxy/config"
)

func main() {
	homeDir, _ := os.UserHomeDir()
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   fmt.Sprintf("%s/.config/freeproxy/freeproxy.yml", homeDir),
				Usage:   "Set config file path",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "config",
				Aliases: []string{"c"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "default",
						Aliases: []string{"d"},
						Usage:   "Display default config",
					},
				},
				Usage: "Display config",
				Action: func(c *cli.Context) error {
					var cfg *config.Config
					if c.Bool("default") {
						cfg = config.DefaultConfig()
					} else {
						var err error
						cfg, err = config.Init(c.String("config"))
						if err != nil {
							return err
						}
					}
					data, err := cfg.Marshal()
					if err != nil {
						return err
					}
					fmt.Printf(string(data))
					return nil
				},
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Parse, validate and save proxies",
				Action: func(c *cli.Context) error {
					cfg, err := config.Init(c.String("config"))
					if err != nil {
						return err
					}
					h, err := freeproxy.Init(cfg)
					if err != nil {
						return err
					}
					h.Start(c.Context)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
