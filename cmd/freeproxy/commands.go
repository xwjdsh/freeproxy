package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/xwjdsh/freeproxy"
	"github.com/xwjdsh/freeproxy/config"
)

var (
	configCommand = &cli.Command{
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
	}

	fetchCommand = &cli.Command{
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
				cc := cfg.App.Fetch
				if v := c.Int("worker"); v != 0 {
					cc.Worker = v
				}
			})
			if err != nil {
				return err
			}
			return h.Fetch(c.Context, c.Bool("quiet"))
		},
	}

	tidyCommand = &cli.Command{
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
				cc := cfg.App.Tidy
				if v := c.Int("worker"); v != 0 {
					cc.Worker = v
				}
			})
			if err != nil {
				return err
			}
			return h.Tidy(c.Context, c.Bool("quiet"))
		},
	}

	summaryCommand = &cli.Command{
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
			h, err := getHandler(c, func(cfg *config.Config) {
				cc := cfg.App.Summary
				if v := c.String("template"); v != "" {
					cc.TemplateFilePath = v
				}
			})
			if err != nil {
				return err
			}
			return h.Summary(c.Context)
		},
	}

	exportCommand = &cli.Command{
		Name:    "export",
		Aliases: []string{"e"},
		Usage:   "Export proxies",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "template",
				Aliases: []string{"t"},
				Usage:   "Set template file path",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Set output file path, stdout if not set",
			},
			&cli.StringFlag{
				Name:    "country-code",
				Aliases: []string{"cc"},
				Usage:   "Filter proxies by country codes, for example 'US,DE'",
			},
			&cli.StringFlag{
				Name:    "not-country-code",
				Aliases: []string{"ncc"},
				Usage:   "Filter proxies other than country codes, for example 'CN,IN'",
			},
			&cli.UintFlag{
				Name:  "id",
				Usage: "Filter proxies by id",
			},
			&cli.IntFlag{
				Name:    "count",
				Aliases: []string{"c"},
				Usage:   "Get the top N fastest proxies",
			},
		},
		Action: func(c *cli.Context) error {
			h, err := getHandler(c, func(cfg *config.Config) {
				cc := cfg.App.Export
				if v := c.String("output"); v != "" {
					cc.OutputFilePath = v
				}
				if v := c.String("template"); v != "" {
					cc.TemplateFilePath = v
				}
				if v := c.String("country-code"); v != "" {
					cc.ProxyCountryCodes = v
				}
				if v := c.String("not-country-code"); v != "" {
					cc.ProxyNotCountryCodes = v
				}
				if v := c.Uint("id"); v != 0 {
					cc.ProxyID = v
				}
				if v := c.Int("count"); v != 0 {
					cc.ProxyCount = v
				}
			})
			if err != nil {
				return err
			}
			return h.Export(c.Context)
		},
	}

	proxyCommand = &cli.Command{
		Name:    "proxy",
		Aliases: []string{"p"},
		Usage:   "Start proxy server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Aliases: []string{"a"},
				Usage:   "Server listen address",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Server listen port",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Verbose log",
			},
			&cli.StringFlag{
				Name:    "country-code",
				Aliases: []string{"cc"},
				Usage:   "Filter proxies by country codes, for example 'US,DE'",
			},
			&cli.StringFlag{
				Name:    "not-country-code",
				Aliases: []string{"ncc"},
				Usage:   "Filter proxies other than country codes, for example 'CN,IN'",
			},
			&cli.BoolFlag{
				Name:    "fast",
				Aliases: []string{"f"},
				Usage:   "Get the fastest proxy",
			},
			&cli.BoolFlag{
				Name:    "switch",
				Aliases: []string{"s"},
				Usage:   "Switch proxy server",
			},
			&cli.UintFlag{
				Name:  "id",
				Usage: "Filter proxies by id",
			},
		},
		Action: func(c *cli.Context) error {
			h, err := getHandler(c, func(cfg *config.Config) {
				cc := cfg.App.Proxy
				if v := c.String("address"); v != "" {
					cc.BindAddress = v
				}
				if v := c.Int("port"); v != 0 {
					cc.Port = v
				}
				if v := c.Bool("verbose"); v {
					cc.Verbose = v
				}
				if v := c.String("country-code"); v != "" {
					cc.ProxyCountryCodes = v
				}
				if v := c.String("not-country-code"); v != "" {
					cc.ProxyNotCountryCodes = v
				}
			})
			if err != nil {
				return err
			}
			return h.Proxy(c.Context, c.Bool("fast"), c.Bool("switch"))
		},
	}
)

func getIntOrDefault(v, d int) int {
	if v != 0 {
		return v
	}

	return d
}

func getStringOrDefault(v, d string) string {
	if v != "" {
		return v
	}

	return d
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
