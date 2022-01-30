package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Dreamacro/clash/adapter"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/proxy"
)

type Result struct {
	Proxy       proxy.Proxy
	Delay       uint16
	Country     string
	CountryCode string
}

type Validator struct {
	cfg *config.ValidatorConfig
}

func New(cfg *config.ValidatorConfig) *Validator {
	return &Validator{
		cfg: cfg,
	}
}

func (v *Validator) Validate(p proxy.Proxy) (*Result, error) {
	r := &Result{Proxy: p}
	m, err := p.ConfigMap()
	if err != nil {
		return nil, err
	}
	clashProxy, err := adapter.ParseProxy(m)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), v.cfg.Timeout)
	defer cancel()

	r.Delay, err = clashProxy.URLTest(ctx, v.cfg.TestURL)
	if err != nil {
		return nil, err
	}

	r.CountryCode, r.Country, err = getCountryInfo(p.GetBase().Server)
	if err != nil {
		return nil, err
	}
	return r, nil
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
