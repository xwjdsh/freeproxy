package freeproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func (h *Handler) Proxy(ctx context.Context, fast bool, switchProxy bool) error {
	if switchProxy {
		resp, err := http.PostForm(fmt.Sprintf("http://%s/switch", h.cfg.Proxy.SwitchServer), nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		p := &storage.Proxy{}
		if err := json.Unmarshal(data, p); err != nil {
			return err
		}

		fmt.Printf("select id: %d, server: %s, type: %s, country: %s\n", p.ID, p.Server, p.Type, p.Country)
		return nil
	}

	if _, err := h.startProxyServer(ctx, fast); err != nil {
		return err
	}

	go func() {
		http.HandleFunc("/switch", func(rw http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				code := http.StatusMethodNotAllowed
				rw.Write([]byte(http.StatusText(code)))
				rw.WriteHeader(code)
				return
			}

			p, err := h.startProxyServer(ctx, fast)
			if err != nil {
				rw.Write([]byte(err.Error()))
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			data, err := json.Marshal(p)
			if err != nil {
				rw.Write([]byte(err.Error()))
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			rw.Write(data)
			rw.Header().Set("Content-Type", "application/json")
		})
		http.ListenAndServe(h.cfg.Proxy.SwitchServer, nil)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	return nil
}

func (h *Handler) startProxyServer(ctx context.Context, fast bool) (*storage.Proxy, error) {
	cfg := h.cfg.Proxy
	ps, err := h.storage.GetProxies(ctx, &storage.QueryOptions{
		ID:              cfg.ProxyID,
		CountryCodes:    cfg.ProxyCountryCodes,
		NotCountryCodes: cfg.ProxyNotCountryCodes,
		Count:           1,
		Fast:            fast,
	})
	if err != nil {
		return nil, err
	}
	if len(ps) == 0 {
		return nil, fmt.Errorf("no proxy records")
	}

	p := ps[0]
	fmt.Printf("select id: %d, server: %s, type: %s, country: %s\n", p.ID, p.Server, p.Type, p.Country)
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(p.Config), &m); err == nil {
		m["name"] = "proxy"
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	t, err := template.New("").Parse(proxyClashTemplate)
	if err != nil {
		return nil, err
	}

	fp := filepath.Join(os.TempDir(), uuid.NewString()+".yaml")
	f, err := os.Create(fp)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	C.SetConfig(fp)
	if err := hub.Parse(); err != nil {
		return nil, err
	}

	return p, nil
}
