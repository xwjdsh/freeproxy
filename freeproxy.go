package freeproxy

import (
	"context"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"sync"
	"text/template"

	"github.com/fatih/color"
	emoji "github.com/jayco/go-emoji-flag"
	"github.com/olekukonko/tablewriter"

	"github.com/xwjdsh/freeproxy/config"
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

	var pb progressbar.ProgressBar
	if quiet {
		pb = progressbar.NewMock()
	} else {
		pb = progressbar.New()
	}
	pb.AddBar("", len(ps))

	bar := pb.Bar("")
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
							bar.SetSuffix("removed: %d, setCountry: %d, emptyCountry: %d", removedCount.Get(), setCountryCount.Get(), emptyCountryCount.Get())
							bar.Incr()
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

	pb.Wait()
	wg.Wait()

	return nil
}

func (h *Handler) Fetch(ctx context.Context, quiet bool) error {
	var pb progressbar.ProgressBar
	if quiet {
		pb = progressbar.NewMock()
	} else {
		pb = progressbar.New()
	}

	parserResultChan := make(chan *parser.Result)

	wg := sync.WaitGroup{}
	wg.Add(h.cfg.Worker)

	createdCountMap := map[string]int{}
	createdCountMutex := sync.Mutex{}

	barMutex := sync.Mutex{}

	for i := 0; i < h.cfg.Worker; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case r, ok := <-parserResultChan:
					if !ok {
						return
					}

					source := r.Source
					bar := func() progressbar.Bar {
						barMutex.Lock()
						defer barMutex.Unlock()

						if b := pb.Bar(source); b != nil {
							return b
						}

						return pb.AddBar(source, 0)
					}()

					if r.SourceDone {
						if r.Err != nil {
							bar.SetSuffix(color.RedString(r.Err.Error()))
						}

						bar.Wait()
						bar.TriggerComplete()
						continue
					}

					bar.TotalInc(1)

					if err := func() error {
						defer func() {
							createdCountMutex.Lock()
							defer createdCountMutex.Unlock()

							if v := createdCountMap[source]; v > 0 {
								bar.SetSuffix(color.GreenString("new: %d", v))
							}

							bar.Incr()
						}()

						if err := h.validator.Validate(ctx, r.Proxy); err != nil {
							return nil
						}
						_, created, err := h.storage.Create(ctx, r.Proxy)
						if created {
							func() {
								createdCountMutex.Lock()
								defer createdCountMutex.Unlock()

								createdCountMap[source] += 1
							}()
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
	pb.Wait()

	return nil
}

type SummaryGroup struct {
	CountryCode  string
	CountryEmoji string
	Country      string
	Total        int
	ProxyTypeMap map[string]int
}

type SummaryData struct {
	Items             []*SummaryGroup
	TotalProxyTypeMap map[string]int
	Total             int
}

func (h *Handler) Summary(ctx context.Context, templatePath string) error {
	ps, err := h.storage.GetProxies(ctx)
	if err != nil {
		return nil
	}

	sd := &SummaryData{
		Total:             len(ps),
		TotalProxyTypeMap: map[string]int{},
	}

	m := map[string]*SummaryGroup{}
	for _, p := range ps {
		sg := m[p.CountryCode]
		if sg == nil {
			sg = &SummaryGroup{
				CountryCode:  p.CountryCode,
				CountryEmoji: emoji.GetFlag(p.CountryCode),
				Country:      p.Country,
				ProxyTypeMap: map[string]int{},
			}
			m[p.CountryCode] = sg
			sd.Items = append(sd.Items, sg)
		}
		sg.Total += 1
		sg.ProxyTypeMap[p.Type.String()] += 1
		sd.TotalProxyTypeMap[p.Type.String()] += 1
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
	header := []string{"CountryCode", "Country"}
	footer := []string{"", "Total"}
	if v := sd.TotalProxyTypeMap[proxy.SS.String()]; v != 0 {
		header = append(header, "SS")
		footer = append(footer, strconv.Itoa(v))
	}
	if v := sd.TotalProxyTypeMap[proxy.SSR.String()]; v != 0 {
		header = append(header, "SSR")
		footer = append(footer, strconv.Itoa(v))
	}
	if v := sd.TotalProxyTypeMap[proxy.Vmess.String()]; v != 0 {
		header = append(header, "Vmess")
		footer = append(footer, strconv.Itoa(v))
	}
	if v := sd.TotalProxyTypeMap[proxy.Trojan.String()]; v != 0 {
		header = append(header, "Trojan")
		footer = append(footer, strconv.Itoa(v))
	}
	header = append(header, "Total")
	footer = append(footer, strconv.Itoa(sd.Total))

	for _, item := range sd.Items {
		dataItem := []string{
			item.CountryEmoji + " " + item.CountryCode, item.Country,
		}

		if v := sd.TotalProxyTypeMap[proxy.SS.String()]; v != 0 {
			dataItem = append(dataItem, strconv.Itoa(item.ProxyTypeMap[proxy.SS.String()]))
		}
		if v := sd.TotalProxyTypeMap[proxy.SSR.String()]; v != 0 {
			dataItem = append(dataItem, strconv.Itoa(item.ProxyTypeMap[proxy.SSR.String()]))
		}
		if v := sd.TotalProxyTypeMap[proxy.Vmess.String()]; v != 0 {
			dataItem = append(dataItem, strconv.Itoa(item.ProxyTypeMap[proxy.Vmess.String()]))
		}
		if v := sd.TotalProxyTypeMap[proxy.Trojan.String()]; v != 0 {
			dataItem = append(dataItem, strconv.Itoa(item.ProxyTypeMap[proxy.Trojan.String()]))
		}
		dataItem = append(dataItem, strconv.Itoa(item.Total))
		data = append(data, dataItem)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader(header)
	table.SetFooter(footer)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.Render()
	return nil
}
