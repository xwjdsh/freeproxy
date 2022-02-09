package freeproxy

import (
	"context"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"sync"
	"text/template"

	emoji "github.com/jayco/go-emoji-flag"
	"github.com/olekukonko/tablewriter"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/exporter"
	"github.com/xwjdsh/freeproxy/internal/counter"
	"github.com/xwjdsh/freeproxy/internal/progressbar"
	"github.com/xwjdsh/freeproxy/log"
	"github.com/xwjdsh/freeproxy/parser"
	"github.com/xwjdsh/freeproxy/proxy"
	"github.com/xwjdsh/freeproxy/storage"
	"github.com/xwjdsh/freeproxy/validator"
)

type Handler struct {
	cfg       *config.AppConfig
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
		cfg:       cfg.App,
		parser:    p,
		validator: validator.New(cfg.Validator),
		storage:   h,
		exporter:  exporter.New(cfg.Exporter),
	}, nil
}

func (h *Handler) Tidy(ctx context.Context, quiet bool) error {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return err
	}

	if err := h.validator.CheckNetwork(ctx); err != nil {
		return err
	}

	pb := progressbar.New(len(ps), quiet)
	proxyChan := make(chan *storage.Proxy)

	var (
		removedCount      counter.Count
		setCountryCount   counter.Count
		emptyCountryCount counter.Count
	)

	wg := sync.WaitGroup{}
	wg.Add(h.cfg.Worker)

	for i := 0; i < h.cfg.Worker; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case p, ok := <-proxyChan:
					if !ok {
						return
					}

					if err := func() error {
						defer func() {
							pb.SetSuffix("removed: %d, setCountry: %d, emptyCountry: %d", removedCount.Get(), setCountryCount.Get(), emptyCountryCount.Get())
							pb.Incr()
						}()

						pp, err := p.Restore(p.Config)
						if err != nil {
							return err
						}
						if err := h.validator.Validate(ctx, pp); err != nil {
							err := h.storage.Remove(ctx, p.ID)
							if err == nil {
								removedCount.Inc()
							}
							return err
						}

						if p.CountryCode == "" {
							p.CountryCode, p.Country, _ = h.validator.GetCountryInfo(ctx, p.Server)
							if p.CountryCode != "" {
								setCountryCount.Inc()
							} else {
								emptyCountryCount.Inc()
							}
						}

						return h.storage.Update(ctx, p)
					}(); err != nil {
						// TODO log
					}
				}
			}
		}()
	}

	for _, p := range ps {
		p := p
		proxyChan <- p
	}
	close(proxyChan)

	wg.Wait()
	pb.Wait()

	return nil
}

func (h *Handler) Fetch(ctx context.Context, quiet bool) error {
	var (
		createdCount counter.Count
		total        counter.Count
	)

	pb := progressbar.New(0, quiet)
	parserResultChan := make(chan *parser.Result)

	wg := sync.WaitGroup{}
	wg.Add(h.cfg.Worker)

	for i := 0; i < h.cfg.Worker; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case r, ok := <-parserResultChan:
					if !ok {
						return
					}
					if r.Err != nil {
						continue
					}

					total.Inc()
					pb.SetTotal(int(total.Get()), false)

					if err := func() error {
						defer func() {
							pb.SetSuffix("created: %d", createdCount.Get())
							pb.Incr()
						}()

						if err := h.validator.Validate(ctx, r.Proxy); err != nil {
							return nil
						}
						_, created, err := h.storage.Create(ctx, r.Proxy)
						if created {
							createdCount.Inc()
						}
						return err
					}(); err != nil {
						// TODO log
					}
				}
			}
		}()
	}

	h.parser.Parse(ctx, parserResultChan)
	close(parserResultChan)
	wg.Wait()

	pb.SetTotal(total.Get(), true)
	pb.Wait()

	return nil
}

func (h *Handler) Export(ctx context.Context) error {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return nil
	}

	return h.exporter.Export(ps)
}

type SummaryGroup struct {
	CountryCode  string
	CountryEmoji string
	Country      string
	Total        int
	ProxyTypeMap map[proxy.Type]int
}

type SummaryData struct {
	Items             []*SummaryGroup
	TotalProxyTypeMap map[proxy.Type]int
	Total             int
}

func (h *Handler) Summary(ctx context.Context, templatePath string) error {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return nil
	}

	sd := &SummaryData{
		Total:             len(ps),
		TotalProxyTypeMap: map[proxy.Type]int{},
	}

	m := map[string]*SummaryGroup{}
	for _, p := range ps {
		sg := m[p.CountryCode]
		if sg == nil {
			sg = &SummaryGroup{
				CountryCode:  p.CountryCode,
				CountryEmoji: emoji.GetFlag(p.CountryCode),
				Country:      p.Country,
				ProxyTypeMap: map[proxy.Type]int{},
			}
			m[p.CountryCode] = sg
			sd.Items = append(sd.Items, sg)
		}
		sg.Total += 1
		sg.ProxyTypeMap[p.Type] += 1
		sd.TotalProxyTypeMap[p.Type] += 1
	}

	sort.Slice(sd.Items, func(i, j int) bool {
		return sd.Items[i].Total > sd.Items[j].Total
	})

	if templatePath != "" {
		data, err := ioutil.ReadFile(templatePath)
		if err != nil {
			return err
		}

		t, err := template.New("").Parse(string(data))
		if err != nil {
			return err
		}

		return t.Execute(os.Stdout, sd)
	}

	data := [][]string{}
	for _, item := range sd.Items {
		data = append(data, []string{
			item.CountryEmoji + " " + item.CountryCode, item.Country, strconv.Itoa(item.ProxyTypeMap[proxy.SS]),
			strconv.Itoa(item.ProxyTypeMap[proxy.SSR]), strconv.Itoa(item.ProxyTypeMap[proxy.Vmess]),
			strconv.Itoa(item.Total),
		})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"CountryCode", "Country", "SS", "SSR", "Vmess", "Total"})
	table.SetFooter([]string{
		"", "Total",
		strconv.Itoa(sd.TotalProxyTypeMap[proxy.SS]),
		strconv.Itoa(sd.TotalProxyTypeMap[proxy.SSR]),
		strconv.Itoa(sd.TotalProxyTypeMap[proxy.Vmess]),
		strconv.Itoa(sd.Total)})
	table.SetBorder(false)
	table.AppendBulk(data)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.Render()
	return nil
}
