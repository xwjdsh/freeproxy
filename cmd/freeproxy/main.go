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
		Usage: "A command-line free proxy manager",
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
				Name:    "fetch",
				Aliases: []string{"f"},
				Usage:   "Fetch new proxies",
				Action: func(c *cli.Context) error {
					h, err := getHandler(c)
					if err != nil {
						return err
					}
					return h.Fetch(c.Context)
				},
			},
			{
				Name:    "tidy",
				Aliases: []string{"t"},
				Usage:   "Tidy saved proxies, remove disabled records",
				Action: func(c *cli.Context) error {
					h, err := getHandler(c)
					if err != nil {
						return err
					}
					return h.Tidy(c.Context)
				},
			},
			{
				Name:    "summary",
				Aliases: []string{"s"},
				Usage:   "Display saved proxies summary",
				Action: func(c *cli.Context) error {
					h, err := getHandler(c)
					if err != nil {
						return err
					}
					return h.Summary(c.Context)
				},
			},
			{
				Name:    "export",
				Aliases: []string{"e"},
				Usage:   "Export proxies",
				Action: func(c *cli.Context) error {
					h, err := getHandler(c)
					if err != nil {
						return err
					}
					return h.Export(c.Context)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getHandler(c *cli.Context) (*freeproxy.Handler, error) {
	cfg, err := config.Init(c.String("config"))
	if err != nil {
		return nil, err
	}
	h, err := freeproxy.Init(cfg)
	if err != nil {
		return nil, err
	}
	return h, nil
}
