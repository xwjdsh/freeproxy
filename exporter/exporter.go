package exporter

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/template"

	emoji "github.com/jayco/go-emoji-flag"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/storage"
)

//go:embed clash.tmpl
var defaultTemplate string

type RenderItem struct {
	Proxy  *storage.Proxy
	Config string
}

type RenderData struct {
	Items []*RenderItem
}

type Exporter struct {
	cfg *config.ExporterConfig
}

func New(cfg *config.ExporterConfig) *Exporter {
	return &Exporter{
		cfg: cfg,
	}
}

func (e *Exporter) Export(ps []*storage.Proxy) error {
	rd := &RenderData{}
	for _, p := range ps {
		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(p.Config), &m); err == nil {
			countryEmoji := emoji.GetFlag(p.CountryCode)
			p.Name = fmt.Sprintf("%s-%s-%d", countryEmoji, p.CountryCode, p.ID)
			m["name"] = p.Name
		}
		data, err := json.Marshal(m)
		if err != nil {
			return err
		}
		rd.Items = append(rd.Items, &RenderItem{
			Proxy:  p,
			Config: string(data),
		})
	}

	text := defaultTemplate
	if fp := e.cfg.TemplateFilePath; fp != "" {
		data, err := ioutil.ReadFile(fp)
		if err != nil {
			return err
		}
		text = string(data)
	}

	t, err := template.New("").Parse(text)
	if err != nil {
		return err
	}

	var wr io.Writer = os.Stdout
	outputPath := e.cfg.OutputFilePath
	if fp := outputPath; fp != "" {
		f, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		wr = f
	}

	if err := t.Execute(wr, rd); err != nil {
		return err
	}

	if outputPath != "" {
		fmt.Printf("file save to: %s\n", outputPath)
	}

	return nil
}
