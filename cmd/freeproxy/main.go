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
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "quiet",
						Aliases: []string{"q"},
						Usage:   "Quiet mode, do not display progress bar",
					},
					&cli.IntFlag{
						Name:    "worker",
						Aliases: []string{"w"},
						Usage:   "Worker count",
					},
				},
				Action: func(c *cli.Context) error {
					h, err := getHandler(c, func(cfg *config.Config) {
						cfg.App.Worker = c.Int("worker")
					})
					if err != nil {
						return err
					}
					return h.Fetch(c.Context, c.Bool("quiet"))
				},
			},
			{
				Name:    "tidy",
				Aliases: []string{"t"},
				Usage:   "Tidy saved proxies, remove disabled records",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "quiet",
						Aliases: []string{"q"},
						Usage:   "Quiet mode, do not display progress bar",
					},
					&cli.BoolFlag{
						Name:    "worker",
						Aliases: []string{"w"},
						Usage:   "Worker count",
					},
				},
				Action: func(c *cli.Context) error {
					h, err := getHandler(c, func(cfg *config.Config) {
						cfg.App.Worker = c.Int("worker")
					})
					if err != nil {
						return err
					}
					return h.Tidy(c.Context, c.Bool("quiet"))
				},
			},
			{
				Name:    "summary",
				Aliases: []string{"s"},
				Usage:   "Display saved proxies summary",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "template",
						Aliases: []string{"t"},
						Usage:   "Output template",
					},
				},
				Action: func(c *cli.Context) error {
					h, err := getHandler(c, nil)
					if err != nil {
						return err
					}
					return h.Summary(c.Context, c.String("template"))
				},
			},
			{
				Name:    "export",
				Aliases: []string{"e"},
				Usage:   "Export proxies",
				Action: func(c *cli.Context) error {
					h, err := getHandler(c, nil)
					if err != nil {
						return err
					}
					return h.Export(c.Context)
				},
			},
			{
				Name:    "proxy",
				Aliases: []string{"p"},
				Usage:   "Start http proxy server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "address",
						Aliases: []string{"a"},
						Usage:   "Server listen address",
						Value:   "127.0.0.1",
					},
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "Server listen port",
						Value:   10000,
					},
					&cli.StringFlag{
						Name:    "country-code",
						Aliases: []string{"cc"},
						Usage:   "Filter proxies by country code",
					},
					&cli.UintFlag{
						Name:  "id",
						Usage: "Filter proxies by id",
					},
				},
				Action: func(c *cli.Context) error {
					h, err := getHandler(c, nil)
					if err != nil {
						return err
					}
					opts := &freeproxy.ProxyOptions{
						BindAddress: c.String("address"),
						Port:        c.Int("port"),
						ID:          c.Uint("id"),
						CountryCode: c.String("country-code"),
					}
					return h.Proxy(c.Context, opts)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getHandler(c *cli.Context, f func(cfg *config.Config)) (*freeproxy.Handler, error) {
	cfg, err := config.Init(c.String("config"))
	if err != nil {
		return nil, err
	}

	if f != nil {
		f(cfg)
	}

	h, err := freeproxy.Init(cfg)
	if err != nil {
		return nil, err
	}
	return h, nil
}
