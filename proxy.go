package freeproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"

	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/hub"
	"github.com/google/uuid"

	"github.com/xwjdsh/freeproxy/storage"
)

type ProxyOptions struct {
	BindAddress string
	Port        int
	Verbose     bool

	ID              uint
	CountryCodes    string
	NotCountryCodes string
}

var proxyClashTemplate = `
---
bind-address: {{ .BindAddress }}
mixed-port: {{ .Port }}
allow-lan: false
mode: rule
log-level: info
profile:
  store-selected: false
  store-fake-ip: false

proxies:
  - {{ .Proxy }}
rules:
  - MATCH,proxy
`

type proxyRenderData struct {
	BindAddress string
	Port        int
	Proxy       string
}

func (h *Handler) Proxy(ctx context.Context) error {
	cfg := h.cfg.Proxy
	ps, err := h.storage.GetProxies(ctx, &storage.QueryOptions{
		ID:              cfg.ProxyID,
		CountryCodes:    cfg.ProxyCountryCodes,
		NotCountryCodes: cfg.ProxyNotCountryCodes,
		Count:           1,
	})
	if err != nil {
		return err
	}
	if len(ps) == 0 {
		return fmt.Errorf("no proxy records")
	}

	p := ps[0]
	log.Printf("select id: %d, server: %s, country: %s\n", p.ID, p.Server, p.Country)
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(p.Config), &m); err == nil {
		m["name"] = "proxy"
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	t, err := template.New("").Parse(proxyClashTemplate)
	if err != nil {
		return err
	}

	fp := filepath.Join(os.TempDir(), uuid.NewString()+".yaml")
	f, err := os.Create(fp)
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
		os.Remove(fp)
	}()

	if err := t.Execute(f, &proxyRenderData{
		BindAddress: cfg.BindAddress,
		Port:        cfg.Port,
		Proxy:       string(data),
	}); err != nil {
		return err
	}

	C.SetConfig(fp)
	if err := hub.Parse(); err != nil {
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	return nil
}
