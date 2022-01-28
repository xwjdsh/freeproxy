package exporter

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/storage"
)

//go:embed clash.tmpl
var defaultTemplate string

type Exporter struct {
	cfg *config.ExporterConfig
}

func New(cfg *config.ExporterConfig) *Exporter {
	return &Exporter{
		cfg: cfg,
	}
}

func (e *Exporter) Export(ps []*storage.Proxy, tfp string) (string, error) {
	proxies := []map[string]interface{}{}
	for _, p := range ps {
		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(p.Config), &m); err == nil {
			m["name"] = fmt.Sprintf("%s-%s-%d", p.CountryCode, p.CountryEmoji, p.ID)
			proxies = append(proxies, m)
		}
	}

	if tfp == "" {
		tfp = e.cfg.TemplateFilePath
	}

	text := defaultTemplate
	if tfp != "" {
		data, err := ioutil.ReadFile(tfp)
		if err != nil {
			return "", err
		}
		text = string(data)
	}

	t, err := template.New("").Parse(text)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, proxies); err != nil {
		return "", err
	}

	return buf.String(), nil
}
