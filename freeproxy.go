package freeproxy

import (
	"context"
	"os"
	"sort"
	"strconv"
	"time"

	emoji "github.com/jayco/go-emoji-flag"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/sync/errgroup"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/exporter"
	"github.com/xwjdsh/freeproxy/internal/progressbar"
	"github.com/xwjdsh/freeproxy/log"
	"github.com/xwjdsh/freeproxy/parser"
	"github.com/xwjdsh/freeproxy/storage"
	"github.com/xwjdsh/freeproxy/validator"
)

type Handler struct {
	cfg       *config.Config
	parser    *parser.Handler
	validator *validator.Validator
	storage   *storage.Handler
	exporter  *exporter.Exporter
}

func Init(cfg *config.Config) (*Handler, error) {
	log.Init(cfg.Log)

	h, err := storage.Init(cfg.Storage)
	if err != nil {
		return nil, err
	}
	p, err := parser.Init(cfg.Parser)
	if err != nil {
		return nil, err
	}
	return &Handler{
		cfg:       cfg,
		parser:    p,
		validator: validator.New(cfg.Validator),
		storage:   h,
		exporter:  exporter.New(cfg.Exporter),
	}, nil
}

func (h *Handler) Tidy(ctx context.Context) error {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return err
	}

	if err := h.validator.CheckNetwork(ctx); err != nil {
		return err
	}

	g := new(errgroup.Group)
	pb := progressbar.New(len(ps))
	removed := 0
	setCountry, emptyCountry := 0, 0
	for _, p := range ps {
		p := p
		g.Go(func() error {
			defer func() {
				pb.SetSuffix("removed: %d, setCountry: %d, emptyCountry: %d", removed, setCountry, emptyCountry)
				pb.Incr()
			}()
			pp, err := p.Restore(p.Config)
			if err != nil {
				return err
			}
			if err := h.validator.Validate(ctx, pp); err != nil {
				err := h.storage.Remove(ctx, p.ID)
				if err == nil {
					removed += 1
				}
				return err
			}

			if p.CountryCode == "" {
				p.CountryCode, p.Country, _ = h.validator.GetCountryInfo(ctx, p.Server)
				if p.CountryCode != "" {
					setCountry += 1
				} else {
					emptyCountry += 1
				}
			}

			return h.storage.Update(ctx, p)
		})
	}

	pb.Wait()
	return g.Wait()
}

func (h *Handler) Fetch(ctx context.Context) error {
	parserResultChan := make(chan *parser.Result)
	g := new(errgroup.Group)

	g.Go(func() error {
		h.parser.Parse(ctx, parserResultChan)
		close(parserResultChan)
		return nil
	})

	pb := progressbar.New(0)
	g.Go(func() error {
		ng := new(errgroup.Group)
		total := 0
		createdCount := 0
		for {
			select {
			case r, ok := <-parserResultChan:
				if !ok {
					err := ng.Wait()
					pb.SetTotal(total, true)
					// wait progressbar 100%
					time.Sleep(100 * time.Millisecond)
					return err
				}
				if r.Err != nil {
					continue
				}
				total += 1
				pb.SetTotal(total, false)

				ng.Go(func() error {
					defer func() {
						pb.SetSuffix("created: %d", createdCount)
						pb.Incr()
					}()

					if err := h.validator.Validate(ctx, r.Proxy); err != nil {
						return nil
					}
					_, created, err := h.storage.Create(ctx, r.Proxy)
					if created {
						createdCount += 1
					}
					return err
				})
			}
		}
	})

	return g.Wait()
}

func (h *Handler) Export(ctx context.Context) error {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return nil
	}

	return h.exporter.Export(ps)
}

func (h *Handler) Summary(ctx context.Context) error {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return nil
	}

	countryMap := map[string]string{}
	countMap := map[string]int{}
	for _, p := range ps {
		countMap[p.CountryCode] += 1
		if _, ok := countryMap[p.CountryCode]; !ok {
			countryMap[p.CountryCode] = p.Country
		}
	}

	codes := []string{}
	for k := range countMap {
		codes = append(codes, k)
	}

	sort.Slice(codes, func(i, j int) bool {
		return countMap[codes[i]] > countMap[codes[j]]
	})

	data := [][]string{}
	for _, code := range codes {
		countryEmoji := emoji.GetFlag(code)
		data = append(data, []string{
			countryEmoji, code, countryMap[code], strconv.Itoa(countMap[code]),
		})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"", "CountryCode", "Country", "Amount"})
	table.SetFooter([]string{"", "", "Total", strconv.Itoa(len(ps))}) // Add Footer
	table.SetBorder(false)                                            // Set Border to false
	table.AppendBulk(data)                                            // Add Bulk Data
	table.Render()
	return nil
}
