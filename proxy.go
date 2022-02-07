package freeproxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Dreamacro/clash/adapter"
	C "github.com/Dreamacro/clash/constant"
	"github.com/elazarl/goproxy"
)

type ProxyOptions struct {
	Address string
	Verbose bool

	ID          uint
	CountryCode string
}

func (h *Handler) Proxy(ctx context.Context, opts *ProxyOptions) error {
	p, err := h.storage.GetProxy(ctx, opts.ID, opts.CountryCode)
	if err != nil {
		return err
	}

	log.Printf("select id: %d, server: %s, country: %s\n", p.ID, p.Server, p.Country)
	pp, err := p.Restore(p.Config)
	if err != nil {
		return err
	}

	m, err := pp.ConfigMap()
	if err != nil {
		return err
	}

	clashProxy, err := adapter.ParseProxy(m)
	if err != nil {
		return err
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = opts.Verbose
	proxy.OnRequest().DoFunc(h.proxyOnRequest(ctx, clashProxy))

	log.Printf("server running on: %s", opts.Address)
	return http.ListenAndServe(opts.Address, proxy)
}

func (h *Handler) proxyOnRequest(ctx context.Context, p C.Proxy) func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	return func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		addr, err := urlToMetadata(r.URL)
		if err != nil {
			return r, goproxy.NewResponse(r,
				goproxy.ContentTypeText, http.StatusInternalServerError,
				err.Error())
		}

		instance, err := p.DialContext(context.Background(), &addr)
		if err != nil {
			return r, goproxy.NewResponse(r,
				goproxy.ContentTypeText, http.StatusInternalServerError,
				err.Error())
		}
		defer instance.Close()

		transport := &http.Transport{
			Dial: func(string, string) (net.Conn, error) {
				return instance, nil
			},
			// from http.DefaultTransport
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		client := http.Client{
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		defer client.CloseIdleConnections()

		r.RequestURI = ""
		resp, err := client.Do(r)
		if err != nil {
			return r, goproxy.NewResponse(r,
				goproxy.ContentTypeText, http.StatusInternalServerError,
				err.Error())
		}

		return r, resp
	}
}

func urlToMetadata(u *url.URL) (addr C.Metadata, err error) {
	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			err = fmt.Errorf("%s scheme not Support", u.String())
			return
		}
	}

	addr = C.Metadata{
		AddrType: C.AtypDomainName,
		Host:     u.Hostname(),
		DstIP:    nil,
		DstPort:  port,
	}
	return
}
