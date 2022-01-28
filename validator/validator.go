package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/Dreamacro/clash/adapter"
	emoji "github.com/jayco/go-emoji-flag"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/parser"
	"github.com/xwjdsh/freeproxy/proxy"
)

type Result struct {
	Proxy        proxy.Proxy
	Delay        uint16
	Country      string
	CountryCode  string
	CountryEmoji string
	Error        error
}

type Validator struct {
	parserChan <-chan *parser.Result
	ch         chan *Result
	cfg        *config.ValidatorConfig
}

func New(parserChan <-chan *parser.Result, cfg *config.ValidatorConfig) *Validator {
	return &Validator{
		ch:         make(chan *Result),
		parserChan: parserChan,
		cfg:        cfg,
	}
}

func (v *Validator) Validate() {
	wg := sync.WaitGroup{}
	for {
		select {
		case r := <-v.parserChan:
			if r == nil {
				wg.Wait()
				v.ch <- nil
				return
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				v.ch <- v.ValidateOne(r.Proxy)
			}()
		}
	}
}

func (v *Validator) ValidateOne(p proxy.Proxy) *Result {
	r := &Result{Proxy: p}
	m, err := p.GetClashMapping()
	if err != nil {
		r.Error = err
		return r
	}
	clashProxy, err := adapter.ParseProxy(m)
	if err != nil {
		r.Error = err
		return r
	}

	ctx, cancel := context.WithTimeout(context.Background(), v.cfg.Timeout)
	defer cancel()

	r.Delay, r.Error = clashProxy.URLTest(ctx, v.cfg.TestURL)
	if r.Error != nil {
		return r
	}

	r.CountryCode, r.Country, r.Error = getCountryInfo(p.GetBase().Server)
	if r.Error != nil {
		return r
	}
	r.CountryEmoji = emoji.GetFlag(r.CountryCode)
	return r
}

func (v *Validator) Chan() <-chan *Result {
	return v.ch
}

func getCountryInfo(server string) (string, string, error) {
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=countryCode,country", server))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	var res struct {
		Country     string `json:"country"`
		CountryCode string `json:"countryCode"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return "", "", err
	}

	return res.CountryCode, res.Country, nil
}
